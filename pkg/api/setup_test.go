package api

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	v1 "github.com/xmtp/proto/v3/go/message_api/v1"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"github.com/xmtp/xmtp-node-go/pkg/crdt"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"google.golang.org/grpc/metadata"
)

const (
	testMaxMsgSize = 2 * 1024 * 1024
)

func newTestServer(t *testing.T) (*Server, func()) {
	log := test.NewLog(t)
	authzDB, _, authzDBCleanup := test.NewAuthzDB(t)
	allowLister := authz.NewDatabaseWalletAllowLister(authzDB, log)
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
			Authn: AuthnOptions{
				Enable:     true,
				AllowLists: true,
			},
			MaxMsgSize: testMaxMsgSize,
		},
		Log:         test.NewLog(t),
		AllowLister: allowLister,
		CRDT:        crdtNode,
	})
	require.NoError(t, err)
	return s, func() {
		s.Close()
		authzDBCleanup()
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

func withAuthWithDetails(t *testing.T, ctx context.Context, when time.Time) (context.Context, *v1.AuthData) {
	token, data, err := GenerateToken(when, false)
	require.NoError(t, err)
	et, err := EncodeToken(token)
	require.NoError(t, err)
	return metadata.AppendToOutgoingContext(ctx, authorizationMetadataKey, "Bearer "+et), data
}
