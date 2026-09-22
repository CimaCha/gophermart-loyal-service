package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/core/deps"
	"github.com/go-chi/chi/v5"
)

// Метод для регистрации маршрутов для сервиса order

func registerOrderRoutes(r chi.Router, deps *deps.Dependencies) {
	r.Post("/", deps.OrderHandler.CreateOrder)

	r.Get("/", deps.OrderHandler.GetOrders)
}
