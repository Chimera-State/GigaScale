package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/Chimera-State/GigaScale/api/proto/payment/v1"
	payment "github.com/Chimera-State/GigaScale/internal/payment/service"
	"github.com/Chimera-State/go-otel-kit/interceptor"
	"github.com/Chimera-State/go-otel-kit/setup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50052"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	ctx := context.Background()
	if err := setup.Init(ctx,
		setup.WithServiceName("gigascale-payment"),
		setup.WithServiceVersion("1.0.0"),
		setup.WithExporterEndpoint("otel-collector:4317"),
	); err != nil {
		log.Fatalf("OTel initialization failed: %v", err)
	}
	defer setup.Shutdown(ctx)

	s := grpc.NewServer(interceptor.ServerOptions()...)
	pb.RegisterPaymentServiceServer(s, payment.NewPaymentService())
	reflection.Register(s)
	log.Printf("Payment gRPC server listening on %s", lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
