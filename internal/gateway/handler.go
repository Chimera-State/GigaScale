package gateway
<<<<<<< HEAD

//handler.go
=======
>>>>>>> origin/fetaure/sprint3
import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
)
<<<<<<< HEAD

func (s *Server) HandleReserve(w http.ResponseWriter, r *http.Request) {

	var req ReserveHTTPRequest

	//JSON decode
=======
var ReserveClient pb.ReservationServiceClient
func HandleReserve(w http.ResponseWriter, r *http.Request) {
	var req ReserveHTTPRequest
	var validate = validator.New()
>>>>>>> origin/fetaure/sprint3
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON Decoding Error %v", err)
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}
<<<<<<< HEAD

	if err := s.validator.Struct(req); err != nil {
=======
	if err := validate.Struct(req); err != nil {
>>>>>>> origin/fetaure/sprint3
		log.Printf("Validation Error: %v", err)
		http.Error(w, "Data Validation Error: "+err.Error(), http.StatusBadRequest)
		return
	}
	grpcReq := &pb.ReserveSeatRequest{
		UserId:         req.UserID,
		TripId:         req.TripID,
		SeatId:         req.SeatID,
		IdempotencyKey: req.IdempotencyKey,
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
<<<<<<< HEAD

	resp, err := s.reserveClient.ReserveSeat(ctx, grpcReq)
=======
	resp, err := ReserveClient.ReserveSeat(ctx, grpcReq)
>>>>>>> origin/fetaure/sprint3
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
