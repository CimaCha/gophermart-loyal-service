package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	orderrepo "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/repository"
	"github.com/google/uuid"
	"log/slog"
)

type OrderNotifier interface {
	Notify(order model.Order)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrder(ctx context.Context, orderId string) (model.Order, error)
	UpdateOrder(ctx context.Context, order *model.Order) error
}

type OrderService struct {
	repo     OrderRepository
	logger   *slog.Logger
	notifier OrderNotifier
}

var (
	ErrInvalidOrderNum        = errors.New("invalid order number")
	ErrOrderAlreadyProcessing = errors.New("order has already been uploaded by user")
	ErrOrdersNotFound         = errors.New("orders not found")
)

func New(
	repo OrderRepository,
	log *slog.Logger,
) *OrderService {
	return &OrderService{
		repo:   repo,
		logger: log,
	}
}

func (s *OrderService) GetOrder(ctx context.Context, uid uuid.UUID) (model.Order, error) {
	//TODO
	return model.Order{}, nil
}

func (s *OrderService) UploadOrder(ctx context.Context, order model.Order) error {

	newOrder, err := model.NewOrder(order.OrderNum, order.Goods).Validate()
	if err != nil {
		return ErrInvalidOrderNum
	}

	if err = s.repo.CreateOrder(ctx, newOrder); err != nil {
		switch {
		case errors.Is(err, orderrepo.ErrOrderAlreadyProcessing):
			return ErrOrderAlreadyProcessing
		default:
			return fmt.Errorf("repository create order: %w", err)
		}
	}

	s.notifier.Notify(*newOrder)

	return nil
}
