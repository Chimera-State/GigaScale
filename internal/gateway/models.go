package gateway

//istekler için struct
type ReserveHTTPRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	TripID         string `json:"trip_id" validate:"required"`
	SeatID         string `json:"seat_id"  validate:"required,alphanum"`
	IdempotencyKey string `json:"Idempotency_key" validate:"required"`
}

//yanıtlar için struct
type ReserveHTTPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
