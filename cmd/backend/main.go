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
	"github.com/Chimera-State/go-otel-kit/interceptor"
	"github.com/Chimera-State/go-otel-kit/setup"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	if err := setup.Init(ctx,
		setup.WithServiceName("gigascale-backend"),
		setup.WithServiceVersion("1.0.0"),
		setup.WithExporterEndpoint("otel-collector:4317"),
	); err != nil {
		log.Fatalf("OTel initaliziton failed: %v", err)
	}
	defer setup.Shutdown(ctx)

	redisclient.InitRedisCluster()
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
		append(interceptor.ServerOptions(),
			grpc.UnaryInterceptor(loggingInterceptor),
		)...,
	)
	clusterAddrs := []string{
		"redis-node-1:6379", "redis-node-2:6379", "redis-node-3:6379",
		"redis-node-4:6379", "redis-node-5:6379", "redis-node-6:6379",
	}

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        clusterAddrs,
		MaxRedirects: 8,
		ReadOnly:     false,
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
