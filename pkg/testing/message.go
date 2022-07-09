package testing

import (
	"testing"

	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/utils"
)

func NewMessage(contentTopic string, timestamp int64, content string) *pb.WakuMessage {
	return &pb.WakuMessage{
		Payload:      []byte(content),
		ContentTopic: contentTopic,
		Timestamp:    timestamp,
	}
}

func NewEnvelope(t *testing.T, msg *pb.WakuMessage, pubSubTopic string) *protocol.Envelope {
	return protocol.NewEnvelope(msg, utils.GetUnixEpoch(), pubSubTopic)
}
