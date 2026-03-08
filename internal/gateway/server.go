package gateway

import (
	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/go-playground/validator/v10"
)

// server method skill
type Server struct {
	reserveClient pb.ReservationServiceClient
	validator     *validator.Validate
	limiter       RateLimiter
}

// constructor
func NewServer(client pb.ReservationServiceClient, limiter RateLimiter, v *validator.Validate) *Server {
	return &Server{
		reserveClient: client,
		validator:     v,
		limiter:       limiter,
	}
}
