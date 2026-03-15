package payment
import (
	"context"
	"math/rand"
	"os"
	"time"
	pb "github.com/Chimera-State/GigaScale/api/proto/payment/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
type PaymentServiceServer struct {
	pb.UnimplementedPaymentServiceServer
}
func NewPaymentService() *PaymentServiceServer {
	return &PaymentServiceServer{}
}
func (s *PaymentServiceServer) Charge(ctx context.Context, req *pb.ChargeRequest) (*pb.ChargeResponse, error) {
	if req.UserId == "" || req.Amount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "gecersiz user_id veya miktar")
	}
	time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)
	failSimulation := os.Getenv("PAYMENT_SIMULATE_FAIL")
	if failSimulation == "true" {
		failChance := rand.Intn(100)
		if failChance < 30 {
			return &pb.ChargeResponse{
				Success:   false,
				PaymentId: "",
				Message:   "odeme reddedildi - bakiye yetersiz veya kart hatasi",
			}, nil
		}
	}
	paymentID := uuid.New().String()
	return &pb.ChargeResponse{
		Success:   true,
		PaymentId: paymentID,
		Message:   "odeme basariyla gerceklestirildi",
	}, nil
}
func (s *PaymentServiceServer) Refund(ctx context.Context, req *pb.RefundRequest) (*pb.RefundResponse, error) {
	return &pb.RefundResponse{
		Success: true,
		Message: "iade islemi basariyla tamamlandi",
	}, nil
}
