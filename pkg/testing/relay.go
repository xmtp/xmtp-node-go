package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	wakunode "github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
)

func Subscribe(t *testing.T, n *wakunode.WakuNode) <-chan *protocol.Envelope {
	ctx := context.Background()
	sub, err := n.Relay().Subscribe(ctx)
	require.NoError(t, err)
	return sub.Ch
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
	case env := <-envC:
		require.FailNow(t, fmt.Sprintf("expected no message, got: %v", env.Message()))
	case <-time.After(100 * time.Millisecond):
	}
}
