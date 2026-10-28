package httpmiddleware

import (
	"net/http"

	"github.com/google/uuid"
)

// RequestID возвращает Middleware, которое проверяет наличие уникального идентификатора запроса
// в заголовках. Если заголовок requestHeader пуст, middleware генерирует новый UUID (строку)
// и устанавливает его как в заголовки запроса, так и в заголовки ответа.
// Это позволяет сквозным образом отслеживать всю цепочку вызовов (tracing).
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
