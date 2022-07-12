package testing

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func NewLog(t *testing.T) *zap.Logger {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	return log
}
