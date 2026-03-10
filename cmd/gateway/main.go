package main

//main.go
import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/Chimera-State/GigaScale/internal/gateway"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
func main() {
	//localhost
	backendAddr := getEnv("BACKEND_ADDR", "localhost:50051")
	serverPort := getEnv("SERVER_PORT", ":8080")

	conn, err := grpc.NewClient(backendAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to the backend %v", err)
	}
	defer conn.Close()

	client := pb.NewReservationServiceClient(conn)

	//dp injection
	v := validator.New()
	l := gateway.NewIpLimiter(10, 2)

	srv := gateway.NewServer(client, l, v)

	mux := http.NewServeMux()
	secureHandler := srv.RateLimiter(http.HandlerFunc(srv.HandleReserve))
	mux.Handle("POST /api/v1/reserve", secureHandler)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GigaScale is alive"))
	})

	httpServer := &http.Server{
		Addr:    serverPort,
		Handler: mux,
	}

	go func() {
		log.Printf("GigaScale gateway starting on %s...", serverPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Printf("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}
	log.Println("GigaScale exited clean.")

}
