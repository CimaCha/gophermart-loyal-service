// Package router отвечает за централизованное конфигурирование маршрутов (эндпоинтов) API
// и связывание путей запросов с соответствующими HTTP-обработчиками и middleware.
package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/core/deps"
	"github.com/go-chi/chi/v5"
)

// SetupRoutes регистрирует все основные группы маршрутов приложения в корневом роутере chi.Router.
// Разделяет эндпоинты на логические блоки (/api/goods и /api/orders) и пробрасывает в них
// инициализированный контейнер зависимостей deps.
func SetupRoutes(r chi.Router, deps *deps.Dependencies) {
	r.Route("/api/goods", func(r chi.Router) {
		registerGoodsRoutes(r, deps)
	})

	r.Route("/api/orders", func(r chi.Router) {
		registerOrdersRoutes(r, deps)
	})
}
