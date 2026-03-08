package gateway

//handler.go
import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
)

func (s *Server) HandleReserve(w http.ResponseWriter, r *http.Request) {

	var req ReserveHTTPRequest

	//JSON decode
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON Decoding Error %v", err)
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}

	if err := s.validator.Struct(req); err != nil {
		log.Printf("Validation Error: %v", err)
		http.Error(w, "Data Validation Error: "+err.Error(), http.StatusBadRequest)
		return
	}

	//mapping
	grpcReq := &pb.ReserveSeatRequest{
		UserId:         req.UserID,
		TripId:         req.TripID,
		SeatId:         req.SeatID,
		IdempotencyKey: req.IdempotencyKey,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

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
	json.NewEncoder(w).Encode(httpResp)

}
