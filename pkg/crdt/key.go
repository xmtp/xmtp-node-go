package crdt

import (
	"encoding/hex"

	"github.com/libp2p/go-libp2p/core/crypto"
)

func GenerateNodeKey() (crypto.PrivKey, error) {
	privKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	return privKey, err
}

func HexEncodeNodeKey(key crypto.PrivKey) (string, error) {
	keyBytes, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(keyBytes), nil
}
