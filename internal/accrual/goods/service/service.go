package service

import (
	"context"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/model"
)

type GoodsRepository interface {
	RegisterGoods(ctx context.Context, goods model.GoodsInfo) error
}

type GoodsService struct {
	repo GoodsRepository
}

func New(goodsRepo GoodsRepository) *GoodsService {
	return &GoodsService{
		repo: goodsRepo,
	}
}

func (s *GoodsService) RegisterGoods(ctx context.Context, goods model.GoodsInfo) error {
	if err := goods.Validate(); err != nil {
		return err
	}
	return s.repo.RegisterGoods(ctx, goods)
}
