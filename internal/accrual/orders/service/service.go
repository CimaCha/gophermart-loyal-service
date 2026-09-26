package service

import (
	"context"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	"github.com/google/uuid"
	"log/slog"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrders(ctx context.Context, uid uuid.UUID) ([]model.Order, error)
}

type OrderService struct {
	repo OrderRepository
	l    *slog.Logger
}

func New(
	repo OrderRepository,
	log *slog.Logger,
) *OrderService {
	return &OrderService{
		repo: repo,
		l:    log,
	}
}

func (s *OrderService) GetOrders(ctx context.Context, uid uuid.UUID) ([]model.Order, error) {
	//TODO
	return nil, nil
}

func (s *OrderService) UploadOrder(ctx context.Context, orderNumStr string, uid uuid.UUID) error {
	//TODO
	return nil

}
