package repository

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/model"
	"github.com/CimaCha/gophermart-loyal-service/internal/testenv"
	accrualMigrations "github.com/CimaCha/gophermart-loyal-service/migrations/accrual"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	env := testenv.Setup(ctx, accrualMigrations.EmbedMigrations)
	testPool = env.Pool

	code := m.Run()
	env.Close(ctx)
	os.Exit(code)
}

func cleanDB(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `TRUNCATE rewards RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

// TestGoodsRepository_RegisterGoods_Success проверяет, что валидное правило
// сохраняется в таблицу rewards.
func TestGoodsRepository_RegisterGoods_Success(t *testing.T) {
	tests := []struct {
		name       string
		rewardType model.RewardType
		storedType string
	}{
		{name: "percent", rewardType: model.RewardTypePercent, storedType: "PERCENT"},
		{name: "points", rewardType: model.RewardTypePoints, storedType: "POINTS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanDB(t)
			repo := New(testPool)
			ctx := context.Background()
			d := mustDecimal("10")
			goods := model.GoodsInfo{Match: "Bork", Reward: &d, RewardType: tt.rewardType}
			require.NoError(t, repo.RegisterGoods(ctx, goods))
			var match, rewardType string
			var reward decimal.Decimal
			err := testPool.QueryRow(ctx,
				`SELECT match, reward_value, reward_type FROM rewards WHERE match = $1`, "Bork",
			).Scan(&match, &reward, &rewardType)
			require.NoError(t, err)
			assert.Equal(t, "Bork", match)
			assert.True(t, d.Equal(reward))
			assert.Equal(t, tt.storedType, rewardType)
		})
	}
}

// TestGoodsRepository_RegisterGoods_Duplicate проверяет, что повторная
// регистрация того же match возвращает ErrMatchAlreadyExists.
func TestGoodsRepository_RegisterGoods_Duplicate(t *testing.T) {
	cleanDB(t)
	repo := New(testPool)
	ctx := context.Background()

	d := mustDecimal("10")
	goods := model.GoodsInfo{
		Match:      "Bork",
		Reward:     &d,
		RewardType: model.RewardTypePercent,
	}

	require.NoError(t, repo.RegisterGoods(ctx, goods))

	err := repo.RegisterGoods(ctx, goods)
	require.ErrorIs(t, err, model.ErrMatchAlreadyExists)
}

// TestGoodsRepository_GetAllGoods_Success проверяет, что метод возвращает
// все правила вознаграждений из таблицы rewards.
func TestGoodsRepository_GetAllGoods_Success(t *testing.T) {
	cleanDB(t)

	repo := New(testPool)
	ctx := context.Background()

	d1 := mustDecimal("10")
	d2 := mustDecimal("500")

	_, err := testPool.Exec(ctx, `INSERT INTO rewards (match, reward_value, reward_type) VALUES
        ('Bork', 10, 'PERCENT'), ('iPhone', 500, 'POINTS')`)
	require.NoError(t, err)

	got, err := repo.GetAllGoods(ctx)
	require.NoError(t, err)

	require.Len(t, got, 2)

	expected := map[string]model.GoodsInfo{
		"Bork": {
			Reward:     &d1,
			RewardType: model.RewardTypePercent,
		},
		"iPhone": {
			Reward:     &d2,
			RewardType: model.RewardTypePoints,
		},
	}

	for _, g := range got {
		exp, ok := expected[g.Match]
		require.True(t, ok, "unexpected match: %s", g.Match)
		require.NotNil(t, g.Reward)
		delete(expected, g.Match)

		assert.Equal(t, exp.RewardType, g.RewardType)
		assert.True(t, exp.Reward.Equal(*g.Reward))
	}
	require.Empty(t, expected)
}

// TestGoodsRepository_GetAllGoods_Empty проверяет, что при отсутствии правил
// метод возвращает пустой слайс без ошибки.
func TestGoodsRepository_GetAllGoods_Empty(t *testing.T) {
	cleanDB(t)

	repo := New(testPool)
	ctx := context.Background()

	got, err := repo.GetAllGoods(ctx)
	require.NoError(t, err)

	assert.Empty(t, got)
}
