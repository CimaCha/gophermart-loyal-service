// Package ctxkeys предоставляет строго типизированные инструменты для работы с ключами контекста,
// обеспечивая безопасное сохранение и извлечение идентификаторов пользователей в рамках HTTP-запросов.
package ctxkeys

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type contextKey struct{}

var (
	userIDKey = contextKey{}
)

// WithUserID принимает родительский контекст и UUID пользователя, упаковывает их
// с использованием неэкспортируемого ключа типа contextKey и возвращает новый дочерний контекст.
// Используется в middleware аутентификации для передачи ID пользователя вниз по цепочке обработчиков.
func WithUserID(ctx context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

// GetUserID извлекает UUID пользователя из переданного контекста.
// Если идентификатор отсутствует в контексте или сохраненное значение имеет некорректный тип,
// метод возвращает uuid.Nil и структурированную ошибку для последующего логирования и прерывания HTTP-запроса.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	uid, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID not found in context or has invalid type")
	}
	return uid, nil
}
