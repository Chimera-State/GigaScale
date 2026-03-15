package service

import (
	"context"
	"fmt"
	"log"
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

	log.Println("Gelen Rezervasyon İsteği:")
	log.Printf("- Kullanıcı ID: %s\n", userID)
	log.Printf("- Trip ID: %s\n", tripID)
	log.Printf("- Koltuk ID: %s\n", seatID)
	log.Printf("- Idempotency Key: %s\n", idempotencyKey)

	var isSuccess bool

	if idempotencyKey != "" {
		idempotencyKeyRedis := "idempotency:" + idempotencyKey
		isNew, err := s.locker.CheckIdempotency(ctx, idempotencyKeyRedis, 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("idempotency kontrol hatası: %w", err)
		}
		if !isNew {
			return &reservationv1.ReserveSeatResponse{
				Success: true,
				Message: "Mükerrer İstek (Idempotency): İşleminiz sistemde zaten başarılı şekilde kaydedilmiş.",
			}, nil
		}

		// Rollback if the overarching operation is not successful
		defer func() {
			if !isSuccess {
				_ = s.locker.RemoveIdempotency(ctx, idempotencyKeyRedis)
			}
		}()
	}

	lockKey := fmt.Sprintf("lock:reservation:trip:%s:seat:%s", tripID, seatID)
	lockTTL := 2 * time.Second

	retryConfig := redislock.RetryConfig{
		MaxRetries: 4,
		RetryDelay: 100 * time.Millisecond,
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

	stateKey := fmt.Sprintf("seat:booked:trip:%s:seat:%s", tripID, seatID)
	bookedState, err := s.locker.GetState(ctx, stateKey)
	if err != nil {
		return nil, fmt.Errorf("state kontrol hatası: %w", err)
	}

	if bookedState != "" {
		return &reservationv1.ReserveSeatResponse{
			Success: false,
			Message: "Locked (Koltuk Başka Bir İşlem Tarafından Rezerve Edildi)",
		}, nil
	}

	err = s.locker.SetState(ctx, stateKey, userID, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("state yazma hatası: %w", err)
	}

	isSuccess = true
	return &reservationv1.ReserveSeatResponse{
		Success: true,
		Message: "İşlem başarılı. Kilit ile güvenli bir şekilde alındı.",
	}, nil
}
