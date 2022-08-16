package authn

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	envelope "github.com/xmtp/proto/go/message_contents"
	"github.com/xmtp/xmtp-node-go/pkg/crypto"
	"github.com/xmtp/xmtp-node-go/pkg/types"
	"google.golang.org/protobuf/proto"
)

var (
	TokenExpiration = time.Hour

	ErrTokenExpired         = errors.New("token expired")
	ErrFutureToken          = errors.New("token timestamp is in the future")
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrWalletMismatch       = errors.New("wallet address mismatch")
	ErrUnsignedKey          = errors.New("identity key is not signed")
	ErrUnknownSignatureType = errors.New("unknown signature type")
	ErrUnknownKeyType       = errors.New("unknown public key type")
)

func validateToken(ctx context.Context, token *messagev1.Token, now time.Time) (wallet types.WalletAddr, err error) {
	// Validate IdentityKey signature.
	pubKey := token.IdentityKey
	recoveredWalletAddress, err := recoverWalletAddress(ctx, pubKey)
	if err != nil {
		return wallet, err
	}

	// Validate AuthData signature.
	data, err := verifyAuthSignature(ctx, pubKey, token.AuthDataBytes, token.AuthDataSignature)
	if err != nil {
		return wallet, err
	}
	suppliedWalletAddress := types.WalletAddr(data.WalletAddr)

	// To protect against spoofing, ensure the IdentityKey wallet address matches the AuthData wallet address.
	if recoveredWalletAddress != suppliedWalletAddress {
		return wallet, ErrWalletMismatch
	}

	// Check expiration
	created := time.Unix(0, int64(data.CreatedNs))
	if created.After(now) {
		return wallet, ErrFutureToken
	}
	if now.Sub(created) > TokenExpiration {
		return wallet, ErrTokenExpired
	}

	return recoveredWalletAddress, nil
}

//
func createIdentitySignRequest(identityKey *envelope.PublicKey) crypto.Message {
	// We need a bare key to generate the key bytes to sign.
	unsignedKey := &envelope.PublicKey{
		Timestamp: identityKey.Timestamp,
		Union:     identityKey.Union,
	}
	identityBytes, _ := proto.Marshal(unsignedKey)
	return []byte(fmt.Sprintf(
		"XMTP : Create Identity\n%s\n\nFor more info: https://xmtp.org/signatures/",
		hex.EncodeToString(identityBytes),
	))
}

func recoverWalletAddress(ctx context.Context, identityKey *envelope.PublicKey) (wallet types.WalletAddr, err error) {
	isrBytes := createIdentitySignRequest(identityKey)
	signature := identityKey.Signature
	if signature == nil {
		return "", ErrUnsignedKey
	}
	switch sig := signature.Union.(type) {
	case *envelope.Signature_EcdsaCompact:
		cSig, err := crypto.SignatureFromBytes(sig.EcdsaCompact.Bytes)
		if err != nil {
			return wallet, err
		}
		return crypto.RecoverWalletAddress(isrBytes, cSig, uint8(sig.EcdsaCompact.Recovery))
	default:
		return "", ErrUnknownSignatureType
	}
}

func verifyAuthSignature(ctx context.Context, pubKey *envelope.PublicKey, authDataBytes []byte, authSig *envelope.Signature) (data *messagev1.AuthData, err error) {
	switch key := pubKey.Union.(type) {
	case *envelope.PublicKey_Secp256K1Uncompressed_:
		pub, err := crypto.PublicKeyFromBytes(key.Secp256K1Uncompressed.Bytes)
		if err != nil {
			return nil, err
		}

		switch sig := authSig.Union.(type) {
		case *envelope.Signature_EcdsaCompact:
			signature, err := crypto.SignatureFromBytes(sig.EcdsaCompact.Bytes)
			if err != nil {
				return nil, err
			}

			isVerified := crypto.Verify(pub, authDataBytes, signature)
			if !isVerified {
				return nil, ErrInvalidSignature
			}

			var authData messagev1.AuthData
			err = proto.Unmarshal(authDataBytes, &authData)
			if err != nil {
				return nil, err
			}

			return &authData, nil
		default:
			return nil, ErrUnknownSignatureType
		}
	default:
		return nil, ErrUnknownKeyType
	}
}
