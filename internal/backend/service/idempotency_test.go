package service
import (
	"context"
	"path/filepath"
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
func TestIdempotency(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testredis.Run(ctx, "redis:alpine")
	if err != nil {
		t.Fatalf("Redis container başlatılamadı: %s", err)
	}
	defer redisContainer.Terminate(ctx)
	redisURI, _ := redisContainer.ConnectionString(ctx)
	redisOpts, _ := redis.ParseURL(redisURI)
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()
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
	locker := redislock.NewLocker(rdb)
	repo, _ := repository.NewPostgresReservationRepository(dbPool)
	svc := NewReservationService(locker, repo)
	userID := uuid.New().String()
	tripID := uuid.New().String()
	seatID := "seat_B15"
	idempotencyKey := "idem_key_12345"
	req := &reservationv1.ReserveSeatRequest{
		UserId:         userID,
		TripId:         tripID,
		SeatId:         seatID,
		IdempotencyKey: idempotencyKey,
	}
	resp1, err1 := svc.ReserveSeat(ctx, req)
	if err1 != nil {
		t.Fatalf("İlk çağrıda beklenmeyen hata: %v", err1)
	}
	if !resp1.Success {
		t.Fatalf("İlk çağrı başarısız oldu: %s", resp1.Message)
	}
	t.Logf("İlk çağrı sonucu: %s", resp1.Message)
	resp2, err2 := svc.ReserveSeat(ctx, req)
	if err2 != nil {
		t.Fatalf("İkinci çağrıda beklenmeyen hata: %v", err2)
	}
	if !resp2.Success {
		t.Fatalf("İkinci çağrı (idempotency gereksinimi) success true dönmeliydi ama false döndü: %s", resp2.Message)
	}
	t.Logf("İkinci çağrı (Mükerrer) sonucu: %s", resp2.Message)
	var count int
	err = dbPool.QueryRowContext(ctx, "SELECT COUNT(*) FROM reservations WHERE idempotency_key = $1", idempotencyKey).Scan(&count)
	if err != nil {
		t.Fatalf("DB kontrolü yapılamadı: %v", err)
	}
	if count != 1 {
		t.Fatalf("Veritabanında 1 kayıt olması bekleniyordu ama %d kayıt bulundu!", count)
	}
	t.Logf("DB doğrulaması başarılı! Veritabanında mükerrer kayıt yok, toplam: %d", count)
}
