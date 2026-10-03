package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	ordersvc "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/service"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type OrderService interface {
	UploadOrder(ctx context.Context, order model.Order) error
	GetOrder(ctx context.Context, uid uuid.UUID) (model.Order, error)
}

type Handler struct {
	logger  *slog.Logger
	service OrderService
}

func New(
	log *slog.Logger,
	service OrderService,
) *Handler {
	return &Handler{
		logger:  log,
		service: service,
	}
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	//TODO
}
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {

	var req model.Order
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := h.service.UploadOrder(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			h.logger.Debug("create order canceled by client")
			w.WriteHeader(499)
			return
		case errors.Is(err, ordersvc.ErrInvalidOrderNum):
			h.logger.Info(
				"invalid order number",
				"order", req.OrderNum,
				"err", err,
			)
			http.Error(
				w,
				http.StatusText(http.StatusBadRequest),
				http.StatusBadRequest,
			)
			return
		case errors.Is(err, ordersvc.ErrOrderAlreadyProcessing):
			h.logger.Info(
				"order already uploaded by user",
				"err", err,
			)
			http.Error(
				w,
				http.StatusText(http.StatusConflict),
				http.StatusConflict,
			)
			return
		default:
			h.logger.Error(
				"unexpected error during order saving",
				"err", err,
			)

			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
