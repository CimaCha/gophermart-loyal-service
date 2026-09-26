package repository

import (
	"context"
	"os"
	"testing"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/testenv"
	gophermartMigrations "github.com/CimaCha/gophermart-loyal-service/migrations/gophermart"
	"github.com/google/uuid"
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

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	env := testenv.Setup(ctx, gophermartMigrations.EmbedMigrations)
	testPool = env.Pool

	code := m.Run()
	env.Close(ctx)
	os.Exit(code)
}

func cleanDB(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `TRUNCATE users, balance RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func TestBalanceRepository_GetBalance(t *testing.T) {
	cleanDB(t)
	repo := New(testPool)
	ctx := context.Background()

	userID := uuid.New()
	_, err := testPool.Exec(ctx,
		`INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`,
		userID, "test_user_"+userID.String()[:8], "hash",
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})

	t.Run("no balance row returns zero", func(t *testing.T) {
		got, err := repo.GetBalance(ctx, userID)
		require.NoError(t, err)
		assert.True(t, got.Current.Equal(decimal.Zero), "current should be zero")
		assert.True(t, got.Withdrawn.Equal(decimal.Zero), "withdrawn should be zero")
		assert.Equal(t, userID, got.UserID)
	})

	t.Run("returns current and withdrawn", func(t *testing.T) {
		_, err := testPool.Exec(ctx,
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
