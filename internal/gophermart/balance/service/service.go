package service

import (
	"context"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
	"github.com/google/uuid"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (model.Balance, error)
}

type BalanceService struct {
	repo BalanceRepository
}

func New(balanceRepo BalanceRepository) *BalanceService {
	return &BalanceService{
		repo: balanceRepo,
	}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID uuid.UUID) (model.Balance, error) {
	return s.repo.GetBalance(ctx, userID)
}
