package handler

import (
	"context"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
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
	//TODO
}
