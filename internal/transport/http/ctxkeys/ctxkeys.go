package ctxkeys

import (
	"context"
	"fmt"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
)

func WithUserID(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

func GetUserID(ctx context.Context) (string, error) {
	uid, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return "", fmt.Errorf(
			"failed to get user ID from context, actual type: %T",
			ctx.Value(userIDKey),
		)
	}
	return uid, nil
}
