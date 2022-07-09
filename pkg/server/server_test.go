package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestServer_NewShutdown(t *testing.T) {
	t.Parallel()

	_, cleanup := newTestServer(t, nil)
	defer cleanup()
}

func TestServer_StaticNodesReconnect(t *testing.T) {
	t.Parallel()

	n1, cleanup := test.NewNode(t, nil)
	defer cleanup()
	n1ID := n1.Host().ID()

	n2, cleanup := test.NewNode(t, nil)
	defer cleanup()
	n2ID := n2.Host().ID()

	server, cleanup := newTestServer(t, []string{
		n1.ListenAddresses()[0].String(),
		n2.ListenAddresses()[0].String(),
	})
	defer cleanup()
	serverID := server.wakuNode.Host().ID()

	// Expect connect to static nodes.
	test.ExpectPeers(t, server.wakuNode, n1ID, n2ID)

	// Disconnect from static nodes via server.
	test.Disconnect(t, server.wakuNode, n1ID)
	test.Disconnect(t, server.wakuNode, n2ID)

	// Expect reconnect to static nodes.
	test.ExpectPeers(t, server.wakuNode, n1ID, n2ID)

	// Disconnect from static nodes via the nodes.
	test.Disconnect(t, n1, serverID)
	test.Disconnect(t, n2, serverID)

	// Expect reconnect to static nodes.
	test.ExpectPeers(t, server.wakuNode, n1ID, n2ID)
}

func newTestServer(t *testing.T, staticNodes []string) (*Server, func()) {
	_, dbDSN, dbCleanup := test.NewDB(t)
	s := New(context.Background(), Options{
		NodeKey: newNodeKey(t),
		Address: "localhost",
		Store: StoreOptions{
			DbConnectionString: dbDSN,
		},
		StaticNodes: staticNodes,
		WSAddress:   "0.0.0.0",
		WSPort:      0,
		EnableWS:    true,
	})
	require.NotNil(t, s)
	return s, func() {
		s.Shutdown()
		dbCleanup()
	}
}

func newNodeKey(t *testing.T) string {
	key, err := generatePrivateKey()
	require.NoError(t, err)
	return string(key)
}
