package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/model"
	orderrepo "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/repository"
	"github.com/google/uuid"
)

type OrderNotifier interface {
	Notify(order model.Order)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrders(ctx context.Context, uid uuid.UUID) ([]model.Order, error)
}

type OrderService struct {
	repo     OrderRepository
	l        *slog.Logger
	notifier OrderNotifier
}

var (
	ErrInvalidOrderNum                  = errors.New("invalid order number")
	ErrOrderAlreadyProcessing           = errors.New("order has already been uploaded by this user")
	ErrOrderAlreadyCreatedByAnotherUser = errors.New("order already created by another user")
	ErrOrdersNotFound                   = errors.New("orders not found")
)

func New(
	repo OrderRepository,
	log *slog.Logger,
	notifier OrderNotifier,
) *OrderService {
	return &OrderService{
		repo:     repo,
		l:        log,
		notifier: notifier,
	}
}

func (s *OrderService) GetOrders(ctx context.Context, uid uuid.UUID) ([]model.Order, error) {

	result, err := s.repo.GetOrders(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("repository get orders: %w", err)
	}

	return result, nil
}

func (s *OrderService) UploadOrder(ctx context.Context, orderNumStr string, uid uuid.UUID) error {

	order, err := model.NewOrder(orderNumStr, uid).Validate()
	if err != nil {
		return ErrInvalidOrderNum
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		switch {
		case errors.Is(err, orderrepo.ErrOrderAlreadyCreatedByAnotherUser):
			return ErrOrderAlreadyCreatedByAnotherUser
		case errors.Is(err, orderrepo.ErrOrderAlreadyProcessing):
			return ErrOrderAlreadyProcessing
		default:
			return fmt.Errorf("repository create order: %w", err)
		}
	}

	s.notifier.Notify(*order)

	return nil

}
