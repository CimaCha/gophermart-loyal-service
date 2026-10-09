// Package model содержит структуры данных и доменные модели для работы с товарами,
// правилами начисления вознаграждений и связанными с ними ошибками.
package model

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	// ErrInvalidGoods возвращается, когда структура с информацией о товаре не прошла валидацию.
	ErrInvalidGoods = errors.New("invalid goods info")
	// ErrMatchAlreadyExists возвращается, если правило для данного шаблона сопоставления уже зарегистрировано.
	ErrMatchAlreadyExists = errors.New("match already exists")
)

// RewardType определяет строковый тип для перечисления поддерживаемых видов вознаграждений.
type RewardType string

const (
	// RewardTypePercent указывает, что вознаграждение рассчитывается как процент от стоимости товара.
	RewardTypePercent RewardType = "%"
	// RewardTypePoints указывает, что за товар начисляется фиксированное количество баллов.
	RewardTypePoints RewardType = "pt"
)

// GoodsInfo описывает правило расчета вознаграждения для конкретного типа или группы товаров.
type GoodsInfo struct {
	// Match задает шаблон или ключевое слово для сопоставления с позициями в заказе.
	Match string `json:"match"`
	// Reward определяет размер вознаграждения (дробное число произвольной точности).
	Reward *decimal.Decimal `json:"reward,omitempty"`
	// RewardType задает логику применения вознаграждения (процент или фиксированные баллы).
	RewardType RewardType `json:"reward_type"`
}

// Validate выполняет бизнес-валидацию структуры GoodsInfo.
// Возвращает ErrInvalidGoods, если не заполнено поле Match, отсутствует или является
// неположительным размер Reward, либо передан неподдерживаемый RewardType.
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
