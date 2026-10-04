package model

import "errors"

var (
	// ErrInvalidGoods возвращается, когда структура с информацией о товаре не прошла валидацию.
	ErrInvalidGoods = errors.New("invalid goods info")
	// ErrMatchAlreadyExists возвращается, если правило для данного шаблона сопоставления уже зарегистрировано.
	ErrMatchAlreadyExists = errors.New("match already exists")
)
