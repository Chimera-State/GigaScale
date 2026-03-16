package repository
import (
	"context"
	"time"
	"github.com/google/uuid"
)
type Reservation struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	TripID         uuid.UUID
	SeatID         string
	IdempotencyKey string
	PaymentID      uuid.NullUUID
	Amount         float64
	Status         string
	CreatedAt      time.Time
	CancelledAt    *time.Time
}
type ReservationRepository interface {
	Create(ctx context.Context, req *Reservation) error
	GetByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (*Reservation, error)
	Exists(ctx context.Context, tripID uuid.UUID, seatID string) (bool, error)
	Cancel(ctx context.Context, idempotencyKey string) error
}
