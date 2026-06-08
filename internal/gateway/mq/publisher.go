package mq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	otelkafka "github.com/Chimera-State/go-otel-kit/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
)

type ReservationCreatedEvent struct {
	UserID         string    `json:"user_id"`
	TripID         string    `json:"trip_id"`
	SeatID         string    `json:"seat_id"`
	PaymentID      string    `json:"payment_id"`
	IdempotencyKey string    `json:"idempotency_key"`
	Timestamp      time.Time `json:"timestamp"`
}

type Publisher struct {
	client *kgo.Client
	topic  string
}

func New(brokerAddr string) (*Publisher, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokerAddr),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, err
	}
	return &Publisher{client: cl, topic: "reservations.created"}, nil
}

func (p *Publisher) Publish(ctx context.Context, event ReservationCreatedEvent) error {
	event.Timestamp = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(event.IdempotencyKey),
		Value: data,
	}

	// Inject OTel trace context into Kafka headers so the notifier
	// can continue the same trace chain via ExtractFromRecord.
	otelkafka.InjectToRecord(ctx, record)

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		log.Printf("[KAFKA] publish error: %v", err)
		return err
	}

	log.Printf("[KAFKA] Event published: user=%s seat=%s payment=%s",
		event.UserID, event.SeatID, event.PaymentID)
	return nil
}

func (p *Publisher) Close() {
	p.client.Close()
}
