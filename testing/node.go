package testing

import (
	"context"
	"crypto/ecdsa"
	"net"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/status-im/go-waku/tests"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	wakustore "github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/stretchr/testify/require"
)

func Connect(t *testing.T, n1 *wakunode.WakuNode, n2 *wakunode.WakuNode) {
	ctx := context.Background()
	err := n1.DialPeer(ctx, n2.ListenAddresses()[0].String())
	require.NoError(t, err)

	if len(protocols) > 0 {
		_, err = n1.AddPeer(n2.ListenAddresses()[0], protocols...)
		require.NoError(t, err)
	}

	// This delay is necessary, but it's unclear why at this point. We see
	// similar delays throughout the waku codebase as well for this reason.
	time.Sleep(100 * time.Millisecond)
}

func ConnectWithAddr(t *testing.T, n *wakunode.WakuNode, addr string) {
	ctx := context.Background()
	err := n.DialPeer(ctx, addr)
	require.NoError(t, err)

	// This delay is necessary, but it's unclear why at this point. We see
	// similar delays throughout the waku codebase as well for this reason.
	time.Sleep(100 * time.Millisecond)
}

func Disconnect(t *testing.T, n1 *wakunode.WakuNode, peerID peer.ID) {
	err := n1.ClosePeerById(peerID)
	require.NoError(t, err)
}

func NewTopic() string {
	return "test-" + RandomStringLower(5)
}

func NewNode(t *testing.T, storeNodes []*wakunode.WakuNode, opts ...wakunode.WakuNodeOption) (*wakunode.WakuNode, func()) {
	hostAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:0")

	prvKey := NewPrivateKey(t)

	ctx := context.Background()
	opts = append([]wakunode.WakuNodeOption{
		wakunode.WithPrivateKey(prvKey),
		wakunode.WithHostAddress(hostAddr),
		wakunode.WithWakuRelay(),
		wakunode.WithWakuFilter(true),
		wakunode.WithWebsockets("0.0.0.0", 0),
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

func NewPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	key, err := tests.RandomHex(32)
	require.NoError(t, err)
	prvKey, err := crypto.HexToECDSA(key)
	require.NoError(t, err)
	return prvKey
}

func ExpectPeers(t *testing.T, n *wakunode.WakuNode, expected ...peer.ID) {
	require.Eventually(t, func() bool {
		return len(n.Host().Network().Peers()) == len(expected)
	}, 5*time.Second, 100*time.Millisecond)
	require.ElementsMatch(t, expected, n.Host().Network().Peers())
}

func ExpectNoPeers(t *testing.T, n *wakunode.WakuNode) {
	require.Eventually(t, func() bool {
		return len(n.Host().Network().Peers()) == 0
	}, 1*time.Second, 10*time.Millisecond)
}
