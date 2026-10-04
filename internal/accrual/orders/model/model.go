package model

import (
	"errors"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/pkg/luhn"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidOrderNumber = errors.New("invalid order number checksum")
)

type OrderStatus string

const (
	OrderStatusRegistered OrderStatus = "REGISTERED"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
	OrderStatusInvalid    OrderStatus = "INVALID"
)

func (ost OrderStatus) String() string {
	return string(ost)
}

type Order struct {
	OrderNum   string           `json:"order"`
	Status     OrderStatus      `json:"status,omitempty"`
	Accrual    *decimal.Decimal `json:"accrual,omitempty"`
	Goods      []Good           `json:"goods,omitempty"`
	UploadedAt time.Time        `json:"uploaded_at"`
}

type Good struct {
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
}

func NewOrder(orderNum string, goods []Good) *Order {
	return &Order{
		OrderNum:   orderNum,
		Status:     OrderStatusRegistered,
		Accrual:    nil,
		Goods:      goods,
		UploadedAt: time.Now(),
	}
}
func (o *Order) Validate() (*Order, error) {
	if !luhn.Validate(o.OrderNum) {
		return nil, ErrInvalidOrderNumber
	}
	return o, nil
}
