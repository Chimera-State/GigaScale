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

	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
}

// constructor
func NewServer(client pb.ReservationServiceClient) *Server {
	return &Server{
		reserveClient: client,
		validator:     validator.New(),

		capacity:   10,
		tokens:     10,
		refillRate: 1,
		lastRefill: time.Now(),
	}
}
