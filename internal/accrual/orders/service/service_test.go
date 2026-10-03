package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	orderrepo "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/repository"
	mocks "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUploadOrder(t *testing.T) {
	dbErr := errors.New("database unavailable")
	tests := []struct {
		name    string
		number  string
		repoErr error
		wantErr error
	}{
		{name: "success", number: "12345678903"},
		{name: "empty number", wantErr: ErrInvalidOrderNum},
		{name: "non digits", number: "123abc", wantErr: ErrInvalidOrderNum},
		{name: "invalid checksum", number: "12345678904", wantErr: ErrInvalidOrderNum},
		{name: "all zeros", number: "0000", wantErr: ErrInvalidOrderNum},
		{name: "duplicate", number: "12345678903", repoErr: orderrepo.ErrOrderAlreadyProcessing, wantErr: ErrOrderAlreadyProcessing},
		{name: "wrapped duplicate", number: "12345678903", repoErr: fmt.Errorf("insert: %w", orderrepo.ErrOrderAlreadyProcessing), wantErr: ErrOrderAlreadyProcessing},
		{name: "repository failure", number: "12345678903", repoErr: dbErr, wantErr: dbErr},
		{name: "canceled", number: "12345678903", repoErr: context.Canceled, wantErr: context.Canceled},
		{name: "deadline", number: "12345678903", repoErr: context.DeadlineExceeded, wantErr: context.DeadlineExceeded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockOrderRepository(t)
			notifier := mocks.NewMockOrderNotifier(t)
			svc := New(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
			// Inject the dependency for unit tests; New currently leaves it nil.
			svc.notifier = notifier
			ctx := t.Context()
			accrual := 999.0
			input := model.Order{
				OrderNum: tt.number, Goods: []model.Good{{Description: "Чайник Bork", Price: 7000}},
				Status: model.OrderStatusProcessed, Accrual: &accrual, UploadedAt: time.Unix(1, 0),
			}
			started := time.Now()
			var saved model.Order
			if tt.wantErr != ErrInvalidOrderNum {
				repo.EXPECT().CreateOrder(ctx, mock.Anything).Run(func(_ context.Context, order *model.Order) {
					saved = *order
					assert.Equal(t, input.OrderNum, order.OrderNum)
					assert.Equal(t, input.Goods, order.Goods)
					assert.Equal(t, model.OrderStatusRegistered, order.Status)
					assert.Nil(t, order.Accrual)
					assert.False(t, order.UploadedAt.Before(started))
					assert.False(t, order.UploadedAt.After(time.Now()))
				}).Return(tt.repoErr).Once()
				if tt.repoErr == nil {
					notifier.EXPECT().Notify(mock.Anything).Run(func(order model.Order) {
						// A notification must follow successful persistence and carry the saved order.
						require.NotEmpty(t, saved.OrderNum)
						assert.Equal(t, saved, order)
					}).Once()
				}
			}
			err := svc.UploadOrder(ctx, input)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			if tt.wantErr == ErrInvalidOrderNum {
				repo.AssertNotCalled(t, "CreateOrder", mock.Anything, mock.Anything)
			}
			if tt.wantErr != nil {
				notifier.AssertNotCalled(t, "Notify", mock.Anything)
			}
		})
	}
}
