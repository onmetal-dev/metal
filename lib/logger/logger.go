package logger

import (
	"context"
	"log/slog"
)

// Define a custom key type for context values
type contextKey int

const (
	loggerKey contextKey = iota
)

// AddToContext adds the logger to the context
func AddToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves the logger from the context
func FromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return logger
}
