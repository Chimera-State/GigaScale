package gateway

import (
	"sync"
	"time"
)

type RateLimiter interface {
	Allow() bool
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
		refillRate: 0.5,
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
