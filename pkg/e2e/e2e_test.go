package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	localNetworkEnv = "local"
	localNodesURL   = "http://localhost:8000"
	localAPIURL     = "http://localhost:8080"
)

func TestE2E(t *testing.T) {
	ctx := context.Background()
	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	apiURL := os.Getenv("XMTPD_E2E_API_URL")
	if apiURL == "" {
		apiURL = localAPIURL
	}
	s := NewSuite(ctx, log, &Config{
		NetworkEnv: localNetworkEnv,
		NodesURL:   localNodesURL,
		APIURL:     apiURL,
	})

	for _, test := range s.Tests() {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			err := test.Run(log)
			require.NoError(t, err)
		})
	}
}
