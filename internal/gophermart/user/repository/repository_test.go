package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/testenv"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/user/model"
	gophermartMigrations "github.com/CimaCha/gophermart-loyal-service/migrations/gophermart"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	_, err := testPool.Exec(context.Background(), `TRUNCATE users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func newRepo(t *testing.T) *UserRepository {
	t.Helper()

	return New(testPool)
}

// CreateUser
func TestCreateUser_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	repo := newRepo(t)

	var (
		expectedUUID         = uuid.New()
		expectedLogin        = "test-login"
		expectedPasswordHash = "test-password-hash"
		expectedCreatedAt    = time.Now().UTC()
	)

	user := model.UserInfo{
		UUID:         expectedUUID,
		Login:        expectedLogin,
		PasswordHash: expectedPasswordHash,
		CreatedAt:    expectedCreatedAt,
	}

	err := repo.SaveUserInfo(ctx, user)
	require.NoError(t, err)

	queryCheck := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	result := &model.UserInfo{}

	err = testPool.QueryRow(ctx, queryCheck, expectedUUID).
		Scan(
			&result.UUID,
			&result.Login,
			&result.PasswordHash,
			&result.CreatedAt,
		)
	require.NoError(t, err)

	assert.Equal(t, user.UUID, result.UUID)
	assert.Equal(t, user.Login, result.Login)
	assert.Equal(t, user.PasswordHash, result.PasswordHash)

	assert.WithinDuration(t, user.CreatedAt, result.CreatedAt, time.Second)
}

func TestCreateUser_Duplicate(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	repo := newRepo(t)

	user1 := model.UserInfo{
		UUID:         uuid.New(),
		Login:        "same-login",
		PasswordHash: "test-password-hash",
		CreatedAt:    time.Now().UTC(),
	}

	user2 := model.UserInfo{
		UUID:         uuid.New(),
		Login:        "same-login",
		PasswordHash: "test-password-hash",
		CreatedAt:    time.Now().UTC(),
	}

	err := repo.SaveUserInfo(ctx, user1)
	require.NoError(t, err)

	err = repo.SaveUserInfo(ctx, user2)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUserAlreadyExists)

	queryCheck := `
		SELECT COUNT(*) AS users_count FROM users
	`

	var count int

	err = testPool.QueryRow(ctx, queryCheck).Scan(&count)
	require.NoError(t, err)

	require.Equal(t, 1, count)
}

// FindUserInfo
func TestFindUserInfo_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	repo := newRepo(t)

	var (
		expectedUUID         = uuid.New()
		expectedLogin        = "test-login"
		expectedPasswordHash = "test-password-hash"
		expectedCreatedAt    = time.Now().UTC()
	)

	query := `
		INSERT INTO users(id, login, password_hash, created_at)
		VALUES ($1,$2, $3, $4)
	`

	_, err := testPool.Exec(
		ctx,
		query,
		expectedUUID,
		expectedLogin,
		expectedPasswordHash,
		expectedCreatedAt,
	)
	require.NoError(t, err)

	result, err := repo.FindUserInfo(ctx, expectedLogin)
	require.NoError(t, err)

	assert.Equal(t, expectedUUID, result.UUID)
	assert.Equal(t, expectedLogin, result.Login)
	assert.Equal(t, expectedPasswordHash, result.PasswordHash)

	assert.WithinDuration(t, expectedCreatedAt, result.CreatedAt, time.Second)
}

func TestFindUserInfo_UserNotFound(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	repo := newRepo(t)

	result, err := repo.FindUserInfo(ctx, "non-exist")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUserNotFound)

	require.Nil(t, result)
}
