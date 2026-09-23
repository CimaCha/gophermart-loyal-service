package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URI")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URI")
	}
	if dsn == "" {
		t.Skip("neither TEST_DATABASE_URI nor DATABASE_URI is set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(context.Background()))
	t.Cleanup(pool.Close)
	return pool
}

func TestBalanceRepository_GetBalance(t *testing.T) {
	pool := setupTestDB(t)
	repo := New(pool)
	ctx := context.Background()

	userID := uuid.New()
	_, err := pool.Exec(ctx,
		`INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`,
		userID, "test_user_"+userID.String()[:8], "hash",
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})

	t.Run("no balance row returns zero", func(t *testing.T) {
		got, err := repo.GetBalance(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, model.Balance{Current: 0, Withdrawn: 0}, got)
	})

	t.Run("returns current and withdrawn", func(t *testing.T) {
		_, err := pool.Exec(ctx,
			`INSERT INTO balance (user_uuid, current, withdrawn) VALUES ($1, $2, $3)`,
			userID, 500.5, 42,
		)
		require.NoError(t, err)

		got, err := repo.GetBalance(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, model.Balance{Current: 500.5, Withdrawn: 42}, got)
	})
}
