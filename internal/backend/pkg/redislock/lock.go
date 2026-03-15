package redislock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Locker struct {
	client *redis.Client
}

func NewLocker(client *redis.Client) *Locker {
	return &Locker{
		client: client,
	}
}

func (l *Locker) Acquire(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	token := uuid.New().String()

	resp, err := l.client.SetArgs(ctx, key, token, redis.SetArgs{Mode: "NX", TTL: ttl}).Result()

	if err == redis.Nil {
		// Key already exists (Mode: "NX" failed)
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}

	return token, resp == "OK", nil
}

type RetryConfig struct {
	MaxRetries int
	RetryDelay time.Duration
}

func (l *Locker) AcquireWithRetry(ctx context.Context, key string, ttl time.Duration, retry RetryConfig) (string, bool, error) {
	attempts := retry.MaxRetries + 1

	for i := 0; i < attempts; i++ {
		token, acquired, err := l.Acquire(ctx, key, ttl)
		if err != nil {
			return "", false, err
		}

		if acquired {
			return token, true, nil
		}

		if i < attempts-1 {
			select {
			case <-ctx.Done():
				return "", false, ctx.Err()
			case <-time.After(retry.RetryDelay):
			}
		}
	}

	return "", false, nil
}

const releaseScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`

func (l *Locker) Release(ctx context.Context, key string, token string) error {
	err := l.client.Eval(ctx, releaseScript, []string{key}, token).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func (l *Locker) CheckIdempotency(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	resp, err := l.client.SetArgs(ctx, key, "processed", redis.SetArgs{Mode: "NX", TTL: ttl}).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return resp == "OK", nil
}
