package gateway

//models.go

import (
	"sync"
	"time"

	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/go-playground/validator/v10"
)

// istekler için struct
type ReserveHTTPRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	TripID         string `json:"trip_id" validate:"required"`
	SeatID         string `json:"seat_id"  validate:"required,alphanum"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
}

// yanıtlar için struct
type ReserveHTTPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// method skill
type Server struct {
	reserveClient pb.ReservationServiceClient
	validator     *validator.Validate
	limiter       RateLimiter
}
type RateLimiter interface {
	Allow() bool
}

type LocalLimiter struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
}

func NewLocalLimiter() *LocalLimiter {
	return &LocalLimiter{
		capacity:   10.0,
		tokens:     4.0,
		refillRate: 0.5,
		lastRefill: time.Now(),
	}
}

func (l *LocalLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	duration := now.Sub(l.lastRefill).Seconds()
	l.tokens += duration * l.refillRate
	if l.tokens > l.capacity {
		l.tokens = l.capacity
	}
	l.lastRefill = now
	if l.tokens > 1.0 {
		l.tokens -= 1.0
		return true
	}
	return false
}

// constructor
func NewServer(client pb.ReservationServiceClient, limiter RateLimiter) *Server {
	return &Server{
		reserveClient: client,
		validator:     validator.New(),
		limiter:       limiter,
	}
}
