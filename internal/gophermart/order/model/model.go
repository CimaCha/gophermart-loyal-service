package model

import (
	"errors"
	"strconv"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/pkg/luhn"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidOrderNumber = errors.New("invalid order number checksum")
)

type Order struct {
	OrderNum   int64
	UserID     uuid.UUID
	Status     string
	Accrual    *decimal.Decimal
	UploadedAt time.Time
}

func NewOrder(orderNum int64, uid uuid.UUID) *Order {
	return &Order{
		OrderNum:   orderNum,
		UserID:     uid,
		Status:     "NEW",
		Accrual:    nil,
		UploadedAt: time.Now(),
	}
}
func (o *Order) Validate() (*Order, error) {
	if o.OrderNum < 0 {
		return nil, errors.New("order id must be positive")
	}
	if !luhn.Validate(strconv.FormatInt(o.OrderNum, 10)) {
		return nil, ErrInvalidOrderNumber
	}
	return o, nil
}
