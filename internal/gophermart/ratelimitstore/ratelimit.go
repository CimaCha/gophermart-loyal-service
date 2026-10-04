// Package ratelimitstore предоставляет потокобезопасное хранилище в оперативной памяти
// для отслеживания лимитов запросов (Rate Limiting) с автоматической очисткой устаревших записей.
package ratelimitstore

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// RateLimitStorage представляет собой потокобезопасное InMemory-хранилище счетчиков запросов,
// защищенное с помощью sync.RWMutex для безопасной работы из множества горутин.
type RateLimitStorage struct {
	m   map[string]*Item
	rwm sync.RWMutex

	l *slog.Logger
}

// Item описывает внутренний элемент хранилища, содержащий количество запросов
// и временную метку окончания действия текущего временного окна (TTL).
type Item struct {
	c         int
	expiresAt time.Time
}

// New создает и инициализирует новый экземпляр RateLimitStorage.
// Метод автоматически запускает фоновую горутину (cleanupWorker) для периодической
// очистки памяти от записей с истекшим сроком действия, которая завершится при отмене контекста ctx.
func New(ctx context.Context, cfg *Config, log *slog.Logger) *RateLimitStorage {
	r := &RateLimitStorage{m: map[string]*Item{}, l: log}
	go r.cleanupWorker(ctx, cfg.CleanupInterval) // воркер для очистки кэша
	return r
}

// Incr атомарно увеличивает счетчик запросов для указанного ключа (key) в рамках временного окна (window).
// Если ключ создается впервые или время его действия истекло, создается новое временное окно, а счетчик сбрасывается в 1.
// Возвращает текущее количество запросов и точное время истечения окна в формате UTC.
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
