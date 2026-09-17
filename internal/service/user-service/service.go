package user_service

import (
	"context"
)

type Service struct {
	userStorage UserStorage
}

func (s Service) RegisterUser(ctx context.Context, userID string, password string) error {
	//TODO implement me
	panic("implement me")
}

func NewService(storage UserStorage) Service {
	return Service{userStorage: storage}
}
