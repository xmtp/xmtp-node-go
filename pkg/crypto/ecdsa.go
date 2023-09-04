package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func HexToECDSA(key string) (*ecdsa.PrivateKey, error) {
	if len(key) == 60 {
		key = "0000" + key
	}
	if len(key) == 62 {
		key = "00" + key
	}
	return ethcrypto.HexToECDSA(key)
}

func ECDSAToHex(key *ecdsa.PrivateKey) string {
	return hex.EncodeToString(key.D.Bytes())
}
