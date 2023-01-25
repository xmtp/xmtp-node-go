package api

import (
	"encoding/base64"
	"errors"
	"time"

	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	envelope "github.com/xmtp/proto/v3/go/message_contents"
	"github.com/xmtp/xmtp-node-go/pkg/crypto"
	"google.golang.org/protobuf/proto"
)

func decodeToken(s string) (*messagev1.Token, error) {
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
		return nil, errors.New("missing identity key")
	}
	return &token, nil
}

func EncodeToken(token *messagev1.Token) (string, error) {
	b, err := proto.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func GenerateToken(createdAt time.Time, v1 bool) (*messagev1.Token, *messagev1.AuthData, error) {
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
	// The wallet signs the identity public key.
	keySig, keyRec, err := crypto.SignDigest(wPri, crypto.EtherHash(createIdentitySignRequest(key)))
	if v1 {
		// Legacy clients package the identity key signature as .EcdsaCompact.
		key.Signature = &envelope.Signature{
			Union: &envelope.Signature_EcdsaCompact{
				EcdsaCompact: &envelope.Signature_ECDSACompact{
					Bytes:    keySig[:],
					Recovery: uint32(keyRec),
				},
			},
		}
	} else {
		key.Signature = &envelope.Signature{
			Union: &envelope.Signature_WalletEcdsaCompact{
				WalletEcdsaCompact: &envelope.Signature_WalletECDSACompact{
					Bytes:    keySig[:],
					Recovery: uint32(keyRec),
				},
			},
		}
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

	// The identity key signs the auth data.
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
