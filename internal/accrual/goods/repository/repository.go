package repository

import (
	"context"
	model2 "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/model"
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

func (r *GoodsRepository) RegisterGoods(ctx context.Context, goods model2.GoodsInfo) error {
	// TODO
	return nil
}
