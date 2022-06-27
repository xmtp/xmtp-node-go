package testing

import (
	"context"
	"testing"
	"time"

	"github.com/status-im/go-waku/tests"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/stretchr/testify/require"
)

func Subscribe(t *testing.T, n *wakunode.WakuNode) chan *protocol.Envelope {
	ctx := context.Background()
	sub, err := n.Relay().Subscribe(ctx)
	require.NoError(t, err)
	return sub.C
}

func Publish(t *testing.T, n *wakunode.WakuNode, contentTopic string, timestamp int64) {
	ctx := context.Background()
	_, err := n.Relay().Publish(ctx, tests.CreateWakuMessage(contentTopic, timestamp))
	require.NoError(t, err)
}

func SubscribeExpect(t *testing.T, envC chan *protocol.Envelope, msgs []*pb.WakuMessage) {
	receivedMsgs := []*pb.WakuMessage{}
	var done bool
	for !done {
		select {
		case env := <-envC:
			receivedMsgs = append(receivedMsgs, env.Message())
		case <-time.After(500 * time.Millisecond):
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
