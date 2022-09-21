package api

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	wakustore "github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/stretchr/testify/require"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"google.golang.org/grpc/metadata"
)

func newTestServer(t *testing.T) (*Server, func()) {
	waku, wakuCleanup := newTestNode(t, nil)
	s, err := New(&Config{
		Options: Options{
			GRPCAddress: "localhost",
			GRPCPort:    0,
			HTTPAddress: "localhost",
			HTTPPort:    0,
			Authn: AuthnOptions{
				Enable: true,
			},
		},
		Waku: waku,
		Log:  test.NewLog(t),
	})
	require.NoError(t, err)
	return s, func() {
		s.Close()
		wakuCleanup()
	}
}

func newTestNode(t *testing.T, storeNodes []*wakunode.WakuNode, opts ...wakunode.WakuNodeOption) (*wakunode.WakuNode, func()) {
	var dbCleanup func()
	n, nodeCleanup := test.NewNode(t, storeNodes,
		append(
			opts,
			wakunode.WithWakuStore(true, false),
			wakunode.WithWakuStoreFactory(func(w *wakunode.WakuNode) wakustore.Store {
				// Note that the node calls store.Stop() during it's cleanup,
				// but it that doesn't clean up the given DB, so we make sure
				// to return that in the node cleanup returned here.
				// Note that the same host needs to be used here.
				var store *store.XmtpStore
				store, _, _, dbCleanup = newTestStore(t, w.Host())
				return store
			}),
		)...,
	)
	return n, func() {
		nodeCleanup()
		dbCleanup()
	}
}

func newTestStore(t *testing.T, host host.Host) (*store.XmtpStore, *store.DBStore, func(), func()) {
	db, _, dbCleanup := test.NewDB(t)
	dbStore, err := store.NewDBStore(utils.Logger(), store.WithDBStoreDB(db))
	require.NoError(t, err)

	if host == nil {
		host = test.NewPeer(t)
	}
	store, err := store.NewXmtpStore(
		store.WithLog(utils.Logger()),
		store.WithHost(host),
		store.WithDB(db),
		store.WithMessageProvider(dbStore))
	require.NoError(t, err)

	store.Start(context.Background())

	return store, dbStore, store.Stop, dbCleanup
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

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		server, cleanup := newTestServer(t)
		defer cleanup()

		client := messageclient.NewHTTPClient(ctx, server.httpListenAddr(), "")
		defer client.Close()
		f(t, client, server)
	})
}

func withAuth(t *testing.T, ctx context.Context) context.Context {
	token, _, err := GenerateToken(time.Now())
	require.NoError(t, err)
	et, err := EncodeToken(token)
	require.NoError(t, err)
	return metadata.AppendToOutgoingContext(ctx, authorizationMetadataKey, "Bearer "+et)
}

func withExpiredAuth(t *testing.T, ctx context.Context) context.Context {
	token, _, err := GenerateToken(time.Now().Add(-24 * time.Hour))
	require.NoError(t, err)
	et, err := EncodeToken(token)
	require.NoError(t, err)
	return metadata.AppendToOutgoingContext(ctx, authorizationMetadataKey, "Bearer "+et)
}
