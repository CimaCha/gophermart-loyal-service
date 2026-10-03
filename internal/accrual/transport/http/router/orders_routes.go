package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/core/deps"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/transport/http/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Метод для регистрации маршрутов для сервиса order

func registerOrdersRoutes(r chi.Router, deps *deps.Dependencies) {
	r.With(
		chimiddleware.AllowContentType("application/json"),
	).Post("/", deps.OrdersHandler.CreateOrder)

	getorderCfg, ok := deps.Cfg.RLS.Routes["getorders"]
	if !ok {
		panic("getorders rate limit config not found")
	}

	r.With(
		chimiddleware.AllowContentType("application/json"),
		middleware.GlobalRateLimit(deps.RLS, getorderCfg, "getorders"),
	).Get("/{number}", deps.OrdersHandler.GetOrders)
}
