package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messageV1 "github.com/xmtp/proto/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/authn"
)

var authnOptions = Options{
	GRPCPort: 0,
	HTTPPort: 0,
	Authn: authn.Options{
		Enable: true,
	},
}

func Test_AuthnNoToken(t *testing.T) {
	GRPCAndHTTPRunWithOptions(t, authnOptions, func(t *testing.T, client client, server *Server) {
		_, err := client.RawQuery(&messageV1.QueryRequest{
			ContentTopics: []string{"some-random-topic"},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "authorization token is not provided")
	})
}

func Test_AuthnValidToken(t *testing.T) {
	GRPCAndHTTPRunWithOptions(t, authnOptions, func(t *testing.T, client client, server *Server) {
		token, _, err := authn.GenerateToken(time.Now())
		require.NoError(t, err)
		err = client.UseToken(token)
		require.NoError(t, err)
		resp := client.Query(t, &messageV1.QueryRequest{
			ContentTopics: []string{"some-random-topic"},
		})
		require.NotNil(t, resp)
		require.Len(t, resp.Envelopes, 0)
	})
}

func Test_AuthnExpiredToken(t *testing.T) {
	GRPCAndHTTPRunWithOptions(t, authnOptions, func(t *testing.T, client client, server *Server) {
		token, _, err := authn.GenerateToken(time.Now().Add(-24 * time.Hour))
		require.NoError(t, err)
		err = client.UseToken(token)
		require.NoError(t, err)
		_, err = client.RawQuery(&messageV1.QueryRequest{
			ContentTopics: []string{"some-random-topic"},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "token expired")
	})
}
