package middleware

import (
	"log/slog"
	"net/http"
	"time"

	httpresponse "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/http/response"
)

func Logging(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestid := r.Header.Get(requestHeader)
			rw := httpresponse.New(w)
			start := time.Now()

			l := log.With(
				slog.String("request_id", requestid),
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
