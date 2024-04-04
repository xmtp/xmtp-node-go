package envelopes

import (
	"fmt"
	"hash/fnv"

	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
)

func fromWakuTimestamp(ts int64) uint64 {
	if ts < 0 {
		return 0
	}
	return uint64(ts)
}

func BuildNatsSubject(topic string) string {
	hasher := fnv.New64a()
	hasher.Write([]byte(topic))
	return fmt.Sprintf("%x", hasher.Sum64())
}

func BuildEnvelope(msg *wakupb.WakuMessage) *proto.Envelope {
	return &proto.Envelope{
		ContentTopic: msg.ContentTopic,
		TimestampNs:  fromWakuTimestamp(msg.Timestamp),
		Message:      msg.Payload,
	}
}
