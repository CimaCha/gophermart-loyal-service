package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// BalanceRepository работает с балансом пользователя в PostgreSQL.
type BalanceRepository struct {
	pool *pgxpool.Pool
}

// New создаёт репозиторий баланса.
func New(pool *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{pool: pool}
}

// GetBalance возвращает текущий баланс и сумму выводов пользователя.
// Если записи о балансе нет, возвращает нулевой Balance.
func (r *BalanceRepository) GetBalance(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
	const query = `
    SELECT user_uuid, current, withdrawn
    FROM balance
    WHERE user_uuid = $1
`

	var b model.Balance
	err := r.pool.QueryRow(ctx, query, userID).Scan(&b.UserID, &b.Current, &b.Withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Balance{UserID: userID}, nil
		}
		return model.Balance{}, fmt.Errorf("query balance: %w", err)
	}
	return b, nil
}

func (r *BalanceRepository) Withdraw(
	ctx context.Context,
	userID uuid.UUID,
	orderNum string,
	sum decimal.Decimal,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Извлечение текущего баланса
	var current decimal.Decimal
	err = tx.QueryRow(ctx,
		`SELECT current FROM balance WHERE user_uuid = $1 FOR UPDATE`,
		userID,
	).Scan(&current)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrInsufficientFunds
		}
		return fmt.Errorf("select balance for update: %w", err)
	}
	if current.LessThan(sum) {
		return model.ErrInsufficientFunds
	}
	// Списание
	_, err = tx.Exec(ctx,
		`UPDATE balance
			SET current = current - $1,
				withdrawn = withdrawn + $1
			WHERE user_uuid = $2`,
		sum, userID,
	)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	// Добавление операции в историю
	_, err = tx.Exec(ctx,
		`INSERT INTO transactions (user_uuid, order_num, sum) VALUES ($1, $2, $3)`,
		userID, orderNum, sum,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.ErrOrderAlreadyUsed
		}
		return fmt.Errorf("insert transaction: %w", err)
	}

	return tx.Commit(ctx)

}
