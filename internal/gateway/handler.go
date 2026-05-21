package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
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

	err = s.rdb.SetArgs(r.Context(), lockKey, req.UserID, redis.SetArgs{
		Mode: "NX",
		TTL:  5 * time.Second,
	}).Err()

	if err == redis.Nil {
		w.WriteHeader(http.StatusConflict) // 409
		json.NewEncoder(w).Encode(map[string]string{"error": "Seat is being processed or already taken"})
		return
	} else if err != nil {
		http.Error(w, "Redis Lock Error", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	grpcReq := &pb.ReserveSeatRequest{
		UserId:         req.UserID,
		TripId:         req.TripID,
		SeatId:         req.SeatID,
		IdempotencyKey: req.IdempotencyKey,
	}

	resp, err := s.reserveClient.ReserveSeat(ctx, grpcReq)
	if err != nil {
		s.handleGRPCError(w, err)
		return
	}

	if !resp.Success {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	httpResp := ReserveHTTPResponse{
		Success: resp.Success,
		Message: resp.Message,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(httpResp)
}
