package replication

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/replication/registry"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func NewTestServer(t *testing.T, registry registry.NodeRegistry) *Server {
	log := test.NewLog(t)
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	server, err := New(context.Background(), log, Options{
		PrivateKeyString: hex.EncodeToString(crypto.FromECDSA(privateKey)),
		API: ApiOptions{
			Port: 0,
		},
	}, registry)
	require.NoError(t, err)

	return server
}

func TestCreateServer(t *testing.T) {
	registry := registry.NewFixedNodeRegistry([]registry.Node{})
	server1 := NewTestServer(t, registry)
	server2 := NewTestServer(t, registry)
	require.NotEqual(t, server1.Addr(), server2.Addr())
}
