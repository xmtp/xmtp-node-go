package crypto

import (
	"crypto/sha256"
	"errors"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/xmtp/xmtp-node-go/types"
	"go.uber.org/zap"
)

var (
	ErrInvalidSignatureLen          = errors.New("invalid signature length")
	ErrInvalidGeneratedSignatureLen = errors.New("generated signature has incorrect length")
	ErrInvalidPubkey                = errors.New("invalid public key")
)

// Sign generates an RFC1363 formatted signature for the unhashed message provided.
// It returns a signature in IEEE p1363 Format [R||S],the recovery bit and any error encountered
func Sign(priv types.PrivateKey, msg types.Message) (types.Signature, uint8, error) {

	digest := sha256.Sum256(msg)

	signatureBytes, err := secp256k1.Sign(digest[:], (*[32]byte)(priv)[:])
	if err != nil {
		return nil, 0, err
	}

	signature, err := BytesToSignature(signatureBytes[:len(signatureBytes)-1])
	if err != nil {
		return nil, 0, ErrInvalidGeneratedSignatureLen
	}

	recovery := signatureBytes[len(signatureBytes)-1]
	return signature, recovery, nil
}

func Verify(pub types.PublicKey, msg types.Message, sig types.Signature) (bool, error) {
	digest := sha256.Sum256(msg)
	isValid := secp256k1.VerifySignature((*[65]byte)(pub)[:], digest[:], (*[64]byte)(sig)[:])

	return isValid, nil
}

func BytesToPrivateKey(bytes []byte) (types.PrivateKey, error) {
	if len(bytes) != 32 {
		return nil, ErrInvalidSignatureLen
	}

	return (*[32]byte)(bytes), nil
}

func BytesToPublicKey(bytes []byte) (types.PublicKey, error) {
	log, _ := zap.NewDevelopment()
	if len(bytes) != 65 {
		log.Error("pkLen", zap.Int("Len", len(bytes)))
		return nil, ErrInvalidPubkey
	}

	return (*[65]byte)(bytes), nil
}

func BytesToSignature(bytes []byte) (types.Signature, error) {
	if len(bytes) != 64 {
		return nil, ErrInvalidSignatureLen
	}

	return (*[64]byte)(bytes), nil
}
