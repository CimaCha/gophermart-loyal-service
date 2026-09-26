package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
)

func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestService_GetBalance_Success(t *testing.T) {
	userID := uuid.New()
	expected := model.Balance{
		UserID:    userID,
		Current:   mustDecimal("500.5"),
		Withdrawn: mustDecimal("42"),
	}

	repo := NewMockBalanceRepository(t)
	repo.On("GetBalance", context.Background(), userID).Return(expected, nil)

	svc := New(repo)
	got, err := svc.GetBalance(context.Background(), userID)

	require.NoError(t, err)
	assert.True(t, expected.Current.Equal(got.Current), "current mismatch")
	assert.True(t, expected.Withdrawn.Equal(got.Withdrawn), "withdrawn mismatch")
	repo.AssertExpectations(t)
}

func TestService_GetBalance_RepoError(t *testing.T) {
	userID := uuid.New()
	repoErr := errors.New("db is down")

	repo := NewMockBalanceRepository(t)
	repo.On("GetBalance", context.Background(), userID).
		Return(model.Balance{}, repoErr)

	svc := New(repo)
	_, err := svc.GetBalance(context.Background(), userID)

	require.ErrorIs(t, err, repoErr)
	repo.AssertExpectations(t)
}
