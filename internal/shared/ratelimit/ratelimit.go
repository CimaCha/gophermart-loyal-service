package ratelimit

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type RateLimitStorage struct {
	m   map[string]*Item
	rwm sync.RWMutex

	l *slog.Logger
}

type Item struct {
	c         int
	expiresAt time.Time
}

// Создаёт RateLimitStorage и запускает горутину воркера для очистки кэша от устаревших записей
func NewRLS(ctx context.Context, cfg *Config, log *slog.Logger) *RateLimitStorage {
	r := &RateLimitStorage{m: map[string]*Item{}, l: log}
	go r.cleanupWorker(ctx, cfg.CleanupInterval) // воркер для очистки кэша
	return r
}

func (r *RateLimitStorage) Incr(key string, window time.Duration) (int, time.Time) {
	r.rwm.Lock()
	defer r.rwm.Unlock()

	now := time.Now().UTC()
	v, ok := r.m[key]

	if !ok || v.expiresAt.Before(now) {
		newItem := &Item{expiresAt: now.Add(window), c: 1}
		r.m[key] = newItem
		return 1, newItem.expiresAt
	}

	v.c++
	return v.c, v.expiresAt
}

func (r *RateLimitStorage) cleanupWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.cleanup()
		}
	}
}

func (r *RateLimitStorage) cleanup() {
	r.rwm.Lock()
	defer r.rwm.Unlock()

	now := time.Now().UTC()
	deletedCount := 0

	for key, item := range r.m {

		if item.expiresAt.Before(now) {
			delete(r.m, key)
			deletedCount++
		}
	}

	if deletedCount > 0 {
		r.l.Debug(
			"RLS cleanup worker: Cleanup completed",
			"removed_count", deletedCount,
		)
	}
}
