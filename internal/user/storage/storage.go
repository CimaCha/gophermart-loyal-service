package storage

import (
	"context"
	"errors"

	"github.com/CimaCha/gophermart-loyal-service/internal/user/model"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

//go:generate go tool mockgen -source=storage.go -destination=../service/mock/user_storage_gen.go -package=mock

type UserStorage interface {
	FindUserInfo(ctx context.Context, userLogin string) (model.UserInfo, error)
	SaveUserInfo(ctx context.Context, userInfo model.UserInfo) error
}
