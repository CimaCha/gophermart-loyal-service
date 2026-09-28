package model

import "github.com/shopspring/decimal"

type RewardType string

const (
	RewardTypePercent RewardType = "%"  // RFC 9110, 15.2.1
	RewardTypePoints  RewardType = "pt" // RFC 9110, 15.2.2
)

type GoodsInfo struct {
	Match      string
	Reward     *decimal.Decimal
	RewardType string
}
