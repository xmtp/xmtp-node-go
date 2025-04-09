package types

import (
	messageContentsProto "github.com/xmtp/xmtp-node-go/pkg/proto/mls/message_contents"
)

type WelcomeCiphersuite int16

// DO NOT MODIFY THE ORDER OF THESE VALUES
// The values get saved in the DB as int16s, so changing the order will change the meaning of existing rows
const (
	CiphersuiteCurve25519 WelcomeCiphersuite = iota
	CiphersuiteXwingMlkem512
)

// WelcomeCiphersuiteFromProto converts the enum from the proto to a WelcomeCiphersuite enum
// Defaults to Curve25519 if the proto is not set (older clients, which only support Curve25519)
func WelcomeCiphersuiteFromProto(proto messageContentsProto.WelcomeEncryption) WelcomeCiphersuite {
	switch proto {
	case messageContentsProto.WelcomeEncryption_WELCOME_ENCRYPTION_CURVE25519:
		return CiphersuiteCurve25519
	case messageContentsProto.WelcomeEncryption_WELCOME_ENCRYPTION_XWING_MLKEM_512:
		return CiphersuiteXwingMlkem512
	}
	return CiphersuiteCurve25519
}

// Get the WelcomeCiphersuite from the proto, defaulting to Curve25519 if the proto is not set
func WelcomeCiphersuiteToProto(ciphersuite WelcomeCiphersuite) messageContentsProto.WelcomeEncryption {
	switch ciphersuite {
	case CiphersuiteCurve25519:
		return messageContentsProto.WelcomeEncryption_WELCOME_ENCRYPTION_CURVE25519
	case CiphersuiteXwingMlkem512:
		return messageContentsProto.WelcomeEncryption_WELCOME_ENCRYPTION_XWING_MLKEM_512
	default:
		return messageContentsProto.WelcomeEncryption_WELCOME_ENCRYPTION_UNSPECIFIED
	}
}
