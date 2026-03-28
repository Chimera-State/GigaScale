package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	reservationv1 "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/Chimera-State/GigaScale/internal/backend/pkg/redislock"
	"github.com/Chimera-State/GigaScale/internal/backend/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReservationService struct {
	reservationv1.UnimplementedReservationServiceServer
	locker *redislock.Locker
	repo   repository.ReservationRepository
}

func NewReservationService(locker *redislock.Locker, repo repository.ReservationRepository) *ReservationService {
	return &ReservationService{
		locker: locker,
		repo:   repo,
	}
}
func (s *ReservationService) ReserveSeat(ctx context.Context, req *reservationv1.ReserveSeatRequest) (*reservationv1.ReserveSeatResponse, error) {
	userID := req.GetUserId()
	tripID := req.GetTripId()
	seatID := req.GetSeatId()
	idempotencyKey := req.GetIdempotencyKey()
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
		// Eğer idempotency key ile geldiyse ve koltuğu daha önce DA aynı kullanıcı rezerve etmişse → idempotent 200
		if idempotencyKey != "" && bookedState == userID {
			return &reservationv1.ReserveSeatResponse{
				Success: true,
				Message: "Mükerrer İstek (Idempotency): İşleminiz sistemde zaten başarılı şekilde kaydedilmiş.",
			}, nil
		}
		// Aksi halde farklı bir istek bu koltuğu almış demektir → 409
		return &reservationv1.ReserveSeatResponse{
			Success: false,
			Message: "Locked (Koltuk Başka Bir İşlem Tarafından Rezerve Edildi)",
		}, nil
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("geçersiz user_id tabanlı uuid formati: %w", err)
	}
	tid, err := uuid.Parse(tripID)
	if err != nil {
		return nil, fmt.Errorf("geçersiz trip_id formati: %w", err)
	}

	// ... (Önceki kısımlar aynı)

	newReservation := &repository.Reservation{
		ID:             uuid.New(),
		UserID:         uid,
		TripID:         tid,
		SeatID:         seatID,
		IdempotencyKey: idempotencyKey,
		Amount:         100.0,
		Status:         "confirmed",
		CreatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, newReservation); err != nil {
		log.Printf("[DATABASE ERROR] Trip: %s, Seat: %s, Error: %v", tripID, seatID, err)

		if strings.Contains(err.Error(), "ALREADY_BOOKED") {
			return nil, status.Error(codes.AlreadyExists, "Bu koltuk maalesef çoktan satıldı.")
		}
		return nil, status.Errorf(codes.Internal, "Beklenmedik veritabanı hatası: %v", err)
	}

	err = s.locker.SetState(ctx, stateKey, userID, 24*time.Hour)

	if err != nil {
		return nil, fmt.Errorf("state yazma hatası: %w", err)
	}
	isSuccess = true
	return &reservationv1.ReserveSeatResponse{
		Success: true,
		Message: "İşlem başarılı. Rezervasyonunuz oluşturuldu.",
	}, nil
}
func (s *ReservationService) CancelReservation(ctx context.Context, req *reservationv1.CancelRequest) (*reservationv1.CancelResponse, error) {
	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		return &reservationv1.CancelResponse{
			Success: false,
			Message: "idempotency_key gereklidir",
		}, nil
	}
	err := s.repo.Cancel(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("iptal islemi veritabani hatasi: %w", err)
	}
	err = s.locker.RemoveIdempotency(ctx, "idempotency:"+idempotencyKey)
	if err != nil {
		log.Printf("Redis'ten silme hatasi: %v", err)
	}
	return &reservationv1.CancelResponse{
		Success: true,
		Message: "Rezervasyon basariyla iptal edildi",
	}, nil
}
