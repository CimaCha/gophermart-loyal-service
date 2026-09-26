package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/core/deps"
	"github.com/go-chi/chi/v5"
)

func registerBalanceRoutes(r chi.Router, deps *deps.Dependencies) {
	r.Get("/balance", deps.BalanceHandler.GetBalance)
	r.Post("/balance/withdraw", deps.BalanceHandler.Withdraw)
}
