package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

// RequestID Просто проставляет в Header X-Request-ID, чтобы можно было логировать цепочку запросов
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
