package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messageV1 "github.com/xmtp/proto/v3/go/message_api/v1"
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

func Test_AuthnNoTokenNonV0(t *testing.T) {
	ctx := context.Background()
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		_, err := client.Publish(ctx, &messageV1.PublishRequest{
			Envelopes: []*messageV1.Envelope{
				{
					ContentTopic: "/xmtp/0/m-0x1234/proto",
					TimestampNs:  0,
					Message:      []byte{},
				},
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "authorization token is not provided")
	})
}

func Test_AuthnNoTokenMixedV0MLS(t *testing.T) {
	ctx := context.Background()
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		_, err := client.Publish(ctx, &messageV1.PublishRequest{
			Envelopes: []*messageV1.Envelope{
				{
					ContentTopic: "/xmtp/0/m-0x1234/proto",
					TimestampNs:  0,
					Message:      []byte{},
				},
				{
					ContentTopic: "/xmtp/mls/1/m-0x1234/proto",
					TimestampNs:  0,
					Message:      []byte{},
				},
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "authorization token is not provided")
	})
}

// Private key topic queries must be let through without authentication
func Test_AuthnAllowedWithoutAuthn(t *testing.T) {
	ctx := context.Background()
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		_, err := client.Query(ctx, &messageV1.QueryRequest{
			ContentTopics: []string{"privatestore-123"},
		})
		require.NoError(t, err)
	})
}

func Test_AuthnTokenMissingIdentityKey(t *testing.T) {
	ctx := withMissingIdentityKey(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		_, err := client.Publish(ctx, &messageV1.PublishRequest{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing identity key")
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
			"/xmtp/0/contact-" + data.WalletAddr + "/proto",
			"/xmtp/0/privatestore-" + data.WalletAddr + "/key_bundle/proto",
			"/xmtp/0/dm-xxx-yyy/proto",
			"/xmtp/0/m-09238409/proto",
			"/xmtp/0/invite-0x80980980s/proto",
			"/xmtp/0/intro-0xdsl8d09s8df0s/proto",
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
			"/xmtp/0/contact-0x9d0asadf0a9sdf09a0d97f0a97df0a9sdf0a9sd8/proto",
			"/xmtp/0/privatestore-0x9d0asadf0a9sdf09a0d97f0a97df0a9sdf0a9sd8/key_bundle/proto",
		} {
			_, err := client.Publish(ctx, &messageV1.PublishRequest{
				Envelopes: []*messageV1.Envelope{{ContentTopic: topic}},
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), "publishing to restricted topic")
		}
	})
}
