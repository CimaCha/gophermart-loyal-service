package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/ctxkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Handler struct {
	logger  *slog.Logger
	service BalanceService
}

type BalanceService interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (model.Balance, error)
	Withdraw(ctx context.Context, userID uuid.UUID, orderNum string, sum decimal.Decimal) error
}

func New(logger *slog.Logger, balanceService BalanceService) *Handler {
	return &Handler{
		logger:  logger,
		service: balanceService,
	}
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

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, err := ctxkeys.GetUserID(r.Context())
	if err != nil {
		h.logger.Error("get user id from context failed", "err", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var req model.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = h.service.Withdraw(r.Context(), userID, req.Order, req.Sum)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, model.ErrInvalidOrderNumber):
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
	case errors.Is(err, model.ErrInsufficientFunds):
		http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
	case errors.Is(err, model.ErrOrderAlreadyUsed):
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
	default:
		h.logger.Error("withdraw failed", "err", err, "user_id", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
