package types

import (
	"fmt"

	messageContentsProto "github.com/xmtp/xmtp-node-go/pkg/proto/mls/message_contents"
)

type WrapperAlgorithm int16

// DO NOT MODIFY THE ORDER OF THESE VALUES
// The values get saved in the DB as int16s, so changing the order will change the meaning of existing rows
const (
	AlgorithmCurve25519 WrapperAlgorithm = iota
	AlgorithmXwingMlkem768Draft6
	AlgorithmSymmetricKey
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
	case messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_SYMMETRIC_KEY:
		return AlgorithmSymmetricKey
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
	case AlgorithmSymmetricKey:
		return messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_SYMMETRIC_KEY
	default:
		return messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519
	}
}

// WelcomePointerWrapperAlgorithmFromProto converts the proto enum to a WrapperAlgorithm.
// Returns an error if the proto is not set (UNSPECIFIED) or is unknown. The returned
// algorithm value is undefined when err != nil and must be ignored.
func WelcomePointerWrapperAlgorithmFromProto(
	proto messageContentsProto.WelcomePointerWrapperAlgorithm,
) (WrapperAlgorithm, error) {
	switch proto {
	case messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6:
		return AlgorithmXwingMlkem768Draft6, nil
	default:
		err := fmt.Errorf("invalid welcome pointer wrapper algorithm %v", proto)
		return AlgorithmXwingMlkem768Draft6, err
	}
}

// WrapperAlgorithmToWelcomePointerWrapperAlgorithm converts a WrapperAlgorithm to its proto representation
func WrapperAlgorithmToWelcomePointerWrapperAlgorithm(
	algorithm WrapperAlgorithm,
) messageContentsProto.WelcomePointerWrapperAlgorithm {
	switch algorithm {
	case AlgorithmXwingMlkem768Draft6:
		return messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6
	default:
		return messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_UNSPECIFIED
	}
}
