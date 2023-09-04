package crypto_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/crypto"
	"github.com/xmtp/xmtp-node-go/pkg/types"
)

// Tests that signature verification works for a given message and keys.
func TestStaticSignatureRoundTrip(t *testing.T) {

	bPK, _ := hex.DecodeString("6a52887e81142f32dbae00d9ea666484b3de72b859805bbe4694337a63b6ca7c")
	bpk, _ := hex.DecodeString("0497a556a06d5270300967b2d64ae2997af9efe872f8d146c155b91f6bc2315cf6a941a7ea80bb84edea2ffff5637b4f736e2aa64cfb98d6276e168dd1e7cdfc6d")
	msg := []byte("TestPeerID|0x12345")

	PK, _ := crypto.PrivateKeyFromBytes(bPK)
	pk, _ := crypto.PublicKeyFromBytes(bpk)

	generatedSig, recovery, err := crypto.Sign(PK, msg)
	require.NoError(t, err)
	require.True(t, recovery == 0 || recovery == 1, "bad recovery code")

	isValid := crypto.Verify(pk, msg, generatedSig)
	require.True(t, isValid, "Signature validation failed")
}

func TestStaticWalletVerify(t *testing.T) {
	bMsg, _ := hex.DecodeString("584d5450203a20437265617465204964656e746974790a30386234383862366238393133303161343330613431303466393863343937346435343433623538303231626430363233643866663532336564643533616666613532386230353130373561373231393563666435363132626263323737623466323935333561623336663335393565386339356631373830646437646563643731383133373534356237373835396663373338333664380a0a466f72206d6f726520696e666f3a2068747470733a2f2f786d74702e6f72672f7369676e6174757265732f")
	bSig, _ := hex.DecodeString("b6c023b2f93db3f51c392f8b9019ff2a4f19b30cac6b61f8356f027431332173043c8e0553d87740745a953d437d64a747865c4c28938d3fbbe10f961fd05b8f")
	expectedAddr := types.WalletAddr("0x9727188932c3f9a218e8Fc9D8744b1B8b751Abfc")
	recovery := uint8(0)
	sig, _ := crypto.SignatureFromBytes(bSig)

	walletAddr, err := crypto.RecoverWalletAddress(bMsg, sig, recovery)
	require.NoError(t, err)
	require.Equal(t, expectedAddr, walletAddr)
}
