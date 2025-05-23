package server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/api"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestServer_NewShutdown(t *testing.T) {
	_, cleanup := newTestServer(t, nil)
	defer cleanup()
}

func TestServer_StaticNodesReconnect(t *testing.T) {
	n1, cleanup := test.NewNode(t, test.NewLog(t))
	defer cleanup()
	n1ID := n1.Host().ID()

	n2, cleanup := test.NewNode(t, test.NewLog(t))
	defer cleanup()
	n2ID := n2.Host().ID()

	server, cleanup := newTestServer(t, []string{
		n1.ListenAddresses()[0].String(),
		n2.ListenAddresses()[0].String(),
	})
	defer cleanup()

	// Expect connect to static nodes.
	test.ExpectPeers(t, server.wakuNode, n1ID, n2ID)

	// Disconnect from static nodes.
	test.Disconnect(t, server.wakuNode, n1)
	test.Disconnect(t, server.wakuNode, n2)

	// Expect reconnect to static nodes.
	test.ExpectPeers(t, server.wakuNode, n1ID, n2ID)
}

func newTestServer(t *testing.T, staticNodes []string) (*Server, func()) {
	_, dbDSN, _ := test.NewDB(t)
	s, err := New(context.Background(), test.NewLog(t), Options{
		NodeKey: newNodeKey(t),
		Address: "localhost",
		Store: store.Options{
			Enable:                   true,
			DbConnectionString:       dbDSN,
			DbReaderConnectionString: dbDSN,
		},
		StaticNodes: staticNodes,
		WSAddress:   "0.0.0.0",
		WSPort:      0,
		EnableWS:    true,
		API: api.Options{
			HTTPPort: 0,
			GRPCPort: 0,
		},
		Metrics: MetricsOptions{
			Enable: true,
		},
		MetricsPeriod: 5 * time.Second,
	})
	require.NoError(t, err)
	require.NotNil(t, s)
	return s, func() {
		s.Shutdown()
	}
}

func newNodeKey(t *testing.T) string {
	key, err := generatePrivateKey()
	require.NoError(t, err)
	return string(key)
}
