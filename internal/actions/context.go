package actions

import "context"

type contextKey uint8

const actionContextKey contextKey = 0

func NewContext[T any](ctx context.Context, action T) context.Context {
	return context.WithValue(ctx, actionContextKey, action)
}

func FromContext[T any](ctx context.Context) T {
	return ctx.Value(actionContextKey).(T) //nolint:errcheck
}
