// Package crypto provides wrappers functions for cryptographic primitives along with various helper functions
package crypto

import (
	"errors"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/xmtp/xmtp-node-go/pkg/types"
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

// GenerateKey generates a Secp256k1 key pair.
func GenerateKeyPair() (pri PrivateKey, pub PublicKey, err error) {
	ecdsaPri, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	pri = PrivateKey(ethcrypto.FromECDSA(ecdsaPri))
	pub = PublicKey(ethcrypto.FromECDSAPub(&ecdsaPri.PublicKey))
	return pri, pub, nil
}

// Verify evaluates a Secp256k1 signature to determine if the message provided was signed by the given publics
// corresponding private key. It returns true if the message was signed by the corresponding keypair, as
//well as any errors generated in the process
func Verify(pub PublicKey, msg Message, sig Signature) bool {
	digest := ethcrypto.Keccak256(msg)
	return secp256k1.VerifySignature((*pub)[:], digest, (*sig)[:])
}

// Sign generates a signature for the unhashed message provided.
func Sign(privateKey PrivateKey, msg Message) (Signature, uint8, error) {
	digest := ethcrypto.Keccak256(msg)
	return SignDigest(privateKey, digest[:])
}

// SignDigest generates an RFC1363 formatted signature for the provided digest.
// It returns a signature in IEEE p1363 Format [R||S],the recovery bit and any error encountered
func SignDigest(privateKey PrivateKey, digest []byte) (Signature, uint8, error) {
	signatureBytes, err := secp256k1.Sign(digest, (*[32]byte)(privateKey)[:])
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

// EtherHash implements the Ethereum hashing standard used to create signatures. Uses an Ethereum specific prefix and
// the Keccak256 Hashing algorithm
func EtherHash(msg Message) []byte {
	decoratedBytes := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg))
	return ethcrypto.Keccak256(decoratedBytes)
}

// RecoverWalletAddress calculates the WalletAddress of the identity which was used to sign the given message. A value
// is always returned and needs to verified before it can be trusted.
func RecoverWalletAddress(msg Message, signature Signature, recovery uint8) (wallet types.WalletAddr, err error) {

	digest := EtherHash(msg)
	signatureBytes := append(bytesFromSig(signature), recovery)

	epk, err := ethcrypto.SigToPub(digest, signatureBytes)
	if err != nil {
		return wallet, err
	}

	addr := ethcrypto.PubkeyToAddress(*epk)
	return types.WalletAddr(addr.String()), nil
}

// PublicKeyToAddress converts a public key to a wallet address
func PublicKeyToAddress(pub PublicKey) types.WalletAddr {
	epk, _ := ethcrypto.UnmarshalPubkey(pub[:])
	addr := ethcrypto.PubkeyToAddress(*epk)
	return types.WalletAddr(addr.String())
}

// PrivateKeyFromBytes converts from a byte slice to a PrivateKey Type
func PrivateKeyFromBytes(bytes []byte) (PrivateKey, error) {
	if len(bytes) != 32 {
		return nil, ErrInvalidSignatureLen
	}

	return PrivateKey(bytes), nil
}

// PublicKeyFromBytes converts from a byte slice to a PublicKey type
func PublicKeyFromBytes(bytes []byte) (PublicKey, error) {
	if len(bytes) != 65 {
		return nil, ErrInvalidPubkey
	}

	return PublicKey(bytes), nil
}

// SignatureFromBytes converts from a byte slice to a Signature type
func SignatureFromBytes(bytes []byte) (Signature, error) {
	if len(bytes) != 64 {
		return nil, ErrInvalidSignatureLen
	}

	return Signature(bytes), nil
}

// bytesFromSig converts from a Signature type to a byte slice
func bytesFromSig(signature Signature) []byte {
	return (*signature)[:]
}
