package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustDecimal парсит строку в decimal.Decimal, паникует при ошибке.
// Хелпер для тестов: строку парсим надёжнее, чем float64.
func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

// setupTestDB создаёт пул соединений для интеграционного теста.
// Возвращает пул с зарегистрированным decimal в TypeMap.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URI")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URI")
	}
	if dsn == "" {
		t.Skip("neither TEST_DATABASE_URI nor DATABASE_URI is set, skipping integration test")
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)

	poolCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
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
		assert.True(t, got.Current.Equal(decimal.Zero), "current should be zero")
		assert.True(t, got.Withdrawn.Equal(decimal.Zero), "withdrawn should be zero")
		assert.Equal(t, userID, got.UserID)
	})

	t.Run("returns current and withdrawn", func(t *testing.T) {
		_, err := pool.Exec(ctx,
			`INSERT INTO balance (user_uuid, current, withdrawn) VALUES ($1, $2, $3)`,
			userID, mustDecimal("500.5"), mustDecimal("42"),
		)
		require.NoError(t, err)

		got, err := repo.GetBalance(ctx, userID)
		require.NoError(t, err)
		assert.True(t, mustDecimal("500.5").Equal(got.Current), "current mismatch")
		assert.True(t, mustDecimal("42").Equal(got.Withdrawn), "withdrawn mismatch")
		assert.Equal(t, userID, got.UserID)
	})
}
