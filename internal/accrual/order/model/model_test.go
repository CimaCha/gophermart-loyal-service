package model

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrder(t *testing.T) {
	price := decimal.RequireFromString("199.99")

	goods := []Good{
		{
			Description: "Coffee",
			Price:       price,
		},
	}

	order := NewOrder("79927398713", goods)

	require.NotNil(t, order)

	assert.Equal(t, "79927398713", order.OrderNum)
	assert.Equal(t, OrderStatusRegistered, order.Status)
	assert.Nil(t, order.Accrual)
	assert.Equal(t, goods, order.Goods)
	assert.False(t, order.UploadedAt.IsZero())
}

func TestOrder_Validate_Success(t *testing.T) {
	order := &Order{
		OrderNum: "79927398713",
	}

	got, err := order.Validate()

	require.NoError(t, err)
	assert.Equal(t, order, got)
}

func TestOrder_Validate_InvalidOrderNumber(t *testing.T) {
	order := &Order{
		OrderNum: "1234567890",
	}

	got, err := order.Validate()

	require.ErrorIs(t, err, ErrInvalidOrderNumber)
	assert.Nil(t, got)
}

func TestOrderStatus_String(t *testing.T) {
	assert.Equal(t, "REGISTERED", OrderStatusRegistered.String())
	assert.Equal(t, "PROCESSING", OrderStatusProcessing.String())
	assert.Equal(t, "PROCESSED", OrderStatusProcessed.String())
	assert.Equal(t, "INVALID", OrderStatusInvalid.String())
}
