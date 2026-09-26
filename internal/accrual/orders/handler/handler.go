package handler

import (
	"context"
	model2 "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type OrderService interface {
	UploadOrder(ctx context.Context, orderNumStr string, uid uuid.UUID) error
	GetOrders(ctx context.Context, uid uuid.UUID) ([]model2.Order, error)
}

type Handler struct {
	l   *slog.Logger
	svc OrderService
}

func New(
	log *slog.Logger,
	service OrderService,
) *Handler {
	return &Handler{
		l:   log,
		svc: service,
	}
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	//TODO
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	//TODO
}
