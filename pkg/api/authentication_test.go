package api

import (
	"context"
	"testing"
	"time"

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

func Test_AuthnTokenMissingAuthData(t *testing.T) {
	ctx := withMissingAuthData(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		_, err := client.Publish(ctx, &messageV1.PublishRequest{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing auth data")
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

func Test_AuthnRestrictedTopicAllowed(t *testing.T) {
	ctx, data := withAuthWithDetails(t, context.Background(), time.Now())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		for _, topic := range []string{
			"contact-" + data.WalletAddr,
			"privatestore-" + data.WalletAddr,
			"dm-xxx-yyy",
			"m-09238409",
			"invite-0x80980980s",
			"intro-0xdsl8d09s8df0s",
		} {
			_, err := client.Publish(ctx, &messageV1.PublishRequest{
				Envelopes: []*messageV1.Envelope{{ContentTopic: topic}},
			})
			require.NoError(t, err)
		}
	})
}

func Test_AuthnRestrictedTopicDisallowed(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		for _, topic := range []string{
			"contact-0x9d0asadf0a9sdf09a0d97f0a97df0a9sdf0a9sd8",
			"privatestore-0x9d0asadf0a9sdf09a0d97f0a97df0a9sdf0a9sd8",
		} {
			_, err := client.Publish(ctx, &messageV1.PublishRequest{
				Envelopes: []*messageV1.Envelope{{ContentTopic: topic}},
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), "publishing to restricted topic")
		}
	})
}
