package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/core/auth/jwtoken"
	"github.com/CimaCha/gophermart-loyal-service/internal/core/transport/http/ctxkeys"
	httpresponse "github.com/CimaCha/gophermart-loyal-service/internal/core/transport/http/response"
	"github.com/google/uuid"
)

type Middleware func(http.Handler) http.Handler

type TokenValidator interface {
	ValidateToken(tokenString string) (*jwtoken.Claims, error)
}

const (
	auth          = "Authorization"
	requestHeader = "X-Request-ID"
)

// Просто проставляет в Header X-Request-ID, чтобы можно было логировать цепочку запросов
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestHeader)
			if requestID == "" {
				requestID = uuid.NewString()
			}

			r.Header.Set(requestHeader, requestID)
			w.Header().Set(requestHeader, requestID)

			next.ServeHTTP(w, r)
		})
	}
}

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

func Auth(validator TokenValidator, log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get(auth)
			if authHeader == "" {

				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {

				return
			}

			tokenString := parts[1]

			claims, err := validator.ValidateToken(tokenString)

			if err != nil {
				log.Info(
					"token validation failed",
					"err", err,
				)
				http.Error(
					w,
					"invalid or expiration token",
					http.StatusUnauthorized,
				)
				return
			}

			log.Info(
				"user authenticated",
			)

			ctx := ctxkeys.WithUserID(r.Context(), claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
