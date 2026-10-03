package model

import (
	"errors"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/pkg/luhn"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidOrderNumber = errors.New("invalid order number checksum")
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
	OrderStatusInvalid    OrderStatus = "INVALID"
)

func (ost OrderStatus) String() string {
	return string(ost)
}

type Order struct {
	OrderNum   string
	UserID     uuid.UUID
	Status     OrderStatus
	Accrual    *decimal.Decimal
	UploadedAt time.Time
}

func NewOrder(orderNum string, uid uuid.UUID) *Order {
	return &Order{
		OrderNum:   orderNum,
		UserID:     uid,
		Status:     OrderStatusNew,
		Accrual:    nil,
		UploadedAt: time.Now(),
	}
}
func (o *Order) Validate() (*Order, error) {
	if !luhn.Validate(o.OrderNum) {
		return nil, ErrInvalidOrderNumber
	}
	return o, nil
}
