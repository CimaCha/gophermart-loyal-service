package authentication

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type contextKey struct{}

func AuthMiddleware(log *slog.Logger, userLoginParser UserLoginParser) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			jwtCookie, err := request.Cookie("jwt")
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			userID, err := userLoginParser.GetUserID(jwtCookie.Value)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			request = request.WithContext(context.WithValue(request.Context(), contextKey{}, userID))
			handler.ServeHTTP(writer, request)
		})
	}
}
