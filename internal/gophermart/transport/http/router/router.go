package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/core/deps"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/http/middleware"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r chi.Router, deps *deps.Dependencies) {
	r.Route("/api/user", func(r chi.Router) {
		registerUserRoutes(r, deps)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(deps.Tv))

			r.Route("/orders", func(r chi.Router) {
				registerOrderRoutes(r, deps)
			})

		})
	})
}
