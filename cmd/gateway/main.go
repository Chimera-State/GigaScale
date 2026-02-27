package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Chimera-State/GigaScale/internal/gateway"
	pb "github.com/Chimera-State/GigaScale/internal/pb/reservationv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	backendAddr := os.Getenv("BACKEND_ADDR")
	if backendAddr == "" {
		backendAddr = "localhost:50051"
	}

	// Backend gRPC bağlantısı
	conn, err := grpc.NewClient(backendAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Backend'e bağlanılamadı (%s): %v", backendAddr, err)
	}
	defer conn.Close()

	log.Printf("Gateway başlatıldı — backend: %s", backendAddr)

	
	gateway.ReserveClient = pb.NewReservationServiceClient(conn)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/reserve", gateway.HandleReserve)

	log.Println("HTTP :8080 portunda çalışıyor...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("HTTP server hatası: %v", err)
	}
}

