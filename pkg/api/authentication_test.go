package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	messageV1 "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
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

func Test_AuthnExpiredToken(t *testing.T) {
	ctx := withExpiredAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		_, err := client.Publish(ctx, &messageV1.PublishRequest{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "token expired")
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
