package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Chimera-State/GigaScale/internal/backend/reservationv1"
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
	//1.Create gRPC server
	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	// Önemli: Buraya servis kayıt (Register) kodlarını ekleyeceksin.
	// 1. Önce aşçımızı (servisi) işe alıyoruz
	myService := service.NewReservationService()

	// 2. Aşçımızı (myService) restoranın (s) menüsüne kaydediyoruz.
	reservationv1.RegisterReservationServiceServer(s, myService)

	// 4. Graceful Shutdown Hazırlığı
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	// Sunucuyu arka planda başlat
	go func() {
		log.Printf("gRPC sunucusu :50051 portunda başladı...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Sunucu hatası: %v", err)
		}
	}()
	// Kapatma sinyali gelene kadar blokla
	<-stop
	log.Println("\nKapatma sinyali alındı. Sunucu güvenli bir şekilde kapatılıyor...")
	// Sunucuyu zarifçe kapat (İşlemlerin bitmesini bekler)
	s.GracefulStop()
	log.Println("Sunucu tamamen durduruldu.")
}
