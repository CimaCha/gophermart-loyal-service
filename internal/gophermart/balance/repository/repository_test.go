package repository

import (
	"context"
	"os"
	"testing"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
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

func TestBalanceRepository_Withdraw(t *testing.T) {
	ctx := context.Background()
	repo := New(testPool)

	// helper: создаёт юзера и баланс с заданным current
	setupUserWithBalance := func(t *testing.T, current string) uuid.UUID {
		t.Helper()
		userID := uuid.New()
		_, err := testPool.Exec(ctx,
			`INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`,
			userID, "u_"+userID.String()[:8], "hash",
		)
		require.NoError(t, err)

		_, err = testPool.Exec(ctx,
			`INSERT INTO balance (user_uuid, current, withdrawn) VALUES ($1, $2, 0)`,
			userID, mustDecimal(current),
		)
		require.NoError(t, err)
		return userID
	}

	t.Run("success withdraws and records transaction", func(t *testing.T) {
		cleanDB(t)
		userID := setupUserWithBalance(t, "500.5")

		err := repo.Withdraw(ctx, userID, "12345678903", mustDecimal("100"))
		require.NoError(t, err)

		// баланс обновился
		var b model.Balance
		err = testPool.QueryRow(ctx,
			`SELECT current, withdrawn FROM balance WHERE user_uuid = $1`, userID,
		).Scan(&b.Current, &b.Withdrawn)
		require.NoError(t, err)
		assert.True(t, mustDecimal("400.5").Equal(b.Current), "current")
		assert.True(t, mustDecimal("100").Equal(b.Withdrawn), "withdrawn")

		// транзакция записана
		var count int
		err = testPool.QueryRow(ctx,
			`SELECT COUNT(*) FROM transactions WHERE user_uuid = $1`, userID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("insufficient funds does not change balance", func(t *testing.T) {
		cleanDB(t)
		userID := setupUserWithBalance(t, "50")

		err := repo.Withdraw(ctx, userID, "12345678903", mustDecimal("100"))
		require.ErrorIs(t, err, model.ErrInsufficientFunds)

		// баланс не изменился, транзакции нет
		var current decimal.Decimal
		err = testPool.QueryRow(ctx,
			`SELECT current FROM balance WHERE user_uuid = $1`, userID,
		).Scan(&current)
		require.NoError(t, err)
		assert.True(t, mustDecimal("50").Equal(current))

		var count int
		err = testPool.QueryRow(ctx,
			`SELECT COUNT(*) FROM transactions WHERE user_uuid = $1`, userID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("no balance row returns insufficient funds", func(t *testing.T) {
		cleanDB(t)
		userID := uuid.New()
		_, err := testPool.Exec(ctx,
			`INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`,
			userID, "u_"+userID.String()[:8], "hash",
		)
		require.NoError(t, err)

		err = repo.Withdraw(ctx, userID, "12345678903", mustDecimal("100"))
		require.ErrorIs(t, err, model.ErrInsufficientFunds)
	})

	t.Run("duplicate order returns ErrOrderAlreadyUsed", func(t *testing.T) {
		cleanDB(t)
		userID := setupUserWithBalance(t, "500")

		// первое списание
		err := repo.Withdraw(ctx, userID, "12345678903", mustDecimal("100"))
		require.NoError(t, err)

		// второе с тем же номером — unique violation
		err = repo.Withdraw(ctx, userID, "12345678903", mustDecimal("100"))
		require.ErrorIs(t, err, model.ErrOrderAlreadyUsed)

		// после отката баланс остался после первого списания
		var current decimal.Decimal
		err = testPool.QueryRow(ctx,
			`SELECT current FROM balance WHERE user_uuid = $1`, userID,
		).Scan(&current)
		require.NoError(t, err)
		assert.True(t, mustDecimal("400").Equal(current))
	})
}
