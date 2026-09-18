package service

import (
	"context"
	"errors"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserService struct {
	ctx     context.Context
	storage UserStorage
}

func NewUserService(ctx context.Context, storage UserStorage) UserService {
	return UserService{
		ctx:     ctx,
		storage: storage,
	}
}

func (s *UserService) CreateUser(userLogin string, password string) error {
	return nil
}

func (s *UserService) LoginUser(userLogin string, password string) error {
	return nil
}
