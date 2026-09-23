package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/core/deps"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/http/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Метод для регистрации маршрутов сервиса user

func registerUserRoutes(r chi.Router, deps *deps.Dependencies) {
	r.With(
		chimiddleware.AllowContentType("application/json"),
		middleware.RateLimit(deps.Limiter, deps.Cfg.RLS.Register),
	).Post("/register", deps.UserHandler.RegisterUser)

	r.With(
		chimiddleware.AllowContentType("application/json"),
		middleware.RateLimit(deps.Limiter, deps.Cfg.RLS.Login),
	).Post("/login", deps.UserHandler.LoginUser)
}
