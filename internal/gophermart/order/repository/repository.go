package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
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

func (r *OrderRepository) UpdateOrderResultTx(
	ctx context.Context,
	tx pgx.Tx, orderNum string,
	status model.OrderStatus,
	accrual *decimal.Decimal,
) (bool, error) {

	query := `
		UPDATE orders
		SET order_status = $1, accrual = $2
		WHERE order_num = $3
		AND order_status NOT IN ('PROCESSED', 'INVALID')
	`

	result, err := tx.Exec(
		ctx,
		query,
		status.String(),
		accrual,
		orderNum,
	)

	if err != nil {
		return false, fmt.Errorf("tx exec: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (r *OrderRepository) GetPendingOrders(ctx context.Context) ([]model.Order, error) {

	query := `
		SELECT order_num, user_id, order_status, accrual, uploaded_at
		FROM orders
		WHERE order_status IN ('NEW', 'PROCESSING') 
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("pool query: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0, 32)

	for rows.Next() {
		var o model.Order

		if err := rows.Scan(
			&o.OrderNum,
			&o.UserID,
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

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderNum string, status model.OrderStatus) error {

	query := `
		UPDATE orders SET order_status = $1
		WHERE order_num = $2
	`

	result, err := r.pool.Exec(ctx, query, status.String(), orderNum)
	if err != nil {
		return fmt.Errorf("pool exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("there is no order with that num")
	}

	return nil

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
