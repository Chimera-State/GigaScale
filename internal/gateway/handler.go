package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
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

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
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

	httpResp := ReserveHTTPResponse{
		Success: resp.Success,
		Message: resp.Message,
	}
	w.Header().Set("Content-Type", "application/json")
	
	if !resp.Success {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(httpResp)
}
