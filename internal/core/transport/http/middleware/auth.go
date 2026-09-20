package middleware

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net/http"
)

type contextKey struct{}

type TokenValidator interface {
	ValidateUserID(jwt string) (uuid.UUID, error)
}

func AuthMiddleware(tokenValidator TokenValidator) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			jwtCookie, err := request.Cookie("jwt")
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			userID, err := tokenValidator.ValidateUserID(jwtCookie.Value)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			request = request.WithContext(context.WithValue(request.Context(), contextKey{}, userID))
			handler.ServeHTTP(writer, request)
		})
	}
}
