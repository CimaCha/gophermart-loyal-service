package model

import (
	"errors"
	"github.com/CimaCha/gophermart-loyal-service/pkg/luhn"
	"time"
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
	OrderNum   string      `json:"order,omitempty"`
	Status     OrderStatus `json:"status,omitempty"`
	Accrual    *float64    `json:"accrual,omitempty"`
	Goods      []Good      `json:"goods,omitempty"`
	UploadedAt time.Time   `json:"uploaded_at"`
}

type Good struct {
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price,omitempty"`
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
