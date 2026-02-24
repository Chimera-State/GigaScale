package gateway

//istekler için struct
type ReserveHTTPRequest struct {
	UserID         string `json:"user_id"`
	TripID         string `json:"trip_id"`
	SeatID         string `json:"seat_id"`
	IdempotencyKey string `json:"Idempotency_key"`
}

//yanıtlar için struct
type ReserveHTTPResponse struct {
	Succes  bool   `json:"success"`
	Message string `json:"message"`
}
