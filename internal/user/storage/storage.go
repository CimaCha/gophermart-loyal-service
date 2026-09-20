package storage

import (
	"context"
	"github.com/CimaCha/gophermart-loyal-service/internal/user/model"
)

type UserStorage interface {
	FindUserInfo(ctx context.Context, userLogin string) (model.UserInfo, error)
	SaveUserInfo(ctx context.Context, userInfo model.UserInfo) error
}
