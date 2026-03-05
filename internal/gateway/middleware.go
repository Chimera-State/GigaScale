package gateway

import "net/http"

func (s *Server) RateLimiter(next http.Handler)
