package api

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	v1 "github.com/xmtp/proto/v3/go/message_api/v1"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

const (
	testMaxMsgSize = 2 * 1024 * 1024
)

func newTestServer(t *testing.T) (*Server, func()) {
	log := test.NewLog(t)

	ns, err := server.NewServer(&server.Options{
		Port:       server.RANDOM_PORT,
		MaxPayload: testMaxMsgSize,
	})
	require.NoError(t, err)
	go ns.Start()
	require.True(t, ns.ReadyForConnections(4*time.Second), "nats server not ready")
	nats, err := nats.Connect(ns.ClientURL())
	require.NoError(t, err)
	store, _, dbCleanup := newTestStore(t, log)

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
		NATS:        nats,
		Store:       store,
		Log:         test.NewLog(t),
		AllowLister: allowLister,
	})
	require.NoError(t, err)
	return s, func() {
		s.Close()
		dbCleanup()
		authzDBCleanup()
		ns.Shutdown()
	}
}

func newTestStore(t *testing.T, log *zap.Logger) (*store.Store, func(), func()) {
	db, _, dbCleanup := test.NewDB(t)
	store, err := store.New(
		store.WithLog(log),
		store.WithDB(db),
		store.WithReaderDB(db),
		store.WithCleanerDB(db),
	)
	require.NoError(t, err)

	return store, store.Close, dbCleanup
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

func testGRPC(t *testing.T, ctx context.Context, f func(*testing.T, messageclient.Client, *Server)) {
	t.Parallel()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	server, cleanup := newTestServer(t)
	defer cleanup()

	c, err := messageclient.NewGRPCClient(ctx, server.dialGRPC)
	require.NoError(t, err)

	f(t, c, server)
}

func withAuth(t *testing.T, ctx context.Context) context.Context {
	ctx, _ = withAuthWithDetails(t, ctx, time.Now())
	return ctx
}

func withExpiredAuth(t *testing.T, ctx context.Context) context.Context {
	ctx, _ = withAuthWithDetails(t, ctx, time.Now().Add(-24*time.Hour))
	return ctx
}

func withMissingAuthData(t *testing.T, ctx context.Context) context.Context {
	token, _, err := GenerateToken(time.Now(), false)
	require.NoError(t, err)
	token.AuthDataBytes = nil
	token.AuthDataSignature = nil
	et, err := EncodeToken(token)
	require.NoError(t, err)
	return metadata.AppendToOutgoingContext(ctx, authorizationMetadataKey, "Bearer "+et)
}

// Test possible malicious behavior of the client.
func withMissingIdentityKey(t *testing.T, ctx context.Context) context.Context {
	token, _, err := GenerateToken(time.Now(), false)
	require.NoError(t, err)
	token.IdentityKey = nil
	et, err := EncodeToken(token)
	require.NoError(t, err)
	return metadata.AppendToOutgoingContext(ctx, authorizationMetadataKey, "Bearer "+et)
}

func withAuthWithDetails(t *testing.T, ctx context.Context, when time.Time) (context.Context, *v1.AuthData) {
	token, data, err := GenerateToken(when, false)
	require.NoError(t, err)
	et, err := EncodeToken(token)
	require.NoError(t, err)
	return metadata.AppendToOutgoingContext(ctx, authorizationMetadataKey, "Bearer "+et), data
}
