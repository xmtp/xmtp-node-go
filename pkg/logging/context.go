package logging

import (
	"context"

	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
)

var (
	// Re-export the go-waku helpers.
	With = logging.With
)

// From returns a logger from the context or the default logger.
func From(ctx context.Context) *zap.Logger {
	logger := logging.From(ctx)
	if logger == nil {
		logger = utils.Logger()
	}
	return logger
}
