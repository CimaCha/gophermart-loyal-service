package service

import (
	"context"
	model2 "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/model"
)

type GoodsRepository interface {
	RegisterGoods(ctx context.Context, goods model2.GoodsInfo) error
}

type GoodsService struct {
	repo GoodsRepository
}

func New(userRepo GoodsRepository) *GoodsService {
	return &GoodsService{
		repo: userRepo,
	}
}

func (s *GoodsService) RegisterGoods(ctx context.Context, goods model2.GoodsInfo) error {
	//TODO
	return nil
}
