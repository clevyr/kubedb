package notifier

import "context"

type contextKey uint8

const notifierContextKey contextKey = iota

func NewContext(ctx context.Context, n Notifier) context.Context {
	return context.WithValue(ctx, notifierContextKey, n)
}

func FromContext(ctx context.Context) (Notifier, bool) {
	notifier, ok := ctx.Value(notifierContextKey).(Notifier)
	return notifier, ok
}
