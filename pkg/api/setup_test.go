package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"go.uber.org/zap"
)

const (
	testMaxMsgSize = 2 * 1024 * 1024
)

func newTestServerWithLog(t testing.TB, log *zap.Logger) (*Server, func()) {
	waku, wakuCleanup := test.NewNode(t, log)
	store, storeCleanup := newTestStore(t, log)
	authzDB, _, authzDBCleanup := test.NewAuthzDB(t)
	allowLister := authz.NewDatabaseWalletAllowLister(authzDB, log)
	s, err := New(&Config{
		Options: Options{
			GRPCAddress: "localhost",
			GRPCPort:    0,
			HTTPAddress: "localhost",
			HTTPPort:    0,
			Authn: AuthnOptions{
				Enable:     true,
				AllowLists: true,
			},
			MaxMsgSize: testMaxMsgSize,
		},
		Waku:        waku,
		Log:         log,
		Store:       store,
		AllowLister: allowLister,
	})
	require.NoError(t, err)
	return s, func() {
		s.Close()
		wakuCleanup()
		authzDBCleanup()
		storeCleanup()
	}
}

func newTestServer(t testing.TB) (*Server, func()) {
	log := test.NewLog(t)
	return newTestServerWithLog(t, log)
}

func newTestStore(t testing.TB, log *zap.Logger) (*store.Store, func()) {
	db, _, dbCleanup := test.NewDB(t)
	store, err := store.New(&store.Config{
		Log:       log,
		DB:        db,
		ReaderDB:  db,
		CleanerDB: db,
	})
	require.NoError(t, err)

	return store, func() {
		store.Close()
		dbCleanup()
	}
}

func testGRPCAndHTTP(
	t *testing.T,
	ctx context.Context,
	f func(*testing.T, messageclient.Client, *Server),
) {
	t.Run("grpc", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(ctx)
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
		defer func() {
			_ = client.Close()
		}()
		f(t, client, server)
	})
}
