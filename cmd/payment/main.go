package main
import (
	"log"
	"net"
	"os"
	pb "github.com/Chimera-State/GigaScale/api/proto/payment/v1"
	payment "github.com/Chimera-State/GigaScale/internal/payment/service"
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
	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, payment.NewPaymentService())
	reflection.Register(s)
	log.Printf("Payment gRPC server listening on %s", lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
