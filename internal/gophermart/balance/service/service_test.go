package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
)

func TestService_GetBalance_Success(t *testing.T) {
	userID := uuid.New()
	expected := model.Balance{Current: 500.5, Withdrawn: 42}

	repo := NewMockBalanceRepository(t)
	repo.On("GetBalance", context.Background(), userID).Return(expected, nil)

	svc := New(repo)
	got, err := svc.GetBalance(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestService_GetBalance_RepoError(t *testing.T) {
	userID := uuid.New()
	repoErr := errors.New("db is down")

	repo := NewMockBalanceRepository(t)
	repo.On("GetBalance", context.Background(), userID).Return(model.Balance{}, repoErr)

	svc := New(repo)
	_, err := svc.GetBalance(context.Background(), userID)

	require.ErrorIs(t, err, repoErr)
	repo.AssertExpectations(t)
}
