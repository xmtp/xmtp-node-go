package server

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/api"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestServer_NewShutdown(t *testing.T) {
	t.Parallel()

	_, cleanup := newTestServer(t, nil)
	defer cleanup()
}

func newTestServer(t *testing.T, staticNodes []string) (*Server, func()) {
	_, dbDSN, _ := test.NewDB(t)

	ns, err := server.NewServer(&server.Options{
		Port: server.RANDOM_PORT,
	})
	require.NoError(t, err)
	go ns.Start()
	require.True(t, ns.ReadyForConnections(4*time.Second), "nats server not ready")

	s, err := New(context.Background(), test.NewLog(t), Options{
		NATS: NATSOptions{
			URL: ns.ClientURL(),
		},
		Store: StoreOptions{
			Enable:                   true,
			DbConnectionString:       dbDSN,
			DbReaderConnectionString: dbDSN,
		},
		API: api.Options{
			HTTPPort: 0,
			GRPCPort: 0,
		},
		Metrics: MetricsOptions{
			Enable:       true,
			StatusPeriod: 5 * time.Second,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, s)
	return s, func() {
		s.Shutdown()
		ns.Shutdown()
	}
}
