package model

import (
	"github.com/shopspring/decimal"
)

type RewardType string

const (
	RewardTypePercent RewardType = "%"  // процент от стоимости товара
	RewardTypePoints  RewardType = "pt" // фиксированное количество баллов
)

type GoodsInfo struct {
	Match      string           `json:"match"`
	Reward     *decimal.Decimal `json:"reward,omitempty"`
	RewardType RewardType       `json:"reward_type"`
}

func (g GoodsInfo) Validate() error {
	if g.Match == "" {
		return ErrInvalidGoods
	}
	if g.Reward == nil || !g.Reward.IsPositive() {
		return ErrInvalidGoods
	}
	switch g.RewardType {
	case RewardTypePercent, RewardTypePoints:
		return nil
	default:
		return ErrInvalidGoods
	}
}
