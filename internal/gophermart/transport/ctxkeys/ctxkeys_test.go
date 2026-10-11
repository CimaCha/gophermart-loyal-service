package ctxkeys

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWithUserIDAndGetUserID(t *testing.T) {
	t.Parallel()

	want := uuid.New()

	ctx := WithUserID(context.Background(), want)

	got, err := GetUserID(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetUserID_ContextWithoutUserID(t *testing.T) {
	t.Parallel()

	got, err := GetUserID(context.Background())

	require.Error(t, err)
	require.Equal(t, uuid.Nil, got)
	require.ErrorContains(t, err, "user ID not found")
}

func TestGetUserID_InvalidType(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), userIDKey, "not uuid")

	got, err := GetUserID(ctx)

	require.Error(t, err)
	require.Equal(t, uuid.Nil, got)
	require.ErrorContains(t, err, "invalid type")
}

func TestWithUserID_OverridesPreviousValue(t *testing.T) {
	t.Parallel()

	first := uuid.New()
	second := uuid.New()

	ctx := WithUserID(context.Background(), first)
	ctx = WithUserID(ctx, second)

	got, err := GetUserID(ctx)
	require.NoError(t, err)
	require.Equal(t, second, got)
}
