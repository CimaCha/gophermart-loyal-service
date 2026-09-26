package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/core/deps"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/transport/http/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Метод для регистрации маршрутов для сервиса order

func registerOrdersRoutes(r chi.Router, deps *deps.Dependencies) {
	r.With(
		chimiddleware.AllowContentType("application/json"),
	).Post("/", deps.OrdersHandler.CreateOrder)

	r.With(
		chimiddleware.AllowContentType("application/json"),
		middleware.RateLimit(deps.Limiter, deps.Cfg.RLS.Register),
	).Get("/{number}", deps.OrdersHandler.GetOrders)
}
