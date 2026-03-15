package main

//main.go
import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/Chimera-State/GigaScale/internal/gateway"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func healthHandler(rdb *redis.Client, conn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := map[string]string{
			"status":  "ok",
			"redis":   "ok",
			"backend": "ok",
		}

		if rdb != nil {
			if err := rdb.Ping(r.Context()).Err(); err != nil {
				result["redis"] = "error"
				result["status"] = "degraded"
			}
		}

		if conn != nil {
			if conn.GetState() != connectivity.Ready {
				result["backend"] = "error"
				result["status"] = "degraded"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func main() {
	//addr
	backendAddr := getEnv("BACKEND_ADDR", "localhost:50051")
	serverPort := getEnv("SERVER_PORT", ":8080")

	reddisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{
		Addr: reddisAddr,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis bağlantı hatası: %v", err)
	}
	defer rdb.Close()
	//gateway conn
	conn, err := grpc.NewClient(backendAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to the backend %v", err)
	}
	defer conn.Close()

	client := pb.NewReservationServiceClient(conn)

	useRedisLimiter := getEnv("USE_REDIS_LIMITER", "false") == "true"
	var limiter gateway.RateLimiter
	if useRedisLimiter {
		limiter = gateway.NewRedisLimiter(rdb, 10, 2)
	} else {
		limiter = gateway.NewIpLimiter(10, 2)
	}

	//dp injection
	v := validator.New()
	srv := gateway.NewServer(client, limiter, v)

	mux := http.NewServeMux()
	secureHandler := srv.RateLimiter(http.HandlerFunc(srv.HandleReserve))
	mux.Handle("POST /api/v1/reserve", secureHandler)

	mux.HandleFunc("GET /health", healthHandler(rdb, conn))

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
