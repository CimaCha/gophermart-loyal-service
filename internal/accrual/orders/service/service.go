package service

import (
	"context"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	"github.com/google/uuid"
	"log/slog"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrder(ctx context.Context, uid uuid.UUID) (model.Order, error)
	UpdateOrder(ctx context.Context, order *model.Order) error
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

func (s *OrderService) GetOrder(ctx context.Context, uid uuid.UUID) (model.Order, error) {
	//TODO
	return model.Order{}, nil
}

func (s *OrderService) UploadOrder(ctx context.Context, order model.Order) error {
	//TODO
	return nil
}
