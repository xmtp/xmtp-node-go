package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	localNetworkEnv = "local"
	localNodesURL   = "http://localhost:8000"
)

func TestE2E(t *testing.T) {
	ctx := context.Background()
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	e := New(ctx, log, &Config{
		NetworkEnv: localNetworkEnv,
		NodesURL:   localNodesURL,
	})
	err = e.Run()
	require.NoError(t, err)
}
