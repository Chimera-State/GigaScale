package gateway

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	pb "github.com/Chimera-State/GigaScale/internal/pb/reservationv1"
)

var ReserveClient pb.ReservationServiceClient

func HandleReserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST istekleri kabul edilir.", http.StatusMethodNotAllowed)
		return
	}

	var req ReserveHTTPRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON Decoding Error %v", err)
		http.Error(w, "Invalid data format", http.StatusBadRequest)
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

	resp, err := ReserveClient.ReserveSeat(ctx, grpcReq)
	if err != nil {
		log.Printf("Backend gRPC Error %v", err)
		http.Error(w, "Backend service was unavailable.", http.StatusInternalServerError)
		return
	}

	httpResp := ReserveHTTPResponse{
		Success: resp.Success,
		Message: resp.Message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(httpResp)

}
