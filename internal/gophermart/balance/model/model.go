package model

import (
	"time"

	"github.com/CimaCha/gophermart-loyal-service/pkg/luhn"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Balance struct {
	UserID    uuid.UUID       `json:"-"`
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

// Тело запроса POST /api/user/balance/withdraw.
type WithdrawRequest struct {
	Order string          `json:"order"`
	Sum   decimal.Decimal `json:"sum"`
}

func (r WithdrawRequest) Validate() error {
	if !luhn.Validate(r.Order) {
		return ErrInvalidOrderNumber
	}
	if !r.Sum.IsPositive() {
		return ErrInvalidOrderNumber
	}
	return nil
}

type Withdrawal struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}
