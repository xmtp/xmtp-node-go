package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	messageV1 "github.com/xmtp/proto/go/message_api/v1"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
)

func Test_AuthnNoToken(t *testing.T) {
	ctx := context.Background()
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		_, err := client.Publish(ctx, &messageV1.PublishRequest{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "authorization token is not provided")
	})
}

// Private key topic queries must be let through without authn
func Test_AuthnAllowedWithoutAuthn(t *testing.T) {
	ctx := context.Background()
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		_, err := client.Query(ctx, &messageV1.QueryRequest{
			ContentTopics: []string{"privatestore-123"},
		})
		require.NoError(t, err)
	})
}

func Test_AuthnValidToken(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		_, err := client.Publish(ctx, &messageV1.PublishRequest{})
		require.NoError(t, err)
	})
}

func Test_AuthnExpiredToken(t *testing.T) {
	ctx := withExpiredAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		_, err := client.Publish(ctx, &messageV1.PublishRequest{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "token expired")
	})
}
