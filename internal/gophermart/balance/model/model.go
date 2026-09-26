package model

import (
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
