package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messageV1 "github.com/xmtp/proto/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/authn"
)

var authnEnabled = Options{
	GRPCPort: 0,
	HTTPPort: 0,
	Authn: authn.Options{
		Enable: true,
	},
}

func Test_AuthnNoToken(t *testing.T) {
	GRPCAndHTTPRunWithOptions(t, authnEnabled, func(t *testing.T, client client, server *Server) {
		_, err := client.RawPublish(&messageV1.PublishRequest{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "authorization token is not provided")
	})
}

// Private key topic queries must be let through without authn
func Test_AuthnAllowedWithoutAuthn(t *testing.T) {
	GRPCAndHTTPRunWithOptions(t, authnEnabled, func(t *testing.T, client client, server *Server) {
		_, err := client.RawQuery(&messageV1.QueryRequest{
			ContentTopics: []string{"privatestore-123"},
		})
		require.NoError(t, err)
	})
}

func Test_AuthnValidToken(t *testing.T) {
	GRPCAndHTTPRunWithOptions(t, authnEnabled, func(t *testing.T, client client, server *Server) {
		token, _, err := authn.GenerateToken(time.Now())
		require.NoError(t, err)
		err = client.UseToken(token)
		require.NoError(t, err)
		_, err = client.RawPublish(&messageV1.PublishRequest{})
		require.NoError(t, err)
	})
}

func Test_AuthnExpiredToken(t *testing.T) {
	GRPCAndHTTPRunWithOptions(t, authnEnabled, func(t *testing.T, client client, server *Server) {
		token, _, err := authn.GenerateToken(time.Now().Add(-24 * time.Hour))
		require.NoError(t, err)
		err = client.UseToken(token)
		require.NoError(t, err)
		_, err = client.RawPublish(&messageV1.PublishRequest{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "token expired")
	})
}
