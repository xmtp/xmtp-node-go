package api

import (
	"context"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	v1 "github.com/xmtp/proto/v3/go/message_api/v1"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"google.golang.org/grpc/metadata"
)

const (
	testMaxMsgSize = 2 * 1024 * 1024
)

func newTestServer(t *testing.T) (*Server, func()) {
	log := test.NewLog(t)
	nats, natsCleanup := newTestNATS(t)
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
		Log:         test.NewLog(t),
		AllowLister: allowLister,
	})
	require.NoError(t, err)
	return s, func() {
		s.Close()
		natsCleanup()
		authzDBCleanup()
	}
}

func natsWaitConnected(t *testing.T, c *nats.Conn) {
	t.Helper()

	timeout := time.Now().Add(2 * time.Second)
	for time.Now().Before(timeout) {
		if c.IsConnected() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("client connecting timeout")
}

func newTestNATS(t *testing.T) (*nats.Conn, func()) {
	ns := natsserver.New(&natsserver.Options{
		Host:           "127.0.0.1",
		Port:           natsserver.RANDOM_PORT,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 2048,
	})
	go natsserver.Run(ns)
	if !ns.ReadyForConnections(2 * time.Second) {
		t.Fatal("starting nats server: timeout")
	}
	nc, err := nats.Connect("nats://" + ns.Addr().String())
	require.NoError(t, err)
	natsWaitConnected(t, nc)
	return nc, func() {
		ns.Shutdown()
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
