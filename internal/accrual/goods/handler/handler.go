package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/model"
)

type GoodsService interface {
	RegisterGoods(ctx context.Context, goods model.GoodsInfo) error
}

type Handler struct {
	logger  *slog.Logger
	service GoodsService
}

func New(logger *slog.Logger, goodsService GoodsService) *Handler {
	return &Handler{
		logger:  logger,
		service: goodsService,
	}
}

func (h *Handler) RegisterGoods(w http.ResponseWriter, r *http.Request) {

	var req model.GoodsInfo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err := h.service.RegisterGoods(r.Context(), req)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, model.ErrInvalidGoods):
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	case errors.Is(err, model.ErrMatchAlreadyExists):
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
	default:
		h.logger.Error("register goods failed", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

}
