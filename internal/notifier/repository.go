package notifier

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// NotificationRecord represents a single notification attempt logged to the DB.
type NotificationRecord struct {
	IdempotencyKey string
	UserID         string
	TripID         string
	SeatID         string
	PaymentID      string
	Channel        string // "webhook"
	Status         string // "sent" | "failed"
	SentAt         time.Time
}

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// AlreadySent returns true if a notification for the given idempotency key
func (r *NotificationRepository) AlreadySent(ctx context.Context, idempotencyKey, channel string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM notifications
			WHERE idempotency_key = $1
			  AND channel         = $2
			  AND status          = 'sent'
		)`, idempotencyKey, channel,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("notifications.AlreadySent: %w", err)
	}
	return exists, nil
}

// Log inserts a notification record (sent or failed) into the audit table.
func (r *NotificationRepository) Log(ctx context.Context, n NotificationRecord) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notifications
			(idempotency_key, user_id, trip_id, seat_id, payment_id, channel, status, sent_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (idempotency_key, channel) DO UPDATE
			SET status  = EXCLUDED.status,
			    sent_at = EXCLUDED.sent_at`,
		n.IdempotencyKey, n.UserID, n.TripID, n.SeatID, n.PaymentID,
		n.Channel, n.Status, n.SentAt,
	)
	if err != nil {
		return fmt.Errorf("notifications.Log: %w", err)
	}
	return nil
}
