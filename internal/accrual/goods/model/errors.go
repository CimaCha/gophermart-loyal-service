package model

import "errors"

var (
	ErrInvalidGoods       = errors.New("invalid goods info")
	ErrMatchAlreadyExists = errors.New("match already exists")
)
