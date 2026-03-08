package service

import (
	"context"
	"fmt"
	"time"

	reservationv1 "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/Chimera-State/GigaScale/internal/backend/pkg/redislock"
)

type ReservationService struct {
	reservationv1.UnimplementedReservationServiceServer
	locker *redislock.Locker
}

func NewReservationService(locker *redislock.Locker) *ReservationService {
	return &ReservationService{
		locker: locker,
	}
}

func (s *ReservationService) ReserveSeat(ctx context.Context, req *reservationv1.ReserveSeatRequest) (*reservationv1.ReserveSeatResponse, error) {
	userID := req.GetUserId()
	tripID := req.GetTripId()
	seatID := req.GetSeatId()
	idempotencyKey := req.GetIdempotencyKey()

	println("Gelen Rezervasyon İsteği:")
	println("- Kullanıcı ID:", userID)
	println("- Trip ID:", tripID)
	println("- Koltuk ID:", seatID)
	println("- Idempotency Key:", idempotencyKey)

	lockKey := fmt.Sprintf("lock:reservation:trip:%s:seat:%s", tripID, seatID)
	lockTTL := 5 * time.Second

	retryConfig := redislock.RetryConfig{
		MaxRetries: 4,
		RetryDelay: 4 * time.Second,
	}

	lockToken, acquired, err := s.locker.AcquireWithRetry(ctx, lockKey, lockTTL, retryConfig)
	if err != nil {
		return nil, fmt.Errorf("sistem hatası, kilit kontrol edilemedi: %w", err)
	}

	if !acquired {
		return &reservationv1.ReserveSeatResponse{
			Success: false,
			Message: "Sistem şu anda aşırı yoğun veya koltuk çoktan satıldı. Lütfen daha sonra tekrar deneyin.",
		}, nil
	}

	defer s.locker.Release(ctx, lockKey, lockToken)

	return &reservationv1.ReserveSeatResponse{
		Success: true,
		Message: "İşlem Başarılı. Kilit ile güvenli bir şekilde alındı.",
	}, nil
}
