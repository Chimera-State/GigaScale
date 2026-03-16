package redislock
import (
	"context"
	"testing"
	"time"
	"github.com/redis/go-redis/v9"
	testredis "github.com/testcontainers/testcontainers-go/modules/redis"
)
func TestLockIntegration(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testredis.Run(ctx, "redis:alpine")
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to stop Redis container: %v", err)
		}
	}()
	redisURI, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get Redis connection string: %v", err)
	}
	opts, err := redis.ParseURL(redisURI)
	if err != nil {
		t.Fatalf("Failed to parse Redis URI: %v", err)
	}
	rdb := redis.NewClient(opts)
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}
	locker := NewLocker(rdb)
	t.Run("Lock Acquire and Release", func(t *testing.T) {
		key := "test_lock_1"
		ttl := 2 * time.Second
		token, acquired, err := locker.Acquire(ctx, key, ttl)
		if err != nil {
			t.Fatalf("Acquire failed with error: %v", err)
		}
		if !acquired {
			t.Fatal("Expected to acquire lock, but failed")
		}
		if token == "" {
			t.Fatal("Expected token to be non-empty")
		}
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			t.Fatalf("Expected key to exist in Redis, got error: %v", err)
		}
		if val != token {
			t.Fatalf("Expected token %s in Redis, got %s", token, val)
		}
		_, acquiredAgain, err := locker.Acquire(ctx, key, ttl)
		if err != nil {
			t.Fatalf("Second acquire failed with error: %v", err)
		}
		if acquiredAgain {
			t.Fatal("Expected second acquire to fail, but it succeeded")
		}
		err = locker.Release(ctx, key, token)
		if err != nil {
			t.Fatalf("Release failed with error: %v", err)
		}
		_, err = rdb.Get(ctx, key).Result()
		if err != redis.Nil {
			t.Fatalf("Expected key to be missing in Redis after release, got error/value: %v", err)
		}
	})
	t.Run("Token Validation during Release", func(t *testing.T) {
		key := "test_lock_token_validation"
		ttl := 2 * time.Second
		_, acquired, err := locker.Acquire(ctx, key, ttl)
		if err != nil || !acquired {
			t.Fatalf("Failed to acquire lock: %v", err)
		}
		fakeToken := "invalid-token"
		err = locker.Release(ctx, key, fakeToken)
		if err != nil {
			t.Fatalf("Release with wrong token should return nil (it just doesn't delete), got err: %v", err)
		}
		exists, err := rdb.Exists(ctx, key).Result()
		if err != nil || exists == 0 {
			t.Fatal("Lock was incorrectly removed with wrong token!")
		}
	})
	t.Run("TTL Expiration Validation", func(t *testing.T) {
		key := "test_lock_ttl"
		ttl := 1 * time.Second
		_, acquired, err := locker.Acquire(ctx, key, ttl)
		if err != nil || !acquired {
			t.Fatalf("Failed to acquire lock: %v", err)
		}
		time.Sleep(1500 * time.Millisecond)
		exists, err := rdb.Exists(ctx, key).Result()
		if err != nil || exists != 0 {
			t.Fatal("Lock did not expire according to TTL!")
		}
		_, acquiredNew, err := locker.Acquire(ctx, key, ttl)
		if err != nil || !acquiredNew {
			t.Fatalf("Could not acquire lock after TTL expiration: %v", err)
		}
	})
}
