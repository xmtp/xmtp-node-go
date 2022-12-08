package server

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/api"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestServer_NewShutdown(t *testing.T) {
	t.Parallel()

	_, cleanup := newTestServer(t)
	defer cleanup()
}

func newTestServer(t *testing.T) (*Server, func()) {
	_, dbDSN, dbCleanup := test.NewDB(t)
	s, err := New(context.Background(), test.NewLog(t), Options{
		MessageDBDSN: dbDSN,
		API: api.Options{
			HTTPPort: 0,
			GRPCPort: 0,
		},
		Metrics: MetricsOptions{
			Enable:       true,
			StatusPeriod: 5 * time.Second,
			Port:         0,
		},
		DataPath: path.Join(t.TempDir(), test.RandomStringLower(13), "data"),
	})
	require.NoError(t, err)
	require.NotNil(t, s)
	return s, func() {
		s.Shutdown()
		dbCleanup()
	}
}
