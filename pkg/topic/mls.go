package topic

import (
	"fmt"
	"strings"
)

const mlsv1Prefix = "/xmtp/mls/1/"

var (
	AssociationChangedTopic = BuildAssociationChangedTopic()
)

func IsMLSV1(topic string) bool {
	return strings.HasPrefix(topic, mlsv1Prefix)
}

func IsMLSV1Group(topic string) bool {
	return strings.HasPrefix(topic, mlsv1Prefix+"g-")
}

func IsMLSV1Welcome(topic string) bool {
	return strings.HasPrefix(topic, mlsv1Prefix+"w-")
}

func IsAssociationChanged(topic string) bool {
	return topic == AssociationChangedTopic
}

func BuildMLSV1GroupTopic(groupId []byte) string {
	return fmt.Sprintf("%sg-%x/proto", mlsv1Prefix, groupId)
}

func BuildMLSV1WelcomeTopic(installationId []byte) string {
	return fmt.Sprintf("%sw-%x/proto", mlsv1Prefix, installationId)
}

func BuildAssociationChangedTopic() string {
	return fmt.Sprintf("%sassociations/proto", mlsv1Prefix)
}
