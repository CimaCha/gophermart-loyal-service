package model

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestGoodsInfo_Validate(t *testing.T) {
	tests := []struct {
		name    string
		goods   GoodsInfo
		wantErr bool
	}{
		{
			name: "valid percent reward",
			goods: GoodsInfo{
				Match:      "milk",
				Reward:     decimalPtr("10"),
				RewardType: RewardTypePercent,
			},
			wantErr: false,
		},
		{
			name: "valid points reward",
			goods: GoodsInfo{
				Match:      "bread",
				Reward:     decimalPtr("100"),
				RewardType: RewardTypePoints,
			},
			wantErr: false,
		},
		{
			name: "empty match",
			goods: GoodsInfo{
				Match:      "",
				Reward:     decimalPtr("10"),
				RewardType: RewardTypePercent,
			},
			wantErr: true,
		},
		{
			name: "nil reward",
			goods: GoodsInfo{
				Match:      "milk",
				Reward:     nil,
				RewardType: RewardTypePercent,
			},
			wantErr: true,
		},
		{
			name: "zero reward",
			goods: GoodsInfo{
				Match:      "milk",
				Reward:     decimalPtr("0"),
				RewardType: RewardTypePercent,
			},
			wantErr: true,
		},
		{
			name: "negative reward",
			goods: GoodsInfo{
				Match:      "milk",
				Reward:     decimalPtr("-10"),
				RewardType: RewardTypePercent,
			},
			wantErr: true,
		},
		{
			name: "unknown reward type",
			goods: GoodsInfo{
				Match:      "milk",
				Reward:     decimalPtr("10"),
				RewardType: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.goods.Validate()

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func decimalPtr(value string) *decimal.Decimal {
	d := decimal.RequireFromString(value)
	return &d
}
