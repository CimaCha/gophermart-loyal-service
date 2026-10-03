package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/core/deps"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/http/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Метод для регистрации маршрутов сервиса user

func registerUserRoutes(r chi.Router, deps *deps.Dependencies) {

	regCfg, ok := deps.Cfg.RLS.Routes["register"]
	if !ok {
		panic("register rate limit config not found")
	}

	logCfg, ok := deps.Cfg.RLS.Routes["login"]
	if !ok {
		panic("login rate limit config not found")
	}

	r.With(
		chimiddleware.AllowContentType("application/json"),
		middleware.RateLimit(deps.RLS, regCfg, "register"),
	).Post("/register", deps.UserHandler.RegisterUser)

	r.With(
		chimiddleware.AllowContentType("application/json"),
		middleware.RateLimit(deps.RLS, logCfg, "login"),
	).Post("/login", deps.UserHandler.LoginUser)
}
