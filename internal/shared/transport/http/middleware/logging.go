package middleware

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/transport/http/response"
	"log/slog"
	"net/http"
	"time"
)

func Logging(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestHeader)
			rw := httpresponse.New(w)
			start := time.Now()

			l := log.With(
				slog.String("request_id", requestID),
				slog.String("user_agent", r.Header.Get("User-Agent")),
				slog.String("URI", r.RequestURI),
				slog.String("method", r.Method),
			)

			l.Info("HTTP Request started")

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			l.Info(
				"HTTP Request done",
				slog.Int("status", rw.GetStatusCode()),
				slog.Duration("latency", duration),
				slog.Int("size", rw.GetBodySize()),
			)
		})
	}
}
