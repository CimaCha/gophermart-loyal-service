// Package httpmiddleware предоставляет промежуточное программное обеспечение (middleware)
// для обработки HTTP-запросов и ответов в микросервисе gophermart.
package httpmiddleware

import (
	"net/http"
)

// Middleware определяет сигнатуру функции промежуточного слоя,
// которая принимает один http.Handler и возвращает другой.
type Middleware func(http.Handler) http.Handler

const (
	auth          = "Authorization"
	requestHeader = "X-Request-ID"
)
