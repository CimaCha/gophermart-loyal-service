package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/user/model"
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

func (r *UserRepository) FindUserInfo(ctx context.Context, userLogin string) (*model.UserInfo, error) {
	userInfo := &model.UserInfo{}
	err := r.pool.QueryRow(ctx, "SELECT * FROM users WHERE login = $1", userLogin).Scan(userInfo)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (r *UserRepository) SaveUserInfo(ctx context.Context, userInfo model.UserInfo) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO users(id, login, password_hash, created_at) VALUES ($1,$2, $3, $4)", userInfo.UUID, userInfo.Login, userInfo.PasswordHash, userInfo.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrUserAlreadyExists
	}
	return err
}
