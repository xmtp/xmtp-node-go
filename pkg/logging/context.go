package logging

import (
	"context"

	"go.uber.org/zap"
)

var logKey = &struct{}{}

// With associates a Logger with a Context to allow passing
// a logger down the call chain.
func With(ctx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(ctx, logKey, log)
}

// From returns a logger from the context or the default logger.
func From(ctx context.Context) *zap.Logger {
	logger, _ := ctx.Value(logKey).(*zap.Logger)
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return logger
}
