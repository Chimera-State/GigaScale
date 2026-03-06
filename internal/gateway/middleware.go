package gateway

import (
	"net/http"
	"time"
)

// ratelimit anonim önleme
type rateLimitHandler struct {
	server *Server
	next   http.Handler
}

func (s *Server) RateLimiter(next http.Handler) http.Handler {
	return &rateLimitHandler{
		server: s,
		next:   next,
	}
}

func (h *rateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.server.mu.Lock()

	now := time.Now()
	duration := now.Sub(h.server.lastRefill).Seconds()
	h.server.tokens += duration * h.server.refillRate

	if h.server.tokens > h.server.capacity {
		h.server.tokens = h.server.capacity
	}
	h.server.lastRefill = now

	if h.server.tokens < 1 {
		h.server.mu.Unlock()
		http.Error(w, "GigaScale: Too many requests! Please wait.", http.StatusTooManyRequests)
		return
	}

	h.server.tokens--
	h.server.mu.Unlock()

	h.next.ServeHTTP(w, r)

}
