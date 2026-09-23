package ratelimit

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRateLimitStorage_Incr_NewItem(t *testing.T) {
	t.Parallel()

	storage := &RateLimitStorage{
		m: make(map[string]*Item),
		l: newTestLogger(),
	}

	window := time.Minute

	count, expiresAt := storage.Incr("user", window)

	require.Equal(t, 1, count)

	item, ok := storage.m["user"]
	require.True(t, ok)

	require.Equal(t, 1, item.c)
	require.Equal(t, expiresAt, item.expiresAt)

	require.WithinDuration(
		t,
		time.Now().Add(window),
		item.expiresAt,
		100*time.Millisecond,
	)
}

func TestRateLimitStorage_Incr_ExistingItem(t *testing.T) {
	t.Parallel()

	storage := &RateLimitStorage{
		m: make(map[string]*Item),
		l: newTestLogger(),
	}

	window := time.Minute

	_, expiresAt := storage.Incr("user", window)

	count, newExpiresAt := storage.Incr("user", window)

	require.Equal(t, 2, count)
	require.Equal(t, expiresAt, newExpiresAt)
	require.Equal(t, 2, storage.m["user"].c)
}

func TestRateLimitStorage_Incr_ExpiredItem(t *testing.T) {
	t.Parallel()

	storage := &RateLimitStorage{
		m: map[string]*Item{
			"user": {
				c:         15,
				expiresAt: time.Now().Add(-time.Second),
			},
		},
		l: newTestLogger(),
	}

	count, _ := storage.Incr("user", time.Minute)

	require.Equal(t, 1, count)
	require.Equal(t, 1, storage.m["user"].c)
	require.True(t, storage.m["user"].expiresAt.After(time.Now()))
}

func TestRateLimitStorage_Cleanup(t *testing.T) {
	t.Parallel()

	storage := &RateLimitStorage{
		m: map[string]*Item{
			"expired1": {
				c:         1,
				expiresAt: time.Now().Add(-time.Second),
			},
			"expired2": {
				c:         1,
				expiresAt: time.Now().Add(-time.Minute),
			},
			"alive1": {
				c:         1,
				expiresAt: time.Now().Add(time.Minute),
			},
			"alive2": {
				c:         1,
				expiresAt: time.Now().Add(time.Hour),
			},
		},
		l: newTestLogger(),
	}

	storage.cleanup()

	require.Len(t, storage.m, 2)

	_, ok := storage.m["alive1"]
	require.True(t, ok)

	_, ok = storage.m["alive2"]
	require.True(t, ok)
}

func TestRateLimitStorage_Cleanup_EmptyStorage(t *testing.T) {
	t.Parallel()

	storage := &RateLimitStorage{
		m: make(map[string]*Item),
		l: newTestLogger(),
	}

	require.NotPanics(t, func() {
		storage.cleanup()
	})
}

func TestRateLimitStorage_CleanupWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &Config{
		CleanupInterval: 10 * time.Millisecond,
	}

	storage := NewRLS(ctx, cfg, newTestLogger())

	storage.m["expired"] = &Item{
		c:         1,
		expiresAt: time.Now().Add(-time.Second),
	}

	require.Eventually(t, func() bool {
		storage.rwm.RLock()
		defer storage.rwm.RUnlock()

		_, ok := storage.m["expired"]
		return !ok
	}, time.Second, 10*time.Millisecond)
}
