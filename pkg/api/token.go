package api

import (
	"crypto/ecdsa"
	"encoding/base64"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/xmtp/xmtp-node-go/pkg/crypto"
	messagev1 "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	envelope "github.com/xmtp/xmtp-node-go/pkg/proto/message_contents"
	"google.golang.org/protobuf/proto"
)

func decodeAuthToken(s string) (*messagev1.Token, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var token messagev1.Token
	err = proto.Unmarshal(b, &token)
	if err != nil {
		return nil, err
	}
	// Check that IdentityKey pointers are non-nil.
	if token.IdentityKey == nil || token.IdentityKey.Signature == nil {
		return nil, ErrMissingIdentityKey
	}
	return &token, nil
}

func EncodeAuthToken(token *messagev1.Token) (string, error) {
	b, err := proto.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func generateV2AuthToken(createdAt time.Time) (*messagev1.Token, *messagev1.AuthData, error) {
	v2WalletKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	v2IdentityKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	return BuildV2AuthToken(v2WalletKey, v2IdentityKey, createdAt)
}

func BuildV2AuthToken(
	v2WalletKey, v2IdentityKey *ecdsa.PrivateKey,
	createdAt time.Time,
) (*messagev1.Token, *messagev1.AuthData, error) {
	walletPrivateKey := crypto.PrivateKey(ethcrypto.FromECDSA(v2WalletKey))
	walletPublicKey := crypto.PublicKey(ethcrypto.FromECDSAPub(&v2WalletKey.PublicKey))
	identityPrivateKey := crypto.PrivateKey(ethcrypto.FromECDSA(v2IdentityKey))
	identityPublicKey := crypto.PublicKey(ethcrypto.FromECDSAPub(&v2IdentityKey.PublicKey))

	key := &envelope.PublicKey{
		Timestamp: uint64(time.Now().UnixNano()),
		Union: &envelope.PublicKey_Secp256K1Uncompressed_{
			Secp256K1Uncompressed: &envelope.PublicKey_Secp256K1Uncompressed{
				Bytes: identityPublicKey[:],
			},
		},
	}

	// The wallet signs the identity public key.
	keySig, keyRec, err := crypto.SignDigest(
		walletPrivateKey,
		crypto.EtherHash(createIdentitySignRequest(key)),
	)
	key.Signature = &envelope.Signature{
		Union: &envelope.Signature_WalletEcdsaCompact{
			WalletEcdsaCompact: &envelope.Signature_WalletECDSACompact{
				Bytes:    keySig[:],
				Recovery: uint32(keyRec),
			},
		},
	}

	if err != nil {
		return nil, nil, err
	}
	wallet := crypto.PublicKeyToAddress(walletPublicKey)
	data := &messagev1.AuthData{
		WalletAddr: string(wallet),
		CreatedNs:  uint64(createdAt.UnixNano()),
	}
	dataBytes, err := proto.Marshal(data)
	if err != nil {
		return nil, nil, err
	}

	// The identity key signs the auth data.
	sig, rec, err := crypto.Sign(identityPrivateKey, dataBytes)
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
