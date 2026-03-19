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
	acquired, err := l.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", false, err
	}
	return token, acquired, nil
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
	return l.client.SetNX(ctx, key, "processed", ttl).Result()
}

func (l *Locker) RemoveIdempotency(ctx context.Context, key string) error {
	err := l.client.Del(ctx, key).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func (l *Locker) GetState(ctx context.Context, key string) (string, error) {
	val, err := l.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return val, nil
}

func (l *Locker) SetState(ctx context.Context, key string, value string, ttl time.Duration) error {
	return l.client.Set(ctx, key, value, ttl).Err()
}
