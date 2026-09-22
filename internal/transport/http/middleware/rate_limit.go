package middleware

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/ratelimit"
)

type RateLimiter interface {
	Incr(key string, window time.Duration) (int, time.Time)
}

func RateLimit(limiter RateLimiter, cfg ratelimit.RateLimitConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			key := fmt.Sprintf("rate:%s:%s", cfg.KeyPrefix, ip)
			count, exp := limiter.Incr(key, cfg.Window)

			if count > cfg.MaxRequests {
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", time.Until(exp).Seconds()))
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
