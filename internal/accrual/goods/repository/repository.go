package repository

import (
	"context"
	"errors"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/model"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GoodsRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *GoodsRepository {
	return &GoodsRepository{
		pool: pool,
	}
}

func (r *GoodsRepository) RegisterGoods(ctx context.Context, goods model.GoodsInfo) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rewards (match, reward_value, reward_type) VALUES ($1, $2, $3)`,
		goods.Match, goods.Reward, goods.RewardType,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.ErrMatchAlreadyExists
		}
	}
	return err
}
