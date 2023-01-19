package crdt2

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"go.uber.org/zap"
)

func NewNetwork(t *testing.T, ctx context.Context, nodes, topics int) (list []*Node) {
	log := test.NewLog(t)
	ps := NewChanBroadcaster(log)
	for i := 0; i < nodes; i++ {
		name := fmt.Sprintf("n%d", i)
		n := NewNode(ctx,
			log.Named(name),
			NewMapStore(),
			(*nilSyncer)(nil),
			ps)
		ps.AddNode(n)
		for j := 0; j < topics; j++ {
			topic := fmt.Sprintf("t%d", j)
			n.NewTopic(topic)
			log.Debug("creating", zap.String("node", name), zap.String("topic", topic))
		}
		require.Len(t, n.Topics, topics)
		list = append(list, n)
	}
	require.Len(t, list, nodes)
	return list
}
