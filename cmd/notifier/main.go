package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Chimera-State/GigaScale/internal/notifier"
	otelkafka "github.com/Chimera-State/go-otel-kit/kafka"
	"github.com/Chimera-State/go-otel-kit/setup"
	_ "github.com/lib/pq"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// validAppEnvs lists the only accepted values for APP_ENV.
// The service refuses to start with any other value (including empty string).
var validAppEnvs = map[string]bool{
	"development": true,
	"test":        true,
	"production":  true,
}

// ReservationCreatedEvent is the Kafka event payload published by the gateway.
type ReservationCreatedEvent struct {
	UserID         string `json:"user_id"`
	TripID         string `json:"trip_id"`
	SeatID         string `json:"seat_id"`
	PaymentID      string `json:"payment_id"`
	IdempotencyKey string `json:"idempotency_key"`
	Timestamp      string `json:"timestamp"`
}

// sendNotification delivers a webhook notification and logs the result to DB.
// repo and webhookClient may be nil in non-production environments.
func sendNotification(
	ctx context.Context,
	repo *notifier.NotificationRepository,
	webhookClient *notifier.WebhookClient,
	event ReservationCreatedEvent,
) error {
	const channel = "webhook"

	// Idempotency check — skip if already successfully delivered.
	if repo != nil {
		alreadySent, err := repo.AlreadySent(ctx, event.IdempotencyKey, channel)
		if err != nil {
			return err
		}
		if alreadySent {
			log.Printf("[NOTIFIER] Duplicate skipped: idempotency_key=%s", event.IdempotencyKey)
			return nil
		}
	}

	record := notifier.NotificationRecord{
		IdempotencyKey: event.IdempotencyKey,
		UserID:         event.UserID,
		TripID:         event.TripID,
		SeatID:         event.SeatID,
		PaymentID:      event.PaymentID,
		Channel:        channel,
		SentAt:         time.Now().UTC(),
	}

	// Fire the webhook (ctx carries the OTel trace into the HTTP call).
	if webhookClient != nil {
		if err := webhookClient.Send(ctx, record); err != nil {
			record.Status = "failed"
			if repo != nil {
				_ = repo.Log(ctx, record) // best-effort log
			}
			return err
		}
	}

	record.Status = "sent"
	log.Printf("[NOTIFIER] Webhook delivered: UserID=%s SeatID=%s PaymentID=%s",
		event.UserID, event.SeatID, event.PaymentID)

	// Persist the audit record — non-fatal if it fails.
	if repo != nil {
		if err := repo.Log(ctx, record); err != nil {
			log.Printf("[NOTIFIER] Failed to log notification to DB: %v", err)
		}
	}

	return nil
}

func main() {
	// fail-fast: APP_ENV must be explicitly set to a known value.
	appEnv := os.Getenv("APP_ENV")
	if !validAppEnvs[appEnv] {
		log.Fatalf("[NOTIFIER] APP_ENV is not set or invalid (got %q). Must be one of: development, test, production", appEnv)
	}
	log.Printf("[NOTIFIER] Starting in %s mode", appEnv)

	kafkaAddr := getEnv("KAFKA_ADDR", "kafka:9092")
	webhookURL := getEnv("WEBHOOK_URL", "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := setup.Init(ctx,
		setup.WithServiceName("gigascale-notifier"),
		setup.WithServiceVersion("1.0.0"),
		setup.WithExporterEndpoint("otel-collector:4317"),
	); err != nil {
		log.Fatalf("OTel initialization failed: %v", err)
	}
	defer setup.Shutdown(ctx)

	// DB connection — required only in production for idempotency + audit log.
	var repo *notifier.NotificationRepository
	if appEnv == "production" {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			log.Fatal("[NOTIFIER] DATABASE_URL is required in production")
		}
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("[NOTIFIER] DB connection failed: %v", err)
		}
		defer db.Close()
		repo = notifier.NewNotificationRepository(db)
	}

	// Webhook client — required only when WEBHOOK_URL is set.
	var webhookClient *notifier.WebhookClient
	if webhookURL != "" {
		webhookClient = notifier.NewWebhookClient(webhookURL)
		log.Printf("[NOTIFIER] Webhook target: %s", webhookURL)
	} else {
		log.Printf("[NOTIFIER] WEBHOOK_URL not set — running in log-only mode")
	}

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(kafkaAddr),
		kgo.ConsumerGroup("notifier-group"),
		kgo.ConsumeTopics("reservations.created"),
	)
	if err != nil {
		log.Fatalf("[NOTIFIER] Kafka client error: %v", err)
	}
	defer cl.Close()

	log.Printf("[NOTIFIER] Listening on %s, topic: reservations.created", kafkaAddr)

	tracer := otel.Tracer("gigascale-notifier")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			fetches := cl.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				if ctx.Err() != nil {
					return // graceful shutdown
				}
				log.Printf("[NOTIFIER] Fetch error: %v", errs)
				continue
			}

			fetches.EachRecord(func(record *kgo.Record) {
				// Extract TraceID from Kafka headers — continues the gateway's trace
				msgCtx := otelkafka.ExtractFromRecord(ctx, record)

				// Child span of the gateway's publish span
				msgCtx, span := tracer.Start(msgCtx, "notifier.process_reservation")
				defer span.End()

				var event ReservationCreatedEvent
				if err := json.Unmarshal(record.Value, &event); err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, "json unmarshal failed")
					log.Printf("[NOTIFIER] JSON parse error: %v", err)
					return
				}

				span.SetAttributes(
					attribute.String("reservation.user_id", event.UserID),
					attribute.String("reservation.trip_id", event.TripID),
					attribute.String("reservation.seat_id", event.SeatID),
					attribute.String("reservation.payment_id", event.PaymentID),
				)

				// msgCtx → trace propagates into the webhook HTTP call
				if err := sendNotification(msgCtx, repo, webhookClient, event); err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, "notification failed")
					log.Printf("[NOTIFIER] Notification failed: %v", err)
					return
				}

				span.SetStatus(codes.Ok, "")
			})
		}
	}()

	<-stop
	cancel() // signal the consumer goroutine to stop polling
	log.Println("[NOTIFIER] Shutting down...")
}
