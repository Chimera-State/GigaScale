package service
import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"github.com/Chimera-State/GigaScale/internal/backend/pkg/db"
	"github.com/Chimera-State/GigaScale/internal/backend/repository"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)
func TestDoubleBookingGuard_UniqueIndex(t *testing.T) {
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts(filepath.Join("..", "..", "..", "migrations", "001_init.up.sql")),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("PostgreSQL container başlatılamadı: %s", err)
	}
	defer pgContainer.Terminate(ctx)
	pgConnString, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
	dbPool, _ := db.NewDatabase(pgConnString)
	defer dbPool.Close()
	repo, err := repository.NewPostgresReservationRepository(dbPool)
	if err != nil {
		t.Fatalf("Repo başlatılamadı: %v", err)
	}
	tripID := uuid.New()
	seatID := "seat_X99"
	res1 := &repository.Reservation{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		TripID:         tripID,
		SeatID:         seatID,
		IdempotencyKey: "req_1",
		Amount:         100.0,
		Status:         "confirmed",
		CreatedAt:      time.Now(),
	}
	err = repo.Create(ctx, res1)
	if err != nil {
		t.Fatalf("1. İstek başarısız oldu: %v", err)
	}
	t.Log("1. İstek: İlk rezervasyon işlemi veritabanına başarıyla yazıldı.")
	res2 := &repository.Reservation{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		TripID:         tripID,
		SeatID:         seatID,
		IdempotencyKey: "req_2",
		Amount:         150.0,
		Status:         "confirmed",
		CreatedAt:      time.Now(),
	}
	err = repo.Create(ctx, res2)
	if err == nil {
		t.Fatal("CRITICAL ERROR: 2. İstek (Müşteri B) veritabanına BAŞARIYLA yazıldı! Çift rezervasyon (Double-booking) ENGELLENEMEDİ!")
	}
	t.Logf("2. İstek (Müşteri B) beklendiği gibi veritabanı tarafından reddedildi. Alınan hata: %v", err)
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "already_booked") && !strings.Contains(errStr, "benzersizlik ihlali") {
		t.Errorf("Hata alındı fakat UNIQUE INDEX kısıtlamasından kaynaklanmıyor. Gelen hata: %v", err)
	} else {
		t.Log("DB Guard: Harika! Veritabanı UNIQUE INDEX kısıtlaması, süresi dolan kilitte bile çift rezervasyonu kesin olarak durdurdu.")
	}
}
