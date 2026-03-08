package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	reservationv1 "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/Chimera-State/GigaScale/internal/backend/service"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("soket dinlenemedi: %v", err)
	}
	loggingInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Printf("Gelen İstek: %s", info.FullMethod)
		return handler(ctx, req)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)
	myService := service.NewReservationService()

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
