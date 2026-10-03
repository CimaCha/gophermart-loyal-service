package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrOrderAlreadyProcessing = errors.New("order has already been uploaded")
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		pool: pool,
	}
}

func (r *OrderRepository) GetOrder(ctx context.Context, orderID string) (model.Order, error) {
	//TODO
	return model.Order{}, nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	query := `
		INSERT INTO orders 
			(order_num, order_status, uploaded_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (order_num) DO UPDATE SET order_num = EXCLUDED.order_num
			RETURNING (xmax != 0) AS is_conflict`

	var isConflict bool

	err = tx.QueryRow(
		ctx,
		query,
		order.OrderNum,
		order.Status,
		order.UploadedAt,
	).Scan(&isConflict)
	if err != nil {
		return fmt.Errorf("db insert: %w", err)
	}

	if isConflict {
		return ErrOrderAlreadyProcessing
	}

	var rows [][]interface{}
	for _, good := range order.Goods {
		rows = append(rows, []interface{}{good.Description, good.Price, order.OrderNum})
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"goods"},
		[]string{"description", "price", "order_num"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("db copy from: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, order *model.Order) error {
	//TODO
	return nil
}
