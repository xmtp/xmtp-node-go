package testing

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func NewLog(t *testing.T) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	if !testing.Verbose() {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	log, err := cfg.Build()
	require.NoError(t, err)
	return log
}
