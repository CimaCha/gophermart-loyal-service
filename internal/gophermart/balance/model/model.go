package model

import "github.com/google/uuid"

type Balance struct {
	UserID    uuid.UUID `json:"-"`
	Current   float64   `json:"current"`
	Withdrawn float64   `json:"withdrawn"`
}
