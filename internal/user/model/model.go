package model

import "time"

type UserRegisterIn struct {
	Login    string
	Password string
}

type UserLoginIn struct {
	Login    string
	Password string
}

type UserInfo struct {
	UUID      string
	Login     string
	Password  string
	CreatedAt time.Time
}
