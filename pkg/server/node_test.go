package server

import (
	"testing"

	wakunode "github.com/status-im/go-waku/waku/v2/node"
	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestNodes_Deployment(t *testing.T) {
	tcs := []struct {
		name string
		test func(t *testing.T, n1, newN1, n2, newN2 *wakunode.WakuNode)
	}{
		{
			name: "new instances connect to new instances",
			test: func(t *testing.T, n1, newN1, n2, newN2 *wakunode.WakuNode) {
				n1ID := n1.Host().ID()
				n2ID := n2.Host().ID()

				// Deploy started; new instances connect to all new instances.
				test.Connect(t, newN1, newN2)
				test.Connect(t, newN2, newN1)

				// Expect new instances are fully connected.
				test.ExpectPeers(t, newN1, n2ID)
				test.ExpectPeers(t, newN2, n1ID)

				// Deploy ending; old instances disconnect from connected peers.
				test.Disconnect(t, n1, n2)

				// Expect new instances are still fully connected.
				test.ExpectPeers(t, newN1, n2ID)
				test.ExpectPeers(t, newN2, n1ID)
			},
		},
		{
			name: "new instances connect to 1 new 1 old instance",
			test: func(t *testing.T, n1, newN1, n2, newN2 *wakunode.WakuNode) {
				n1ID := n1.Host().ID()
				n2ID := n2.Host().ID()

				// Deploy started; new instances connect to some new instances
				// and some old instances.
				test.Connect(t, newN1, newN2)
				test.Connect(t, newN2, n1)

				// Expect new instances are fully connected.
				test.ExpectPeers(t, newN1, n2ID)
				test.ExpectPeers(t, newN2, n1ID)

				// Deploy ending; old instances disconnect from connected peers.
				test.Disconnect(t, n1, n2)

				// Expect new instances are still fully connected.
				test.ExpectPeers(t, newN1, n2ID)
				test.ExpectPeers(t, newN2, n1ID)
			},
		},
		{
			name: "new instances connect to old instances",
			test: func(t *testing.T, n1, newN1, n2, newN2 *wakunode.WakuNode) {
				n1ID := n1.Host().ID()
				n2ID := n2.Host().ID()

				// Deploy started; new instances connect to all old instances.
				test.Connect(t, newN1, n2)
				test.Connect(t, newN2, n1)

				// Expect new instances are fully connected.
				test.ExpectPeers(t, newN1, n2ID)
				test.ExpectPeers(t, newN2, n1ID)

				// Deploy ending; old instances disconnect from connected peers.
				test.Disconnect(t, n1, n2)

				// Expect new instances are fully disconnected.
				test.ExpectNoPeers(t, newN1)
				test.ExpectNoPeers(t, newN2)
			},
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Initialize private keys for each node.
			n1PrivKey := test.NewPrivateKey(t)
			n2PrivKey := test.NewPrivateKey(t)

			// Spin up initial instances of the nodes.
			n1, cleanup := test.NewNode(t, wakunode.WithPrivateKey(n1PrivKey))
			defer cleanup()
			n2, cleanup := test.NewNode(t, wakunode.WithPrivateKey(n2PrivKey))
			defer cleanup()

			// Connect the nodes.
			test.Connect(t, n1, n2)
			test.Connect(t, n2, n1)

			// Spin up new instances of the nodes.
			newN1, cleanup := test.NewNode(t, wakunode.WithPrivateKey(n1PrivKey))
			defer cleanup()
			newN2, cleanup := test.NewNode(t, wakunode.WithPrivateKey(n2PrivKey))
			defer cleanup()

			// Expect matching peer IDs for new and old instances.
			require.Equal(t, n1.Host().ID(), newN1.Host().ID())
			require.Equal(t, n2.Host().ID(), newN2.Host().ID())

			// Run the test case.
			tc.test(t, n1, newN1, n2, newN2)
		})
	}
}
