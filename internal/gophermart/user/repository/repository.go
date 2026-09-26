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
	query := `
		SELECT id, login, password_hash, created_at
		FROM users WHERE login = $1
	`
	err := r.pool.QueryRow(ctx, query, userLogin).Scan(&userInfo.UUID, &userInfo.Login, &userInfo.PasswordHash, &userInfo.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return userInfo, nil
}

func (r *UserRepository) SaveUserInfo(ctx context.Context, userInfo model.UserInfo) error {
	query := `
		INSERT INTO users(id, login, password_hash, created_at)
		VALUES ($1,$2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, query, userInfo.UUID, userInfo.Login, userInfo.PasswordHash, userInfo.CreatedAt)
	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUserAlreadyExists
		}
		return err
	}
	return nil
}
