package api

import (
	"encoding/base64"
	"time"

	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	envelope "github.com/xmtp/proto/go/message_contents"
	"github.com/xmtp/xmtp-node-go/pkg/crypto"
	"google.golang.org/protobuf/proto"
)

func decodeToken(s string) (*messagev1.Token, error) {
	b, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var token messagev1.Token
	err = proto.Unmarshal(b, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func encodeToken(token *messagev1.Token) (string, error) {
	b, err := proto.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(b), nil
}

func generateToken(createdAt time.Time) (*messagev1.Token, *messagev1.AuthData, error) {
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
	keySig, keyRec, err := crypto.SignDigest(wPri, crypto.EtherHash(createIdentitySignRequest(key)))
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
