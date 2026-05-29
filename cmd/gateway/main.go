package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pbpay "github.com/Chimera-State/GigaScale/api/proto/payment/v1"
	pb "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/Chimera-State/GigaScale/internal/gateway"
	"github.com/Chimera-State/GigaScale/internal/gateway/mq"
	"github.com/Chimera-State/GigaScale/internal/gateway/orchestrator"
	"github.com/Chimera-State/go-otel-kit/interceptor"
	"github.com/Chimera-State/go-otel-kit/middleware"
	"github.com/Chimera-State/go-otel-kit/setup"
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

func healthHandler(rdb redis.UniversalClient, conn *grpc.ClientConn) http.HandlerFunc {
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
			state := conn.GetState()
			if state != connectivity.Ready && state != connectivity.Idle {
				result["backend"] = "error"
				result["status"] = "degraded"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func main() {
	// addr
	backendAddr := getEnv("BACKEND_ADDR", "localhost:50051")
	paymentAddr := getEnv("PAYMENT_ADDR", "localhost:50052")
	kafkaAddr := getEnv("KAFKA_ADDR", "kafka:9092")
	serverPort := getEnv("SERVER_PORT", ":8080")

	ctx := context.Background()
	if err := setup.Init(ctx,
		setup.WithServiceName("gigascale-gateway"),
		setup.WithServiceVersion("1.0.0"),
		setup.WithExporterEndpoint("otel-collector:4317"),
	); err != nil {
		log.Fatalf("Otel initialization failed: %v", err)
	}
	defer setup.Shutdown(ctx)

	// redis cluster
	clusterAddrs := []string{
		"redis-node-1:6379", "redis-node-2:6379", "redis-node-3:6379",
		"redis-node-4:6379", "redis-node-5:6379", "redis-node-6:6379",
	}
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        clusterAddrs,
		MaxRedirects: 8, // tolerance
		ReadOnly:     false,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis bağlantı hatası: %v", err)
	}
	defer rdb.Close()

	//gateway conn
	backendOpts := append(interceptor.ClientOptions(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`))

	conn, err := grpc.NewClient(backendAddr, backendOpts...)

	if err != nil {
		log.Fatalf("Could not connect to the backend: %v", err)
	}
	defer conn.Close()
	reserveClient := pb.NewReservationServiceClient(conn)

	// payment gRPC bağlantısı
	paymentOpts := append(interceptor.ClientOptions(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	paymentConn, err := grpc.NewClient(paymentAddr, paymentOpts...)
	if err != nil {
		log.Fatalf("Could not connect to payment service: %v", err)
	}
	defer paymentConn.Close()
	paymentClient := pbpay.NewPaymentServiceClient(paymentConn)

	// rate limiter
	useRedisLimiter := strings.TrimSpace(getEnv("USE_REDIS_LIMITER", "false")) == "true"
	var limiter gateway.RateLimiter
	if useRedisLimiter {
		limiter = gateway.NewRedisLimiter(rdb, 10, 2)
	} else {
		limiter = gateway.NewIpLimiter(10, 2)
	}

	// kafka publisher
	publisher := mq.New(kafkaAddr)
	defer publisher.Close()

	// orchestrator
	orch := orchestrator.New(reserveClient, paymentClient, publisher)

	// DI
	v := validator.New()
	srv := gateway.NewServer(reserveClient, limiter, v, rdb, orch)

	// router
	mux := http.NewServeMux()
	secureHandler := middleware.TraceMiddleware(srv.RateLimiter(http.HandlerFunc(srv.HandleReserve)))
	mux.Handle("POST /api/v1/reserve", secureHandler)
	mux.HandleFunc("GET /health", healthHandler(rdb, conn))

	httpServer := &http.Server{
		Addr:         serverPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
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
