package domain

import "context"

type contextKey string

const requestIDContextKey contextKey = "request_id"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(requestIDContextKey).(string)
	return value, ok
}
