package gateway
<<<<<<< HEAD:internal/gateway/types.go

//models.go

// istekler için struct
=======
>>>>>>> origin/fetaure/sprint3:internal/gateway/models.go
type ReserveHTTPRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	TripID         string `json:"trip_id" validate:"required"`
	SeatID         string `json:"seat_id"  validate:"required,alphanum"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
}
<<<<<<< HEAD:internal/gateway/types.go

// yanıtlar için struct
=======
>>>>>>> origin/fetaure/sprint3:internal/gateway/models.go
type ReserveHTTPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
