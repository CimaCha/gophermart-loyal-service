package model

import (
	"time"

	"github.com/google/uuid"
)

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserInfo struct {
	UUID         uuid.UUID
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}
