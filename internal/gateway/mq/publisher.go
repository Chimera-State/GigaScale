package mq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
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
	writer *kafka.Writer
}

func New(brokerAddr string) *Publisher {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokerAddr),
		Topic:        "reservations.created",
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
	return &Publisher{writer: w}
}
func (p *Publisher) Publish(ctx context.Context, event ReservationCreatedEvent) error {
	event.Timestamp = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.IdempotencyKey),
		Value: data,
	})
	if err != nil {
		log.Printf("[KAFKA] Publish hatası: %v", err)
		return err
	}

	log.Printf("[KAFKA] Event yayınlandı: user=%s seat=%s payment=%s",
		event.UserID, event.SeatID, event.PaymentID)
	return nil
}

func (p *Publisher) Close() {
	if err := p.writer.Close(); err != nil {
		log.Printf("[KAFKA] Close hatası: %v", err)
	}
}
