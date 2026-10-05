package service

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/order/model"
	"github.com/shopspring/decimal"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUploadOrder_Success(t *testing.T) {
	repo := NewMockOrderRepository(t)
	notifier := NewMockOrderNotifier(t)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := New(repo, log, notifier)
	svc.notifier = notifier

	order := model.Order{
		OrderNum: "79927398713",
		Goods: []model.Good{
			{
				Description: "Tea",
				Price:       decimal.RequireFromString("100"),
			},
		},
	}

	repo.
		EXPECT().CreateOrder(context.Background(), mock.AnythingOfType("*model.Order")).
		Return(nil).
		Once()

	notifier.
		EXPECT().Notify(mock.AnythingOfType("model.Order")).
		Once()

	err := svc.UploadOrder(context.Background(), order)

	require.NoError(t, err)

	repo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

func TestUploadOrder_InvalidOrderNumber(t *testing.T) {
	repo := NewMockOrderRepository(t)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := New(repo, log, nil)

	order := model.Order{
		OrderNum: "12345",
	}

	err := svc.UploadOrder(context.Background(), order)

	require.ErrorIs(t, err, ErrInvalidOrderNum)
}
