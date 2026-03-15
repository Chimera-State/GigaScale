package service

import (
	"fmt"
	"sync"
	"testing"
	"time"

	reservationv1 "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
)

func ProcessData() error {
	time.Sleep(10 * time.Millisecond)
	return nil
}

func TestConcurrentReservationOperations(t *testing.T) {
	const totalConcurrentRequests = 50

	var wg sync.WaitGroup
	var successCount int
	var failureCount int
	var mu sync.Mutex

	targetTripID := "trip_ankara_ist_1001"
	targetSeatID := "seat_A12"

	for i := 0; i < totalConcurrentRequests; i++ {
		wg.Add(1)

		go func(requestID int) {
			defer wg.Done()

			userID := fmt.Sprintf("user_%d", requestID)
			idempotencyKey := fmt.Sprintf("req_key_%d", time.Now().UnixNano())

			req := &reservationv1.ReserveSeatRequest{
				UserId:         userID,
				TripId:         targetTripID,
				SeatId:         targetSeatID,
				IdempotencyKey: idempotencyKey,
			}

			resp, err := mockReserveSeat(req)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if apiErr, ok := err.(*APIError); ok {
					if apiErr.StatusCode == 423 {
						failureCount++
						t.Logf("İstek %s -> 423 Locked: %s", userID, apiErr.Message)
					} else {
						t.Errorf("Beklenmeyen Hata Kodu. Beklenen: 423, Gelen: %d", apiErr.StatusCode)
					}
				} else {
					t.Errorf("Beklenmeyen sistem hatası: %v", err)
				}
			} else {
				if resp.Success {
					successCount++
					t.Logf("İstek %s -> 200 OK: Koltuk başarıyla alındı!", userID)
				} else {
					t.Errorf("İşlem başarılı gözüküyor ama Success bayrağı false!")
				}
			}
		}(i)
	}

	wg.Wait()

	t.Logf("--- TEST SONUCU ---")
	t.Logf("Toplam İstek: %d", totalConcurrentRequests)
	t.Logf("Başarılı Sayısı: %d", successCount)
	t.Logf("Başarısız Sayısı: %d", failureCount)

	if successCount > 1 {
		t.Errorf("CRITICAL ERROR: Aynı koltuk %d kişiye birden satıldı! Lock mekanizması hatalı.", successCount)
	} else if successCount == 0 {
		t.Errorf("Hata: Koltuğu kimse alamadı, sistem gereksiz yere kitlenmiş olabilir.")
	} else {
		t.Logf("Test başarıyla sonuçlandı. Koltuk güvenli bir şekilde sadece 1 kişiye tahsis edildi.")
	}
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

var mockSeatTaken bool
var mockMu sync.Mutex

func mockReserveSeat(req *reservationv1.ReserveSeatRequest) (*reservationv1.ReserveSeatResponse, error) {
	_ = req
	mockMu.Lock()
	defer mockMu.Unlock()

	time.Sleep(5 * time.Millisecond)

	if mockSeatTaken {
		return nil, &APIError{
			StatusCode: 423,
			Message:    "Locked (Koltuk başka bir işleme tahsis edildi)",
		}
	}

	mockSeatTaken = true
	return &reservationv1.ReserveSeatResponse{
		Success: true,
		Message: "OK (Koltuk başarıyla rezerve edildi)",
	}, nil
}
