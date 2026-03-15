package service
import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	reservationv1 "github.com/Chimera-State/GigaScale/api/proto/reservation/v1"
	"github.com/Chimera-State/GigaScale/internal/backend/pkg/db"
	"github.com/Chimera-State/GigaScale/internal/backend/pkg/redislock"
	"github.com/Chimera-State/GigaScale/internal/backend/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	testredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)
func TestConcurrentReservationOperationsRealLock(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testredis.Run(ctx, "redis:alpine")
	if err != nil {
		t.Fatalf("Redis container başlatılamadı: %s", err)
	}
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("Redis container kapatılamadı: %s", err)
		}
	}()
	redisURI, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Redis connection string alınamadı: %s", err)
	}
	redisOpts, err := redis.ParseURL(redisURI)
	if err != nil {
		t.Fatalf("Redis URI parse hatası: %s", err)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts(filepath.Join("..", "..", "..", "migrations", "001_init.sql")),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("PostgreSQL container başlatılamadı: %s", err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("PostgreSQL container kapatılamadı: %s", err)
		}
	}()
	pgConnString, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Postgres connection string alınamadı: %s", err)
	}
	dbPool, err := db.NewDatabase(pgConnString)
	if err != nil {
		t.Fatalf("Veritabanı bağlantısı kurulamadı: %v", err)
	}
	defer dbPool.Close()
	locker := redislock.NewLocker(rdb)
	repo, err := repository.NewPostgresReservationRepository(dbPool)
	if err != nil {
		t.Fatalf("Repository başlatılamadı: %v", err)
	}
	svc := NewReservationService(locker, repo)
	const totalConcurrentRequests = 50
	var wg sync.WaitGroup
	var successCount int
	var failureCount int
	var mu sync.Mutex
	targetTripID := uuid.New().String()
	targetSeatID := "seat_A12"
	for i := 0; i < totalConcurrentRequests; i++ {
		wg.Add(1)
		go func(requestID int) {
			defer wg.Done()
			userID := uuid.New().String()
			idempotencyKey := fmt.Sprintf("req_key_%d", requestID)
			req := &reservationv1.ReserveSeatRequest{
				UserId:         userID,
				TripId:         targetTripID,
				SeatId:         targetSeatID,
				IdempotencyKey: idempotencyKey,
			}
			resp, err := svc.ReserveSeat(context.Background(), req)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				t.Errorf("Beklenmeyen sistem hatası: %v", err)
			} else {
				if resp.Success {
					successCount++
					t.Logf("İstek %s -> 200 OK: Koltuk başarıyla alındı!", userID)
				} else {
					failureCount++
					if strings.Contains(resp.Message, "Sistem şu anda aşırı yoğun") ||
						strings.Contains(resp.Message, "Locked") ||
						strings.Contains(resp.Message, "Bu koltuk maalesef satılmış") {
						t.Logf("İstek %s -> %s", userID, resp.Message)
					} else {
						t.Errorf("Beklenmeyen Hata Mesajı: %s", resp.Message)
					}
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
		t.Logf("Testcontainers (Redis + Postgres) Testi başarıyla sonuçlandı. Koltuk güvenli bir şekilde sadece 1 kişiye tahsis edildi.")
	}
	var dbCount int
	err = dbPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM reservations WHERE trip_id = $1 AND seat_id = $2", targetTripID, targetSeatID).Scan(&dbCount)
	if err != nil {
		t.Fatalf("Veritabanı Count sorgusu hata verdi: %v", err)
	}
	if dbCount != 1 {
		t.Errorf("CRITICAL ERROR: Veritabanında (PostgreSQL) %d adet başarılı rezervasyon kaydı var, 1 olmalıydı!", dbCount)
	} else {
		t.Logf("DB Doğrulaması Başarılı: Veritabanında sadece %d adet kayıt var.", dbCount)
	}
}
<<<<<<< HEAD

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
=======
>>>>>>> origin/fetaure/sprint3
