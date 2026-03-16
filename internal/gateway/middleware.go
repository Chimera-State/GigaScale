package gateway

import (
	"log"
	"net"
	"net/http"
	"strings"
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
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host

}

func (h *rateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ip := clientIP(r)

	log.Printf("[REQUESTED] ip=%s method=%s path=%s time=%s", ip, r.Method, r.URL.Path, start.Format(time.RFC3339))

	ctx := r.Context()
	isAllowed := h.server.limiter.Allow(ctx, ip)

	if !isAllowed {
		http.Error(w, "GigaScale: Too many requests! Please wait.", http.StatusTooManyRequests)
		return
	}

	h.next.ServeHTTP(w, r)

}
