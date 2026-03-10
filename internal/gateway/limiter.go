package gateway

import (
	"sync"
	"time"
)

type RateLimiter interface {
	Allow(ip string) bool
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
