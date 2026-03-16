package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	reservationv1 "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/Chimera-State/GigaScale/internal/backend/pkg/db"
	"github.com/Chimera-State/GigaScale/internal/backend/pkg/redislock"
	"github.com/Chimera-State/GigaScale/internal/backend/redisclient"
	"github.com/Chimera-State/GigaScale/internal/backend/repository"
	"github.com/Chimera-State/GigaScale/internal/backend/service"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {

	redisclient.NewRedisClient()
	redisclient.HealthCheck(context.Background())

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("soket dinlenemedi: %v", err)
	}
	loggingInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Printf("Gelen İstek: %s", info.FullMethod)
		return handler(ctx, req)
	}
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(loggingInterceptor),
	)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	locker := redislock.NewLocker(rdb)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL ortam değişkeni ayarlanmadı")
	}
	dbPool, err := db.NewDatabase(databaseURL)
	if err != nil {
		log.Fatalf("Veritabanı başlatılamadı: %v", err)
	}
	defer dbPool.Close()
	repo, err := repository.NewPostgresReservationRepository(dbPool)
	if err != nil {
		log.Fatalf("Repository başlatılamadı: %v", err)
	}
	myService := service.NewReservationService(locker, repo)
	reservationv1.RegisterReservationServiceServer(s, myService)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Printf("gRPC sunucusu :50051 portunda başladı...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Sunucu hatası: %v", err)
		}
	}()
	<-stop
	log.Println("\nKapatma sinyali alındı. Sunucu güvenli bir şekilde kapatılıyor...")
	s.GracefulStop()
	log.Println("Sunucu tamamen durduruldu.")
}
