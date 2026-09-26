package repository

import (
	"context"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		pool: pool,
	}
}

func (r *OrderRepository) GetOrders(ctx context.Context, uid uuid.UUID) ([]model.Order, error) {
	//TODO
	return nil, nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	//TODO
	return nil

}
