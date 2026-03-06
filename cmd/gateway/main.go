package main

//main.go
import (
	"log"
	"net/http"

	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/Chimera-State/GigaScale/internal/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	//localhost
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to the backend %v", err)
	}
	defer conn.Close()

	client := pb.NewReservationServiceClient(conn)

	srv := gateway.NewServer(client, gateway.NewLocalLimiter())
	//ratelimiter
	secureHandler := srv.RateLimiter(http.HandlerFunc(srv.HandleReserve))
	//router(Mux)
	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/reserve", secureHandler)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start %v", err)
	}
}
