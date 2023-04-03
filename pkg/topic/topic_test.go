package topic_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
)

func Test_EphemeralTopic(t *testing.T) {
	assert.Equal(t, true, topic.IsEphemeral("/xmtp/0/mE-123/proto"))
	assert.Equal(t, false, topic.IsEphemeral("/xmtp/0/m-123/proto"))
}
