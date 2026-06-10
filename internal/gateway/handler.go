package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Chimera-State/GigaScale/internal/gateway/orchestrator"
	"github.com/redis/go-redis/v9"
)

func (s *Server) HandleReserve(w http.ResponseWriter, r *http.Request) {

	//  traceID := uuid.New().String()
	//  log.Printf("[TRACE: %s] [ENTER] New reservation request reached the Gateway", traceID)

	var req ReserveHTTPRequest

	//JSON decode
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		//  log.Printf("[TRACE: %s] [ERROR] JSON Decoding Error %v", traceID, err)
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}

	if err := s.validator.Struct(req); err != nil {
		//  log.Printf("[TRACE: %s] [ERROR] Validation Error: %v", traceID, err)
		http.Error(w, "Data Validation Error: "+err.Error(), http.StatusBadRequest)
		return
	}

	lockKey := "lock:seat:" + req.SeatID

	acquired, err := s.rdb.SetNX(r.Context(), lockKey, req.UserID, 5*time.Second).Result()

	if err != nil {
		http.Error(w, "Redis Lock Error", http.StatusInternalServerError)
		return
	}
	if !acquired {
		w.WriteHeader(http.StatusConflict) // 409
		json.NewEncoder(w).Encode(map[string]string{"error": "Seat is being processed or already taken"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result, err := s.orchestrator.ReserveWithPayment(ctx, orchestrator.ReserveRequest{
		UserID:         req.UserID,
		TripID:         req.TripID,
		SeatID:         req.SeatID,
		IdempotencyKey: req.IdempotencyKey,
		Amount:         req.Amount,
	})

	if err != nil {
		s.handleGRPCError(w, err)
		return
	}

	if !result.Success {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ReserveHTTPResponse{
		Success:   result.Success,
		Message:   result.Message,
		PaymentID: result.PaymentID,
	})
}
