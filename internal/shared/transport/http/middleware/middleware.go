package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

const (
	auth          = "Authorization"
	requestHeader = "X-Request-ID"
)
