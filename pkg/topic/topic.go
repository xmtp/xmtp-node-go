package topic

import (
	"fmt"
	"strings"
)

var topicCategoryByPrefix = map[string]string{
	"test":         "test",
	"contact":      "contact",
	"intro":        "v1-intro",
	"dm":           "v1-conversation",
	"dmE":          "v1-conversation-ephemeral",
	"invite":       "v2-invite",
	"groupInvite":  "v2-group-invite",
	"m":            "v2-conversation",
	"mE":           "v2-conversation-ephemeral",
	"privatestore": "private",
}

func IsEphemeral(contentTopic string) bool {
	return Category(contentTopic) == "v2-conversation-ephemeral" || Category(contentTopic) == "v1-conversation-ephemeral"
}

func Category(contentTopic string) string {
	if strings.HasPrefix(contentTopic, "test-") {
		return "test"
	}
	topic := strings.TrimPrefix(contentTopic, "/xmtp/0/")
	if len(topic) == len(contentTopic) {
		return "invalid"
	}
	prefix, _, hasPrefix := strings.Cut(topic, "-")
	if hasPrefix {
		if category, found := topicCategoryByPrefix[prefix]; found {
			return category
		}
	}
	return "invalid"
}

func BuildGroupTopic(groupId string) string {
	return fmt.Sprintf("/xmtp/3/g-%s/proto", groupId)
}

func BuildWelcomeTopic(installationId string) string {
	return fmt.Sprintf("/xmtp/3/w-%s/proto", installationId)
}
