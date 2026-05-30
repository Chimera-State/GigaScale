package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookPayload is the JSON body sent to the configured webhook URL.
type WebhookPayload struct {
	Event          string `json:"event"` // always "reservation.created"
	UserID         string `json:"user_id"`
	TripID         string `json:"trip_id"`
	SeatID         string `json:"seat_id"`
	PaymentID      string `json:"payment_id"`
	IdempotencyKey string `json:"idempotency_key"`
	SentAt         string `json:"sent_at"`
}

type WebhookClient struct {
	targetURL  string
	httpClient *http.Client
}

func NewWebhookClient(targetURL string) *WebhookClient {
	return &WebhookClient{
		targetURL: targetURL,
		// Short timeout the notifier must not block the Kafka consumer loop.
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Send fires a POST to the target URL with the reservation payload.
// ctx carries the OTel span so the HTTP call appears as a child span.
func (w *WebhookClient) Send(ctx context.Context, n NotificationRecord) error {
	payload := WebhookPayload{
		Event:          "reservation.created",
		UserID:         n.UserID,
		TripID:         n.TripID,
		SeatID:         n.SeatID,
		PaymentID:      n.PaymentID,
		IdempotencyKey: n.IdempotencyKey,
		SentAt:         time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.targetURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: build request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: non-2xx response: %d", resp.StatusCode)
	}

	return nil
}
