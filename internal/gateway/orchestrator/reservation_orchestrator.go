package orchestrator

import (
	"context"
	"fmt"
	"log"

	pbpay "github.com/Chimera-State/GigaScale/api/proto/payment/v1"
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

	reserveResp, err := o.reserveClient.ReserveSeat(ctx, &pbreserv.ReserveSeatRequest{
		UserId:         req.UserID,
		TripId:         req.TripID,
		SeatId:         req.SeatID,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return nil, fmt.Errorf("reservation system error: %w", err)
	}
	if !reserveResp.Success {
		return &ReserveResult{
			Success: false,
			Message: reserveResp.Message,
		}, nil
	}

	chargeResp, err := o.paymentClient.Charge(ctx, &pbpay.ChargeRequest{
		UserId:         req.UserID,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil || !chargeResp.Success {
		log.Printf("[SAGA] Payment failed, reservation.The reservation is being cancelled. error : %v", err)

		_, cancelErr := o.reserveClient.CancelReservation(ctx, &pbreserv.CancelRequest{
			IdempotencyKey: req.IdempotencyKey,
		})
		if cancelErr != nil {
			log.Printf("[SAGA] CRITICAL: Cancellation failed: %v", cancelErr)
		}

		msg := "Payment failed. Reservation cancelled."
		if err == nil && chargeResp != nil {
			msg = chargeResp.Message
		}
		return &ReserveResult{Success: false, Message: msg}, nil
	}

	publishErr := o.publisher.Publish(ctx, mq.ReservationCreatedEvent{
		UserID:         req.UserID,
		TripID:         req.TripID,
		SeatID:         req.SeatID,
		PaymentID:      chargeResp.PaymentId,
		IdempotencyKey: req.IdempotencyKey,
	})
	if publishErr != nil {
		log.Printf("[KAFKA] Event cannot be sent (not critical): %v", publishErr)
	}

	return &ReserveResult{
		Success:   true,
		Message:   "Reservation and payment completed.",
		PaymentID: chargeResp.PaymentId,
	}, nil
}
