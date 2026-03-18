package gateway

//handler.go
import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (s *Server) HandleReserve(w http.ResponseWriter, r *http.Request) {

	traceID := uuid.New().String()
	log.Printf("[TRACE: %s] [ENTER] New reservation request reached the Gateway", traceID)

	var req ReserveHTTPRequest

	//JSON decode
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("[TRACE: %s] [ERROR] JSON Decoding Error %v", traceID, err)
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}

	if err := s.validator.Struct(req); err != nil {
		log.Printf("[TRACE: %s] [ERROR] Validation Error: %v", traceID, err)
		http.Error(w, "Data Validation Error: "+err.Error(), http.StatusBadRequest)
		return
	}

	lockKey := "lock:reservation:trip:" + req.TripID + ":seat:" + req.SeatID
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = s.rdb.SetArgs(ctx, lockKey, req.UserID, redis.SetArgs{
		Mode: "NX",
		TTL:  15 * time.Second,
	}).Err()

	if err != nil {
		if err == redis.Nil {
			log.Printf("[TRACE: %s] [BLOCKED] Seat is already locked.", traceID)
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "Seat is busy"})
			return
		}
		log.Printf("[TRACE: %s] [ERROR] Redis Connection Error: %s", traceID, err)
		http.Error(w, "Service Unavailable", 500)
	}
	defer s.rdb.Del(context.Background(), lockKey)

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
