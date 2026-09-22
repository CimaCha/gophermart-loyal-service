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

func WithUserID(ctx context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

func GetUserID(ctx context.Context) (uuid.UUID, error) {
	uid, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID not found in context or has invalid type")
	}
	return uid, nil
}
