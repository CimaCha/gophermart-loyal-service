package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/CimaCha/gophermart-loyal-service/internal/order/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

var (
	ErrOrderAlreadyCreatedByAnotherUser = errors.New("order already created by another user")
	ErrOrderAlreadyProcessing           = errors.New("order has already been uploaded by this user")
)

func New(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		pool: pool,
	}
}

func (r *OrderRepository) GetOrders(ctx context.Context, uid uuid.UUID) ([]model.Order, error) {

	query := `
		SELECT order_num, order_status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at ASC
	`

	orders := make([]model.Order, 0, 32)

	rows, err := r.pool.Query(ctx, query, uid)
	if err != nil {
		return nil, fmt.Errorf("pool query: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var o model.Order

		if err := rows.Scan(
			&o.OrderNum,
			&o.Status,
			&o.Accrual,
			&o.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}

		orders = append(orders, o)

	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return orders, nil

}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *model.Order) error {

	query := `
		INSERT INTO orders 
			(order_num, user_id, order_status, accrual, uploaded_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (order_num) DO UPDATE SET order_num = EXCLUDED.order_num
			RETURNING user_id, (xmax != 0) AS is_conflict`

	var (
		existingUserID uuid.UUID
		isConflict     bool
	)

	err := r.pool.QueryRow(
		ctx,
		query,
		order.OrderNum,
		order.UserID,
		order.Status,
		order.Accrual,
		order.UploadedAt,
	).Scan(&existingUserID, &isConflict)

	if err != nil {
		return fmt.Errorf("db insert: %w", err)
	}

	if isConflict {
		if existingUserID != order.UserID {
			return ErrOrderAlreadyCreatedByAnotherUser
		}
		return ErrOrderAlreadyProcessing
	}

	return nil

}
