package orchestrator

import (
	"context"
	"fmt"
	"log"

	pbpay    "github.com/Chimera-State/GigaScale/api/proto/payment/v1"
	pbreserv "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/Chimera-State/GigaScale/internal/gateway/mq"
)

type ReservationOrchestrator struct {
	reserveClient pbreserv.ReservationServiceClient
	paymentClient pbpay.PaymentServiceClient
	publisher     *mq.Publisher
}

func New(
	r pbreserv.ReservationServiceClient,
	p pbpay.PaymentServiceClient,
	publisher *mq.Publisher,
) *ReservationOrchestrator {
	return &ReservationOrchestrator{
		reserveClient: r,
		paymentClient: p,
		publisher:     publisher,
	}
}

type ReserveRequest struct {
	UserID         string
	TripID         string
	SeatID         string
	IdempotencyKey string
	Amount         float64
}

type ReserveResult struct {
	Success   bool
	Message   string
	PaymentID string
}

func (o *ReservationOrchestrator) ReserveWithPayment(ctx context.Context, req ReserveRequest) (*ReserveResult, error) {

	// ADIM 1: Koltuğu rezerve et
	reserveResp, err := o.reserveClient.ReserveSeat(ctx, &pbreserv.ReserveSeatRequest{
		UserId:         req.UserID,
		TripId:         req.TripID,
		SeatId:         req.SeatID,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return nil, fmt.Errorf("rezervasyon sistem hatası: %w", err)
	}
	if !reserveResp.Success {
		return &ReserveResult{
			Success: false,
			Message: reserveResp.Message,
		}, nil
	}

	// ADIM 2: Ödeme al
	chargeResp, err := o.paymentClient.Charge(ctx, &pbpay.ChargeRequest{
		UserId:         req.UserID,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil || !chargeResp.Success {
		// COMPENSATION: Ödeme başarısız → rezervasyonu geri al
		log.Printf("[SAGA] Ödeme başarısız, rezervasyon iptal ediliyor. err=%v", err)

		_, cancelErr := o.reserveClient.CancelReservation(ctx, &pbreserv.CancelRequest{
			IdempotencyKey: req.IdempotencyKey,
		})
		if cancelErr != nil {
			log.Printf("[SAGA] KRİTİK: İptal de başarısız: %v", cancelErr)
		}

		msg := "Ödeme başarısız. Rezervasyon iptal edildi."
		if err == nil && chargeResp != nil {
			msg = chargeResp.Message
		}
		return &ReserveResult{Success: false, Message: msg}, nil
	}

	// ADIM 3: Kafka'ya başarı event'i gönder
	publishErr := o.publisher.Publish(ctx, mq.ReservationCreatedEvent{
		UserID:         req.UserID,
		TripID:         req.TripID,
		SeatID:         req.SeatID,
		PaymentID:      chargeResp.PaymentId,
		IdempotencyKey: req.IdempotencyKey,
	})
	if publishErr != nil {
		// Kafka hatası kullanıcıya döndürülmez — sadece loglanır
		// Rezervasyon + ödeme başarılı, Kafka geçici sorun yaratmamalı
		log.Printf("[KAFKA] Event gönderilemedi (kritik değil): %v", publishErr)
	}

	return &ReserveResult{
		Success:   true,
		Message:   "Rezervasyon ve ödeme tamamlandı.",
		PaymentID: chargeResp.PaymentId,
	}, nil
}
