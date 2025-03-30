package tinyurl

import (
	"fmt"
	"net"
	"net/http"
)

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.rateLimiter.Enabled() {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, fmt.Sprintf("parsing ip: %v", err), http.StatusInternalServerError)
				return
			}

			if ok := s.rateLimiter.Allow(ip); !ok {
				http.Error(w, "too many request", http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
