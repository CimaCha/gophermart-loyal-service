package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/ratelimitstore"
)

type RateLimiter interface {
	Incr(key string, window time.Duration) (int, time.Time)
}

// GlobalRateLimit ограничивает суммарное число запросов ко всему сервису
func GlobalRateLimit(limiter RateLimiter, cfg ratelimitstore.RateLimitConfig, keyPrefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := fmt.Sprintf("rate:%s:global", keyPrefix)
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
