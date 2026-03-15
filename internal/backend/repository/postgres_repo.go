package repository
import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)
type PostgresReservationRepository struct {
	db                   *sql.DB
	stmtCreate           *sql.Stmt
	stmtGetByIdempotency *sql.Stmt
	stmtExists           *sql.Stmt
	stmtCancel           *sql.Stmt
}
func NewPostgresReservationRepository(db *sql.DB) (*PostgresReservationRepository, error) {
	stmtCreate, err := db.Prepare(`
		INSERT INTO reservations(id, user_id, trip_id, seat_id, idempotency_key, payment_id, amount, status, created_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare stmtCreate: %w", err)
	}
	stmtGetByIdempotency, err := db.Prepare(`
		SELECT id, user_id, trip_id, seat_id, idempotency_key, payment_id, amount, status, created_at, cancelled_at
		FROM reservations
		WHERE user_id = $1 AND idempotency_key = $2
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare stmtGetByIdempotency: %w", err)
	}
	stmtExists, err := db.Prepare(`
		SELECT EXISTS(
			SELECT 1 FROM reservations
			WHERE trip_id = $1 AND seat_id = $2 AND status = 'confirmed'
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare stmtExists: %w", err)
	}
	stmtCancel, err := db.Prepare(`
		UPDATE reservations
		SET status = 'cancelled', cancelled_at = NOW()
		WHERE idempotency_key = $1 AND status = 'confirmed'
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare stmtCancel: %w", err)
	}
	return &PostgresReservationRepository{
		db:                   db,
		stmtCreate:           stmtCreate,
		stmtGetByIdempotency: stmtGetByIdempotency,
		stmtExists:           stmtExists,
		stmtCancel:           stmtCancel,
	}, nil
}
func (r *PostgresReservationRepository) Create(ctx context.Context, req *Reservation) error {
	_, err := r.stmtCreate.ExecContext(ctx,
		req.ID,
		req.UserID,
		req.TripID,
		req.SeatID,
		req.IdempotencyKey,
		req.PaymentID,
		req.Amount,
		req.Status,
		req.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create reservation sql hatası: %w", err)
	}
	return nil
}
func (r *PostgresReservationRepository) GetByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (*Reservation, error) {
	var res Reservation
	err := r.stmtGetByIdempotency.QueryRowContext(ctx, userID, key).Scan(
		&res.ID,
		&res.UserID,
		&res.TripID,
		&res.SeatID,
		&res.IdempotencyKey,
		&res.PaymentID,
		&res.Amount,
		&res.Status,
		&res.CreatedAt,
		&res.CancelledAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get by idempotency key sql hatası: %w", err)
	}
	return &res, nil
}
func (r *PostgresReservationRepository) Exists(ctx context.Context, tripID uuid.UUID, seatID string) (bool, error) {
	var exists bool
	err := r.stmtExists.QueryRowContext(ctx, tripID, seatID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("exists (kontrol) sql hatası: %w", err)
	}
	return exists, nil
}
func (r *PostgresReservationRepository) Cancel(ctx context.Context, idempotencyKey string) error {
	_, err := r.stmtCancel.ExecContext(ctx, idempotencyKey)
	if err != nil {
		return fmt.Errorf("cancel reservation sql hatası: %w", err)
	}
	return nil
}
