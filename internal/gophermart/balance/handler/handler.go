package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	logger  *slog.Logger
	service BalanceService
}

type BalanceService interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (model.Balance, error)
}

func New(logger *slog.Logger, balanceService BalanceService) *Handler {
	return &Handler{
		logger:  logger,
		service: balanceService,
	}
}

// RegisterProtectedAPI регистрирует роуты, требующие аутентификации.
func (h *Handler) RegisterProtectedAPI(r chi.Router) {
	r.Get("/api/user/balance", h.GetBalance)
	// r.Post("/api/user/balance/withdraw", h.Withdraw)
	// r.Get("/api/user/withdrawals", h.GetWithdrawals)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := ctxkeys.GetUserID(r.Context())
	if err != nil {
		h.logger.Error("get user id from context failed", "err", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	balance, err := h.service.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Error("get balance failed", "err", err, "user_id", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		h.logger.Error("encode balance failed", "err", err)
	}
}
