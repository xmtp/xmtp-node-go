package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
)

func TestCategoryFromTopic(t *testing.T) {
	for topicString, category := range map[string]string{
		"test-092340239":          "test",
		"invite-0x90xc9":          "invalid",
		"invite":                  "invalid",
		"m-9093243984":            "invalid",
		"dm-joe-fred":             "invalid",
		"intro-3230948230":        "invalid",
		"contact-23493840392":     "invalid",
		"privatestore-3920984234": "invalid",
		"/xmtp/0/contact-0x43A3bD1a8a4cfF0aC27bFeda477774542f2B6A2b/proto":                                       "contact",
		"/xmtp/0/privatestore-0x5AD7052d32880a11B654Eb7749429248db42958A/key_bundle/proto":                       "private",
		"/xmtp/0/invite-0xC141028c42eb7871a4FF6BbC22E6AE8d05e2810c/proto":                                        "v2-invite",
		"/xmtp/0/dm-0xAa5CF238A4e000e0A28e6d6Ac2767DD428Bf030B-0xAa5CF238A4e000e0A28e6d6Ac2767DD428Bf030B/proto": "v1-conversation",
		"/xmtp/0/m-TRA9YXFCT8oCx6iRAGkBfA6komD9FLv555w0hJcGIuI/proto":                                            "v2-conversation",
		"/xmtp/0/intro-0xAa5CF238A4e000e0A28e6d6Ac2767DD428Bf030B/proto":                                         "v1-intro",
	} {
		result := topic.Category(topicString)
		assert.Equal(t, category, result, topicString)
	}
}
