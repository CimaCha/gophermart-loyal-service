package worker

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/accrualclient"
	ordermodel "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWorker_ProcessOrder_Processed(t *testing.T) {
	ctx := context.Background()

	order := ordermodel.Order{
		OrderNum: "123",
		UserID:   uuid.New(),
		Status:   ordermodel.OrderStatusNew,
	}

	accrual := decimal.NewFromInt(500)

	accrualMock := NewMockAccrualClient(t)
	accrualMock.
		EXPECT().
		GetOrder(ctx, "123").
		Return(&accrualclient.ResultResponse{
			Order:   "123",
			Status:  accrualclient.StatusProcessed,
			Accrual: &accrual,
		}, nil)

	orderMock := NewMockOrderStore(t)
	orderMock.
		EXPECT().
		UpdateOrderResultTx(
			ctx,
			mock.Anything,
			"123",
			ordermodel.OrderStatusProcessed,
			&accrual,
		).
		Return(true, nil)

	balanceMock := NewMockBalanceAccruer(t)
	balanceMock.
		EXPECT().
		AccrueTx(
			ctx,
			mock.Anything,
			order.UserID,
			"123",
			accrual,
		).
		Return(nil)

	txMock := NewMockTransactor(t)
	txMock.
		EXPECT().
		BeginFunc(ctx, mock.Anything).
		Run(func(ctx context.Context, fn func(pgx.Tx) error) {
			require.NoError(t, fn(nil))
		}).
		Return(nil)

	w := New(
		orderMock,
		balanceMock,
		accrualMock,
		txMock,
		slog.Default(),
		time.Second,
		10,
	)

	err := w.processOrder(ctx, order)

	require.NoError(t, err)
}

func TestWorker_ProcessOrder_Invalid(t *testing.T) {
	ctx := context.Background()

	order := ordermodel.Order{
		OrderNum: "123",
		UserID:   uuid.New(),
	}

	accrualMock := NewMockAccrualClient(t)
	accrualMock.
		EXPECT().
		GetOrder(ctx, "123").
		Return(&accrualclient.ResultResponse{
			Order:  "123",
			Status: accrualclient.StatusInvalid,
		}, nil)

	orderMock := NewMockOrderStore(t)
	orderMock.
		EXPECT().
		UpdateOrderResultTx(
			ctx,
			mock.Anything,
			"123",
			ordermodel.OrderStatusInvalid,
			mock.MatchedBy(func(d *decimal.Decimal) bool {
				return d == nil
			}),
		).
		Return(true, nil)

	balanceMock := NewMockBalanceAccruer(t)

	txMock := NewMockTransactor(t)
	txMock.
		EXPECT().
		BeginFunc(ctx, mock.Anything).
		Run(func(ctx context.Context, fn func(pgx.Tx) error) {
			require.NoError(t, fn(nil))
		}).
		Return(nil)

	w := New(
		orderMock,
		balanceMock,
		accrualMock,
		txMock,
		slog.Default(),
		time.Second,
		10,
	)

	err := w.processOrder(ctx, order)

	require.NoError(t, err)
}

func TestWorker_ProcessOrder_Processing(t *testing.T) {
	ctx := context.Background()

	order := ordermodel.Order{
		OrderNum: "123",
		Status:   ordermodel.OrderStatusNew,
	}

	accrualMock := NewMockAccrualClient(t)
	accrualMock.
		EXPECT().
		GetOrder(ctx, "123").
		Return(&accrualclient.ResultResponse{
			Status: accrualclient.StatusRegistered,
		}, nil)

	orderMock := NewMockOrderStore(t)
	orderMock.
		EXPECT().
		UpdateStatus(
			ctx,
			"123",
			ordermodel.OrderStatusProcessing,
		).
		Return(nil)

	w := New(
		orderMock,
		NewMockBalanceAccruer(t),
		accrualMock,
		NewMockTransactor(t),
		slog.Default(),
		time.Second,
		10,
	)

	err := w.processOrder(ctx, order)

	require.NoError(t, err)
}

func TestWorker_ProcessOrder_NotRegistered(t *testing.T) {
	ctx := context.Background()

	order := ordermodel.Order{
		OrderNum: "123",
	}

	accrualMock := NewMockAccrualClient(t)
	accrualMock.
		EXPECT().
		GetOrder(ctx, "123").
		Return(&accrualclient.ResultResponse{
			Status: accrualclient.StatusNotRegistered,
		}, nil)

	w := New(
		NewMockOrderStore(t),
		NewMockBalanceAccruer(t),
		accrualMock,
		NewMockTransactor(t),
		slog.Default(),
		time.Second,
		10,
	)

	err := w.processOrder(ctx, order)

	require.NoError(t, err)
}
