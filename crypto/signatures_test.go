package crypto

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func TestStaticSignatureRoundTrip(t *testing.T) {

	bPK, _ := hex.DecodeString("6a52887e81142f32dbae00d9ea666484b3de72b859805bbe4694337a63b6ca7c")
	bpk, _ := hex.DecodeString("0497a556a06d5270300967b2d64ae2997af9efe872f8d146c155b91f6bc2315cf6a941a7ea80bb84edea2ffff5637b4f736e2aa64cfb98d6276e168dd1e7cdfc6d")
	bSig, _ := hex.DecodeString("89ee44df0c282d5caafdd527ef314f2539d813304766daea4fab558e5fafbaf644be1943de60e7e9289567cb505041786151f950601f0cbfc4b9cbefa048dbaa")
	msg := []byte("TestPeerID|0x12345")
	expectedRecovery := uint8(0)

	PK, _ := BytesToPrivateKey(bPK)
	pk, _ := BytesToPublicKey(bpk)

	expectedSig, _ := BytesToSignature(bSig)

	generatedSig, recovery, err := Sign(PK, msg)
	require.NoError(t, err)

	require.Equal(t, expectedSig, generatedSig, "signature mismatch")
	require.Equal(t, expectedRecovery, recovery, "bad recovery code")

	isValid, err := Verify(pk, msg, generatedSig)
	require.NoError(t, err)

	require.True(t, isValid, "Signature could not verified")
}

func walletSignatureMessage() []byte {
	b, _ := hex.DecodeString("584d5450203a20437265617465204964656e746974790a30386332633938336633386433303161343330613431303432343437396563376261626230663235646633393133313665336562623533653566626234306633626132336135393137666437613533313636613834373130613132323464616433613362323364353337333262613838303332636337623661666238396630386236616333613463373538316264663661656466613230350a0a466f72206d6f726520696e666f3a2068747470733a2f2f786d74702e6f72672f7369676e6174757265732f")
	return b
}

func walletSignatureTest() []byte {
	return []byte("XMTP : Create Identity\n%\n\nFor more info: https://xmtp.org/signatures/")
}

func TestStaticWalletVerify(t *testing.T) {

	log, _ := zap.NewDevelopment()
	//walletAddr, _ := hex.DecodeString("0xAfAE0373Fb927635e5d87aC5Df231780822a0310")
	//bpk, _ := hex.DecodeString("0497a556a06d5270300967b2d64ae2997af9efe872f8d146c155b91f6bc2315cf6a941a7ea80bb84edea2ffff5637b4f736e2aa64cfb98d6276e168dd1e7cdfc6d")
	bSig, _ := hex.DecodeString("b49b04b6b73c4d2f64f2ec8ef75955004361183c903d90bff384125450872e0a1e1f7909e3fa73f2bd2795a746862e7c519565fade6bb3ccd402410c300b8215")
	//expected_msg, _ := hex.DecodeString("584d5450203a20437265617465204964656e746974790a30386332633938336633386433303161343330613431303432343437396563376261626230663235646633393133313665336562623533653566626234306633626132336135393137666437613533313636613834373130613132323464616433613362323364353337333262613838303332636337623661666238396630386236616333613463373538316264663661656466613230350a0a466f72206d6f726520696e666f3a2068747470733a2f2f786d74702e6f72672f7369676e6174757265732f")
	//expectedRecovery := uint8(0)
	bSig = append(bSig, 1)
	log.Info("bsig", zap.Any("R", bSig[64]), zap.Int("len", len(bSig)), zap.Uint8s("Any", bSig))
	msg := walletSignatureMessage()
	hash := crypto.Keccak256(msg)

	sigPublicKey, err := secp256k1.RecoverPubkey(hash, bSig)
	if err != nil {
		log.Error("Bad recover", zap.Error(err))
	}

	log.Info("RPUB", zap.Any("Any", sigPublicKey))
}

func TestA(t *testing.T) {
	a := walletSignatureMessage()
	b := walletSignatureTest()

	require.Equal(t, a, b)
}
