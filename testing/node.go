package testing

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/status-im/go-waku/tests"
	"github.com/status-im/go-waku/waku/v2/node"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	wakustore "github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/stretchr/testify/require"
)

func Connect(t *testing.T, n1 *wakunode.WakuNode, n2 *wakunode.WakuNode, protocols ...string) {
	ctx := context.Background()
	err := n1.DialPeer(ctx, n2.ListenAddresses()[0].String())
	require.NoError(t, err)

	// This delay is necessary, but it's unclear why at this point. We see
	// similar delays throughout the waku codebase as well for this reason.
	time.Sleep(100 * time.Millisecond)
}

func Disconnect(t *testing.T, n1 *wakunode.WakuNode, n2 *wakunode.WakuNode) {
	err := n1.ClosePeerById(n2.Host().ID())
	require.NoError(t, err)
}

func NewTopic() string {
	return "test-" + RandomStringLower(5)
}

func NewNode(t *testing.T, storeNodes []*wakunode.WakuNode, opts ...node.WakuNodeOption) (*wakunode.WakuNode, func()) {
	hostAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:0")

	key, err := tests.RandomHex(32)
	require.NoError(t, err)
	prvKey, err := crypto.HexToECDSA(key)
	require.NoError(t, err)

	ctx := context.Background()
	opts = append([]node.WakuNodeOption{
		wakunode.WithPrivateKey(prvKey),
		wakunode.WithHostAddress(hostAddr),
		wakunode.WithWakuRelay(),
	}, opts...)
	node, err := wakunode.New(ctx, opts...)
	require.NoError(t, err)

	// Connect to store nodes before starting, similar to what happens in the
	// main entrypoint in server.go.
	for _, storeNode := range storeNodes {
		_, err := node.AddPeer(storeNode.ListenAddresses()[0], string(wakustore.StoreID_v20beta4))
		require.NoError(t, err)
	}

	err = node.Start()
	require.NoError(t, err)

	return node, func() {
		node.Stop()
	}
}

func NewPeer(t *testing.T) host.Host {
	host, err := libp2p.New(libp2p.DefaultTransports, libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	require.NoError(t, err)
	return host
}
