package diag

import "context"

type contextKeys string

const requestIDKey contextKeys = "requestID"

// ContextWithRequestID - create context with requestID
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDValue - returns requestID value taken from context
func RequestIDValue(ctx context.Context) string {
	val := ctx.Value(requestIDKey)
	if val == nil {
		return ""
	}
	return val.(string)
}
