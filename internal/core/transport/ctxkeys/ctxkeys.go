package ctxkeys

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
)

func WithUserID(ctx context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

func GetUserID(ctx context.Context) (uuid.UUID, error) {
	uid, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf(
			"failed to get user ID from context, actual type: %T",
			ctx.Value(userIDKey),
		)
	}
	return uid, nil
}
