package gateway

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Allow(ctx context.Context, ip string) bool
}

type LocalLimiter struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
}

func NewLocalLimiter(capacity, refillRate float64) *LocalLimiter {
	return &LocalLimiter{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (l *LocalLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	duration := now.Sub(l.lastRefill).Seconds()
	l.tokens += duration * l.refillRate
	if l.tokens > l.capacity {
		l.tokens = l.capacity
	}
	l.lastRefill = now
	if l.tokens >= 1.0 {
		l.tokens -= 1.0
		return true
	}
	return false
}

type IpLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*LocalLimiter
	capacity   float64
	refillRate float64
}

func NewIpLimiter(capacity, refillRate float64) *IpLimiter {
	return &IpLimiter{
		buckets:    make(map[string]*LocalLimiter),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (ipl *IpLimiter) Allow(ip string) bool {
	ipl.mu.Lock()
	bucket, exists := ipl.buckets[ip]
	if !exists {
		bucket = NewLocalLimiter(ipl.capacity, ipl.refillRate)
		ipl.buckets[ip] = bucket
	}
	ipl.mu.Unlock()

	return bucket.Allow()
}

type RedisLimiter struct {
	rdb        *redis.Client
	capacity   float64
	refillRate float64
}

func NewRedisLimiter(rdb *redis.Client, cap, rate float64) *RedisLimiter {
	return &RedisLimiter{
		rdb:        rdb,
		capacity:   cap,
		refillRate: rate,
	}
}

func (rl *RedisLimiter) Allow(ctx context.Context, ip string) bool {
	script := `
	local tokens_key = KEYS[1]
	local last_refill_key = KEYS[2]

	local capacity = tonumber(ARGV[1])
	local refill_rate = to number(ARGV[2])
	local now = tonumber(ARGV[3])
	local requested = 1

	local tokens = tonumber(redis.call("get", tokens_key))
	local last_refill = tonumber(redis.call("get", last_refill_key))

	if tokens == nil then
		tokens = capacity
		last_refill = now
	else
		local time_passed= now -last_refill
		tokens= tokens +(tamep)
	`
}
