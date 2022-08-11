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
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	TokenExpiration = time.Hour

	ErrTokenExpired      = errors.New("token expired")
	ErrFutureToken       = errors.New("token timestamp is in the future")
	ErrInvalidWalletAddr = errors.New("invalid walletAddress")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrWalletMismatch    = errors.New("wallet address mismatch")
)

func validateToken(ctx context.Context, token *messagev1.Token, now time.Time) (wallet types.WalletAddr, err error) {
	logger := logging.From(ctx)

	// Validate WalletSignature
	pubKey := token.IdentityKey
	recoveredWalletAddress, err := recoverWalletAddress(ctx, pubKey)
	if err != nil {
		return wallet, err
	}

	// Validate AuthSignature
	data, err := verifyAuthSignature(ctx, pubKey, token.AuthDataBytes, token.AuthDataSignature.GetEcdsaCompact())
	if err != nil {
		return wallet, err
	}
	suppliedWalletAddress := types.WalletAddr(data.WalletAddr)

	// To protect against spoofing, ensure the walletAddresses match in both signatures
	if recoveredWalletAddress != suppliedWalletAddress {
		logger.Error("wallet address mismatch", zap.Error(err), logging.WalletAddressLabelled("recovered", recoveredWalletAddress), logging.WalletAddressLabelled("supplied", suppliedWalletAddress))
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

func CreateIdentitySignRequest(identityKey *envelope.PublicKey) crypto.Message {
	// We need a bare key to generate the key bytes to sign.
	// Make a copy of the key so that we can unset the signature.
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
	isrBytes := CreateIdentitySignRequest(identityKey)
	signature := identityKey.GetSignature().GetEcdsaCompact()
	sig, err := crypto.SignatureFromBytes(signature.GetBytes())
	if err != nil {
		return wallet, err
	}
	return crypto.RecoverWalletAddress(isrBytes, sig, uint8(signature.GetRecovery()))
}

func verifyAuthSignature(ctx context.Context, pubKey *envelope.PublicKey, authDataBytes []byte, authSig *envelope.Signature_ECDSACompact) (data *messagev1.AuthData, err error) {

	pub, err := crypto.PublicKeyFromBytes(pubKey.GetSecp256K1Uncompressed().Bytes)
	if err != nil {
		return nil, err
	}

	signature, err := crypto.SignatureFromBytes(authSig.GetBytes())
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
}

func GenerateToken(createdAt time.Time) (*messagev1.Token, *messagev1.AuthData, error) {
	wPri, wPub, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}
	iPri, iPub, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}
	key := &envelope.PublicKey{
		Timestamp: uint64(time.Now().UnixNano()),
		Union: &envelope.PublicKey_Secp256K1Uncompressed_{
			Secp256K1Uncompressed: &envelope.PublicKey_Secp256K1Uncompressed{
				Bytes: iPub[:],
			},
		},
	}
	keySig, keyRec, err := crypto.SignDigest(wPri, crypto.EtherHash(CreateIdentitySignRequest(key)))
	key.Signature = &envelope.Signature{
		Union: &envelope.Signature_EcdsaCompact{
			EcdsaCompact: &envelope.Signature_ECDSACompact{
				Bytes:    keySig[:],
				Recovery: uint32(keyRec),
			},
		},
	}
	if err != nil {
		return nil, nil, err
	}
	wallet := crypto.PublicKeyToAddress(wPub)
	data := &messagev1.AuthData{
		WalletAddr: string(wallet),
		CreatedNs:  uint64(createdAt.UnixNano()),
	}
	dataBytes, err := proto.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	sig, rec, err := crypto.Sign(iPri, dataBytes)
	if err != nil {
		return nil, nil, err
	}
	return &messagev1.Token{
		IdentityKey:   key,
		AuthDataBytes: dataBytes,
		AuthDataSignature: &envelope.Signature{
			Union: &envelope.Signature_EcdsaCompact{
				EcdsaCompact: &envelope.Signature_ECDSACompact{
					Bytes:    sig[:],
					Recovery: uint32(rec),
				},
			},
		},
	}, data, nil
}
