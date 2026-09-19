package repository

import (
	"context"
	"testing"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/order/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanDB(t *testing.T) {
	t.Helper()

	_, err := testPool.Exec(context.Background(), `TRUNCATE users, orders RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func newRepo(t *testing.T) *OrderRepository {
	t.Helper()

	return New(testPool)
}

// CreateOrder
func TestCreateOrder_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	repo := newRepo(t)

	var (
		expectedOrderNum   int64 = 12345678903
		expectedUserID           = uuid.New()
		expectedStatus           = "NEW"
		expectedUploadedAt       = time.Now().UTC()
	)

	userQuery := `INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`
	_, err := testPool.Exec(ctx, userQuery, expectedUserID, "test_login", "test_password_hash")
	require.NoError(t, err)

	var order = &model.Order{
		OrderNum:   expectedOrderNum,
		UserID:     expectedUserID,
		Status:     expectedStatus,
		UploadedAt: expectedUploadedAt,
	}

	err = repo.CreateOrder(ctx, order)
	require.NoError(t, err)

	query := `
		SELECT order_num, user_id, order_status, accrual, uploaded_at
			FROM orders WHERE order_num = $1
		`

	var result model.Order

	err = testPool.QueryRow(
		ctx,
		query,
		expectedOrderNum,
	).Scan(
		&result.OrderNum,
		&result.UserID,
		&result.Status,
		&result.Accrual,
		&result.UploadedAt,
	)
	require.NoError(t, err)

	require.Nil(t, result.Accrual)

	assert.Equal(t, expectedOrderNum, result.OrderNum)
	assert.Equal(t, expectedUserID, result.UserID)
	assert.Equal(t, expectedStatus, result.Status)
	require.WithinDuration(t, expectedUploadedAt, result.UploadedAt, time.Millisecond)
}

func TestCreateOrder_OrderAlreadyProcessing(t *testing.T) {
	cleanDB(t)

	ctx := context.Background()
	repo := newRepo(t)

	userID := uuid.New()

	_, err := testPool.Exec(
		ctx,
		`INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`,
		userID,
		"test_login",
		"test_password_hash",
	)
	require.NoError(t, err)

	order := &model.Order{
		OrderNum:   12345678903,
		UserID:     userID,
		Status:     "NEW",
		UploadedAt: time.Now().UTC(),
	}

	require.NoError(t, repo.CreateOrder(ctx, order))

	err = repo.CreateOrder(ctx, order)

	require.ErrorIs(t, err, ErrOrderAlreadyProcessing)
}

func TestCreateOrder_OrderAlreadyCreatedByAnotherUser(t *testing.T) {
	cleanDB(t)

	ctx := context.Background()
	repo := newRepo(t)

	users := []struct {
		UserID uuid.UUID
		Login  string
	}{
		{uuid.New(), "user1"},
		{uuid.New(), "user2"},
	}

	for _, user := range users {
		_, err := testPool.Exec(
			ctx,
			`INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`,
			user.UserID,
			user.Login,
			"hash",
		)
		require.NoError(t, err)
	}

	firstOrder := &model.Order{
		OrderNum:   12345678903,
		UserID:     users[0].UserID,
		Status:     "NEW",
		UploadedAt: time.Now().UTC(),
	}

	require.NoError(t, repo.CreateOrder(ctx, firstOrder))

	secondOrder := &model.Order{
		OrderNum:   12345678903,
		UserID:     users[1].UserID,
		Status:     "NEW",
		UploadedAt: time.Now().UTC(),
	}

	err := repo.CreateOrder(ctx, secondOrder)

	require.ErrorIs(t, err, ErrOrderAlreadyCreatedByAnotherUser)
}

// GetOrders
func TestGetOrders_Success(t *testing.T) {
	cleanDB(t)

	ctx := context.Background()
	repo := newRepo(t)

	userID := uuid.New()
	anotherUserID := uuid.New()

	users := []struct {
		id    uuid.UUID
		login string
	}{
		{userID, "user1"},
		{anotherUserID, "user2"},
	}

	for _, u := range users {
		_, err := testPool.Exec(
			ctx,
			`INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`,
			u.id,
			u.login,
			"hash",
		)
		require.NoError(t, err)
	}

	oldTime := time.Now().UTC().Add(-time.Hour).Truncate(time.Millisecond)
	newTime := time.Now().UTC().Truncate(time.Millisecond)

	_, err := testPool.Exec(ctx, `
		INSERT INTO orders(order_num, user_id, order_status, accrual, uploaded_at)
		VALUES
			($1,$2,$3,$4,$5),
			($6,$7,$8,$9,$10),
			($11,$12,$13,$14,$15)
	`,
		12345678903, userID, "NEW", nil, oldTime,
		79927398713, userID, "PROCESSED", nil, newTime,
		11111111111, anotherUserID, "NEW", nil, oldTime,
	)
	require.NoError(t, err)

	expected := []model.Order{
		{
			OrderNum:   12345678903,
			Status:     "NEW",
			UploadedAt: oldTime,
		},
		{
			OrderNum:   79927398713,
			Status:     "PROCESSED",
			UploadedAt: newTime,
		},
	}

	orders, err := repo.GetOrders(ctx, userID)

	require.NoError(t, err)
	require.Len(t, orders, 2)

	assert.Equal(t, expected[0].OrderNum, orders[0].OrderNum)
	assert.Equal(t, expected[0].Status, orders[0].Status)
	assert.True(t, expected[0].UploadedAt.Equal(orders[0].UploadedAt))

	assert.Equal(t, expected[1].OrderNum, orders[1].OrderNum)
	assert.Equal(t, expected[1].Status, orders[1].Status)
	assert.True(t, expected[1].UploadedAt.Equal(orders[1].UploadedAt))
}

func TestGetOrders_Empty(t *testing.T) {
	cleanDB(t)

	ctx := context.Background()
	repo := newRepo(t)

	userID := uuid.New()

	_, err := testPool.Exec(
		ctx,
		`INSERT INTO users(id, login, password_hash) VALUES ($1,$2,$3)`,
		userID,
		"test_login",
		"test_hash",
	)
	require.NoError(t, err)

	orders, err := repo.GetOrders(ctx, userID)

	require.NoError(t, err)
	assert.Empty(t, orders)
}
