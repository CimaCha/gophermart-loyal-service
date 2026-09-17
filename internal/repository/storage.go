package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var (
	ErrPingDataBase  = errors.New("failed to ping database")
	ErrCreateNewPool = errors.New("failed to create new pool")
)

type Storage struct {
	Pool *pgxpool.Pool
}

func NewDatabaseStorage(ctx context.Context, databaseURL string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateNewPool, err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err = pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: %w", ErrPingDataBase, err)
	}
	return &Storage{Pool: pool}, nil
}

func (s Storage) Close() {
	s.Pool.Close()
}
