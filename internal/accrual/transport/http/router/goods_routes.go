package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/core/deps"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Метод для регистрации маршрутов сервиса user

func registerGoodsRoutes(r chi.Router, deps *deps.Dependencies) {
	r.With(
		chimiddleware.AllowContentType("application/json"),
	).Post("/", deps.GoodsHandler.RegisterGoods)
}
