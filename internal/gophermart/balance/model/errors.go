package model

import "errors"

var (
	// Номер заказа не прошёл валидацию Луна.
	ErrInvalidOrderNumber = errors.New("invalid order number")

	// На балансе недостаточно средств.
	ErrInsufficientFunds = errors.New("insufficient funds")

	// Заказ уже использовался для списания.
	ErrOrderAlreadyUsed = errors.New("order already used")
)
