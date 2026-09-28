package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/core/deps"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r chi.Router, deps *deps.Dependencies) {
	r.Route("/api/goods", func(r chi.Router) {
		registerGoodsRoutes(r, deps)
	})

	r.Route("/api/orders", func(r chi.Router) {
		registerOrdersRoutes(r, deps)
	})
}
