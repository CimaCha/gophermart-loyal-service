package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/model"
	orderrepo "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/repository"
	"github.com/google/uuid"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrders(ctx context.Context, uid uuid.UUID) ([]model.Order, error)
}

type OrderService struct {
	repo OrderRepository
	l    *slog.Logger
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
) *OrderService {
	return &OrderService{
		repo: repo,
		l:    log,
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

	orderNum, err := strconv.ParseInt(orderNumStr, 10, 64)
	if err != nil {
		return ErrInvalidOrderNum
	}

	order, err := model.NewOrder(orderNum, uid).Validate()
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

	return nil

}
