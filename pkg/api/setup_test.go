package api

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	"github.com/xmtp/xmtp-node-go/pkg/crdt"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

const (
	testMaxMsgSize = 2 * 1024 * 1024
)

func newTestServer(t *testing.T) (*Server, func()) {
	log := test.NewLog(t)
	dataDir := path.Join(t.TempDir(), "crdt-data", test.RandomStringLower(13))
	crdtNode, err := crdt.NewNode(context.Background(), log, crdt.Options{
		DataPath: dataDir,
		P2PPort:  0,
	})
	require.NoError(t, err)
	s, err := New(&Config{
		Options: Options{
			GRPCAddress: "localhost",
			GRPCPort:    0,
			HTTPAddress: "localhost",
			HTTPPort:    0,
			MaxMsgSize:  testMaxMsgSize,
		},
		Log:  test.NewLog(t),
		CRDT: crdtNode,
	})
	require.NoError(t, err)
	return s, func() {
		s.Close()
		crdtNode.Close()
	}
}

func testGRPCAndHTTP(t *testing.T, ctx context.Context, f func(*testing.T, messageclient.Client, *Server)) {
	t.Run("grpc", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		server, cleanup := newTestServer(t)
		defer cleanup()

		c, err := messageclient.NewGRPCClient(ctx, server.dialGRPC)
		require.NoError(t, err)

		f(t, c, server)
	})

	t.Run("http", func(t *testing.T) {
		t.Parallel()

		server, cleanup := newTestServer(t)
		defer cleanup()

		log := test.NewLog(t)
		client := messageclient.NewHTTPClient(log, server.httpListenAddr(), "", "")
		defer client.Close()
		f(t, client, server)
	})
}
