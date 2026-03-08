package gateway

import (
	"log"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) handleGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		log.Printf("CRITICAL: Unknown error type: %v", err)
		http.Error(w, "An unexpected error occured", http.StatusInternalServerError)
	}

	log.Printf("[BACKEND ERROR] Code: %s, Message: %s", st.Code(), st.Message())

	switch st.Code() {
	case codes.InvalidArgument:
		http.Error(w, st.Message(), http.StatusBadRequest) //400
	case codes.NotFound:
		http.Error(w, "The requested resource could not be found.", http.StatusNotFound)
	case codes.AlreadyExists:
		http.Error(w, "The seat is already reserved.", http.StatusConflict)
	case codes.Unauthenticated:
		http.Error(w, "Unauthorized access.", http.StatusUnauthorized)
	case codes.DeadlineExceeded:
		http.Error(w, "Backend did not respond (Timeout)", http.StatusGatewayTimeout)
	default:
		http.Error(w, "The operation cannot be performed.", http.StatusInternalServerError)
	}
}
