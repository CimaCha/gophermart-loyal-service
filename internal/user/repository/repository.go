package repository

import (
	"context"
	"errors"

	"github.com/CimaCha/gophermart-loyal-service/internal/user/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) FindUserInfo(ctx context.Context, userLogin string) (model.UserInfo, error) {
	// TODO
	return model.UserInfo{}, nil
}

func (r *UserRepository) SaveUserInfo(ctx context.Context, userInfo model.UserInfo) error {
	// TODO
	return nil
}
