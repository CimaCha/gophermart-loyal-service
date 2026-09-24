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
