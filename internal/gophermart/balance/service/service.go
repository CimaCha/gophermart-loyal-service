package service

import (
	"context"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
	"github.com/CimaCha/gophermart-loyal-service/pkg/luhn"
	"github.com/google/uuid"

	"github.com/shopspring/decimal"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (model.Balance, error)
	Withdraw(ctx context.Context, userID uuid.UUID, orderNum string, sum decimal.Decimal) error
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

func (s *BalanceService) Withdraw(
	ctx context.Context,
	userID uuid.UUID,
	orderNum string,
	sum decimal.Decimal,
) error {
	if !luhn.Validate(orderNum) {
		return model.ErrInvalidOrderNumber
	}
	if sum.IsZero() || sum.IsNegative() {
		return model.ErrInvalidOrderNumber
	}
	return s.repo.Withdraw(ctx, userID, orderNum, sum)
}
