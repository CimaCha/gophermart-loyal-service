package service

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/model"
	mock "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/service/mock"
)

func TestService_RegisterGoods(t *testing.T) {
	storageErr := errors.New("storage failed")
	tests := []struct {
		name         string
		match        string
		reward       string
		rewardType   model.RewardType
		storedReward string
		repoErr      error
		wantErr      error
	}{
		{name: "percent", match: "Bork", reward: "10", rewardType: model.RewardTypePercent, storedReward: "10"},
		{name: "points", match: "Bork", reward: "500", rewardType: model.RewardTypePoints, storedReward: "500"},
		{name: "round to database scale", match: "Bork", reward: "10.005", rewardType: model.RewardTypePoints, storedReward: "10.01"},
		{name: "empty match", reward: "10", rewardType: model.RewardTypePercent, wantErr: model.ErrInvalidGoods},
		{name: "unknown type", match: "Bork", reward: "10", rewardType: "unknown", wantErr: model.ErrInvalidGoods},
		{name: "rounds to zero", match: "Bork", reward: "0.004", rewardType: model.RewardTypePoints, wantErr: model.ErrInvalidGoods},
		{name: "maximum reward", match: "Bork", reward: "9999999999999.99", rewardType: model.RewardTypePoints, storedReward: "9999999999999.99"},
		{name: "overflow after rounding", match: "Bork", reward: "9999999999999.995", rewardType: model.RewardTypePoints, wantErr: model.ErrInvalidGoods},
		{name: "duplicate", match: "Bork", reward: "10", rewardType: model.RewardTypePercent, storedReward: "10", repoErr: model.ErrMatchAlreadyExists, wantErr: model.ErrMatchAlreadyExists},
		{name: "storage error", match: "Bork", reward: "10", rewardType: model.RewardTypePercent, storedReward: "10", repoErr: storageErr, wantErr: storageErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := mock.NewMockGoodsRepository(t)
			cache := mock.NewMockGoodsCacheAdder(t)
			reward := decimal.RequireFromString(tt.reward)
			goods := model.GoodsInfo{Match: tt.match, Reward: &reward, RewardType: tt.rewardType}
			if tt.storedReward != "" {
				storedReward := decimal.RequireFromString(tt.storedReward).Round(2)
				stored := model.GoodsInfo{Match: tt.match, Reward: &storedReward, RewardType: tt.rewardType}
				repo.EXPECT().RegisterGoods(ctx, stored).Return(tt.repoErr).Once()
				if tt.repoErr == nil {
					cache.EXPECT().Add(stored).Once()
				}
			}
			err := New(repo, cache).RegisterGoods(ctx, goods)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, decimal.RequireFromString(tt.reward), reward, "input reward must remain unchanged")
		})
	}
}
