package gateway

import (
	"net/http"
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
	isAllowed := h.server.limiter.Allow()

	if !isAllowed {
		http.Error(w, "GigaScale: Too many requests! Please wait.", http.StatusTooManyRequests)
		return
	}

	h.next.ServeHTTP(w, r)

}
