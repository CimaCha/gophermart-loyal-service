package model

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidLuhn(t *testing.T) {
	tests := []struct {
		name      string
		orderNum  int64
		wantValid bool
	}{
		{
			name:      "valid order",
			orderNum:  79927398713,
			wantValid: true,
		},
		{
			name:      "another valid order",
			orderNum:  4242424242424242,
			wantValid: true,
		},
		{
			name:      "invalid checksum",
			orderNum:  79927398714,
			wantValid: false,
		},
		{
			name:      "single digit",
			orderNum:  1,
			wantValid: false,
		},
		{
			name:      "zero",
			orderNum:  0,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantValid, validLuhn(tt.orderNum))
		})
	}
}

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
