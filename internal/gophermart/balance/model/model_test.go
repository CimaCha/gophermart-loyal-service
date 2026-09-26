package model

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestWithdrawRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     WithdrawRequest
		wantErr error
	}{
		{"valid", WithdrawRequest{Order: "12345678903", Sum: mustDecimal("100")}, nil},
		{"empty order", WithdrawRequest{Order: "", Sum: mustDecimal("100")}, ErrInvalidOrderNumber},
		{"bad luhn", WithdrawRequest{Order: "12345678904", Sum: mustDecimal("100")}, ErrInvalidOrderNumber},
		{"zero sum", WithdrawRequest{Order: "12345678903", Sum: decimal.Zero}, ErrInvalidOrderNumber},
		{"negative sum", WithdrawRequest{Order: "12345678903", Sum: mustDecimal("-1")}, ErrInvalidOrderNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
