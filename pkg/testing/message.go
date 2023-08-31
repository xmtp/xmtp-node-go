package testing

import (
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

func NewEnvelope(contentTopic string, timestamp uint64, content string) *messagev1.Envelope {
	return &messagev1.Envelope{
		Message:      []byte(content),
		ContentTopic: contentTopic,
		TimestampNs:  timestamp,
	}
}
