package types

import (
	messageContentsProto "github.com/xmtp/xmtp-node-go/pkg/proto/mls/message_contents"
)

type WrapperAlgorithm int16

// DO NOT MODIFY THE ORDER OF THESE VALUES
// The values get saved in the DB as int16s, so changing the order will change the meaning of existing rows
const (
	AlgorithmCurve25519 WrapperAlgorithm = iota
	AlgorithmXwingMlkem768Draft6
)

// WelcomeWrapperAlgorithm converts the enum from the proto to a WrapperAlgorithm enum
// Defaults to Curve25519 if the proto is not set (older clients, which only support Curve25519)
func WrapperAlgorithmFromProto(
	proto messageContentsProto.WelcomeWrapperAlgorithm,
) WrapperAlgorithm {
	switch proto {
	case messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519:
		return AlgorithmCurve25519
	case messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6:
		return AlgorithmXwingMlkem768Draft6
	}
	return AlgorithmCurve25519
}

// Convert a WrapperAlgorithm to its proto representation
func WrapperAlgorithmToProto(
	algorithm WrapperAlgorithm,
) messageContentsProto.WelcomeWrapperAlgorithm {
	switch algorithm {
	case AlgorithmCurve25519:
		return messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519
	case AlgorithmXwingMlkem768Draft6:
		return messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6
	default:
		return messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519
	}
}
