package testing

import (
	"context"
	"testing"
	"time"

	wakunode "github.com/status-im/go-waku/waku/v2/node"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/filter"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/stretchr/testify/require"
)

func SubscribeTo(t *testing.T, n *wakunode.WakuNode, contentTopics []string) chan *protocol.Envelope {
	ctx := context.Background()
	_, f, err := n.Filter().Subscribe(ctx, filter.ContentFilter{
		ContentTopics: contentTopics,
	})
	require.NoError(t, err)
	return f.Chan
}

func Subscribe(t *testing.T, n *wakunode.WakuNode) chan *protocol.Envelope {
	ctx := context.Background()
	sub, err := n.Relay().Subscribe(ctx)
	require.NoError(t, err)
	return sub.C
}

func Publish(t *testing.T, n *wakunode.WakuNode, msg *pb.WakuMessage) {
	ctx := context.Background()
	_, err := n.Relay().Publish(ctx, msg)
	require.NoError(t, err)
}

func SubscribeExpect(t *testing.T, envC chan *protocol.Envelope, msgs []*pb.WakuMessage) {
	receivedMsgs := []*pb.WakuMessage{}
	var done bool
	for !done {
		select {
		case env := <-envC:
			receivedMsgs = append(receivedMsgs, env.Message())
			if len(receivedMsgs) == len(msgs) {
				done = true
			}
		case <-time.After(5 * time.Second):
			done = true
		}
	}
	require.ElementsMatch(t, msgs, receivedMsgs)
}

func SubscribeExpectNone(t *testing.T, envC chan *protocol.Envelope) {
	select {
	case <-envC:
		require.FailNow(t, "expected no message")
	case <-time.After(100 * time.Millisecond):
	}
}
