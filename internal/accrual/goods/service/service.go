package service

import (
	"context"
	"fmt"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/model"
)

type GoodsRepository interface {
	RegisterGoods(ctx context.Context, goods model.GoodsInfo) error
}

type GoodsCacheAdder interface {
	Add(rule model.GoodsInfo)
}

type GoodsService struct {
	repo       GoodsRepository
	goodsCache GoodsCacheAdder
}

func New(goodsRepo GoodsRepository, goodsCache GoodsCacheAdder) *GoodsService {
	return &GoodsService{
		repo:       goodsRepo,
		goodsCache: goodsCache,
	}
}

func (s *GoodsService) RegisterGoods(ctx context.Context, goods model.GoodsInfo) error {
	if err := goods.Validate(); err != nil {
		return err
	}

	if err := s.repo.RegisterGoods(ctx, goods); err != nil {
		return fmt.Errorf("register goods: %w", err)
	}

	s.goodsCache.Add(goods)

	return nil
}
