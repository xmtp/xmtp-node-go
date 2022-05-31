// Package crypto provides wrappers functions for cryptographic primitives along with various helper functions
package crypto

import (
	"crypto/sha256"
	"errors"
	"fmt"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/xmtp/xmtp-node-go/types"
	"go.uber.org/zap"
)

var (
	ErrInvalidSignatureLen          = errors.New("invalid signature length")
	ErrInvalidGeneratedSignatureLen = errors.New("generated signature has incorrect length")
	ErrInvalidPubkey                = errors.New("invalid public key")
)

type Message []byte
type PrivateKey *[32]byte
type PublicKey *[65]byte
type Signature *[64]byte
type SignatureBytes []byte

// Sign generates an RFC1363 formatted signature for the unhashed message provided.
// It returns a signature in IEEE p1363 Format [R||S],the recovery bit and any error encountered
func Sign(privateKey PrivateKey, msg Message) (Signature, uint8, error) {

	digest := sha256.Sum256(msg)

	signatureBytes, err := secp256k1.Sign(digest[:], (*[32]byte)(privateKey)[:])
	if err != nil {
		return nil, 0, err
	}

	signature, err := SignatureFromBytes(signatureBytes[:len(signatureBytes)-1])
	if err != nil {
		return nil, 0, ErrInvalidGeneratedSignatureLen
	}

	recovery := signatureBytes[len(signatureBytes)-1]
	return signature, recovery, nil
}

// Verify evalutes a Secp256k1 signature to determine if the message provided was signed by the given publics
// corresponding private key. It returns true if the message was signed by the corresponding keypair, as
//well as any errors generated in the process
func Verify(pub PublicKey, msg Message, sig Signature) (bool, error) {
	digest := sha256.Sum256(msg)
	isValid := secp256k1.VerifySignature((*[65]byte)(pub)[:], digest[:], (*[64]byte)(sig)[:])

	return isValid, nil
}

// EtherHash implements the Ethereum hashing standard used to create signatures. Uses an Ethereum specific prefix and
// the Keccak256 Hashing algorithm
func EtherHash(msg Message) []byte {
	decoratedBytes := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg))
	return ethcrypto.Keccak256(decoratedBytes)
}

// RecoverWalletAddress calculates the WalletAddress of the identity which was used to sign the given message. A value
// is always returned and needs to verified before it can be trusted.
func RecoverWalletAddress(msg Message, signature Signature, recovery uint8) (types.WalletAddr, error) {

	digest := EtherHash(msg)
	signatureBytes := append(bytesFromSig(signature), recovery)

	epk, err := ethcrypto.SigToPub(digest, signatureBytes)
	if err != nil {
		return types.InvalidWalletAddr(), err
	}

	addr := ethcrypto.PubkeyToAddress(*epk)
	return types.WalletAddr(addr.String()), nil
}

// PrivateKeyFromBytes converts from a byte slice to a PrivateKey Type
func PrivateKeyFromBytes(bytes []byte) (PrivateKey, error) {
	if len(bytes) != 32 {
		return nil, ErrInvalidSignatureLen
	}

	return (*[32]byte)(bytes), nil
}

// PublicKeyFromBytes converts from a byte slice to a PublicKey type
func PublicKeyFromBytes(bytes []byte) (PublicKey, error) {
	log, _ := zap.NewDevelopment()
	if len(bytes) != 65 {
		log.Error("pkLen", zap.Int("Len", len(bytes)))
		return nil, ErrInvalidPubkey
	}

	return (*[65]byte)(bytes), nil
}

// SignatureFromBytes converts from a byte slice to a Signature type
func SignatureFromBytes(bytes []byte) (Signature, error) {
	if len(bytes) != 64 {
		return nil, ErrInvalidSignatureLen
	}

	return (*[64]byte)(bytes), nil
}

// bytesFromSig converts from a Signature type to a byte slice
func bytesFromSig(signature Signature) []byte {
	return (*[64]byte)(signature)[:]
}