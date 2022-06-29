package testing

import (
	"testing"

	"github.com/status-im/go-waku/tests"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
)

func NewMessage(contentTopic string, timestamp int64) *pb.WakuMessage {
	return tests.CreateWakuMessage(contentTopic, timestamp)
}

func NewEnvelope(t *testing.T, msg *pb.WakuMessage, pubSubTopic string) *protocol.Envelope {
	return protocol.NewEnvelope(msg, msg.Timestamp, pubSubTopic)
}
