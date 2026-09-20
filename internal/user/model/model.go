package model

import (
	"github.com/google/uuid"
	"time"
)

type UserRegisterIn struct {
	Login    string
	Password string
}

type UserLoginIn struct {
	Login    string
	Password string
}

type UserInfo struct {
	UUID         uuid.UUID
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}
