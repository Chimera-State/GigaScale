package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Chimera-State/go-otel-kit/setup"
	"github.com/segmentio/kafka-go"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

type ReservationCreatedEvent struct {
	UserID         string `json:"user_id"`
	TripID         string `json:"trip_id"`
	SeatID         string `json:"seat_id"`
	PaymentID      string `json:"payment_id"`
	IdempotencyKey string `json:"idempotency_key"`
	Timestamp      string `json:"timestamp"`
}

func main() {
	kafkaAddr := getEnv("KAFKA_ADDR", "kafka:9092")

	ctx := context.Background()
	if err := setup.Init(ctx,
		setup.WithServiceName("gigascale-notifier"),
		setup.WithServiceVersion("1.0.0"),
		setup.WithExporterEndpoint("otel-collector:4317"),
	); err != nil {
		log.Fatalf("OTel initialization failed: %v", err)
	}
	defer setup.Shutdown(ctx)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaAddr},
		Topic:    "reservations.created",
		GroupID:  "notifier-group",
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Printf("[NOTIFIER] Kafka is listening on: %s , topic: reservations.created", kafkaAddr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("[NOTIFIER] Reading error: %v", err)
				return
			}

			var event ReservationCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[NOTIFIER] JSON parse error: %v", err)
				continue
			}

			log.Printf("[NOTIFIER] New reservation: UserID=%s TripID=%s SeatID=%s PaymentID=%s",
				event.UserID, event.TripID, event.SeatID, event.PaymentID)
		}
	}()

	<-stop
	log.Println("[NOTIFIER] Shutting down...")
}
