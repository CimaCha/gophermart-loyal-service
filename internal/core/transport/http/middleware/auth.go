package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey struct{}

//go:generate go tool mockgen -source=auth.go -destination=mock/token_validator_gen.go -package=mock

type TokenValidator interface {
	ValidateUserID(jwt string) (uuid.UUID, error)
}

func AuthMiddleware(tokenValidator TokenValidator) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			jwtCookie, err := request.Cookie("jwt")
			if err != nil {
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

func UserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(contextKey{}).(uuid.UUID)
	return userID, ok
}
