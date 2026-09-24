package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BalanceRepository работает с балансом пользователя в PostgreSQL.
type BalanceRepository struct {
	pool *pgxpool.Pool
}

// New создаёт репозиторий баланса.
func New(pool *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{
		pool: pool,
	}
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
