package model

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrder_Validate(t *testing.T) {
	tests := []struct {
		name     string
		orderNum int64
		wantErr  error
	}{
		{
			name:     "valid order",
			orderNum: 79927398713,
		},
		{
			name:     "negative order",
			orderNum: -1,
			wantErr:  errors.New("order id must be positive"),
		},
		{
			name:     "invalid checksum",
			orderNum: 79927398714,
			wantErr:  ErrInvalidOrderNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := NewOrder(tt.orderNum, uuid.New())

			got, err := order.Validate()

			if tt.wantErr != nil {
				require.Error(t, err)

				if errors.Is(tt.wantErr, ErrInvalidOrderNumber) {
					assert.ErrorIs(t, err, ErrInvalidOrderNumber)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
				}

				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, order, got)
		})
	}
}
