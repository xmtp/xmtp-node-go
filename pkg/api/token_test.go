package api

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Reader.Read(b)
	return b
}

func Test_NominalV2(t *testing.T) {
	log, _ := zap.NewDevelopment()
	ctx := context.Background()
	now := time.Now()
	token, data, err := generateV2AuthToken(now.Add(-time.Minute))
	require.NoError(t, err)
	walletAddr, err := validateToken(ctx, log, token, now)
	require.NoError(t, err)
	require.Equal(t, data.WalletAddr, string(walletAddr), "wallet address mismatch")
}

func Test_XmtpjsToken(t *testing.T) {
	// This token was captured from xmtp-js Authn test suite run.
	tokenBytes := "CpQBCPe9zuiqMBJGCkQKQHOOPN+A8lNOnt42Pis2R6jkbDvhsC0EsLDCFf+VGh/yYssKafl7ma+4bmo/fEEvf435Wy/klBXOQN8Nu3lIW+kQARpDCkEEkqxMLK+SGTEE7lnXmOxUoWynlTBg8TAQNIXf2IcometsYfw2UO9C+IryDLH4utr2pHv726oFyuMhpSHi/YNM1hI2CioweDJEMGU2MTRlOERjOEFkZjgyYzcwMDkwRjkwNWU2Y0ZDMURFODQ5MDAQgKumuYqujYYXGkYKRApAQ+8Nug8k6TWLA/pNcPDgW01Sdn9o2h3484jSGlWkxDl1o8mFtsxofxOA5gUxt+/nw33kSER0Yukl1OO2Tfd68hAB"
	tokenAddress := "0x2D0e614e8Dc8Adf82c70090F905e6cFC1DE84900"
	tokenCreatedNs := int64(1660761120550000000)

	log, _ := zap.NewDevelopment()
	ctx := context.Background()
	now := time.Unix(0, tokenCreatedNs).Add(10 * time.Minute)
	token, err := decodeAuthToken(tokenBytes)
	require.NoError(t, err)
	walletAddr, err := validateToken(ctx, log, token, now)
	require.NoError(t, err)
	require.Equal(t, tokenAddress, string(walletAddr))
}

func Test_BadAuthSig(t *testing.T) {
	log, _ := zap.NewDevelopment()
	ctx := context.Background()
	now := time.Now()
	token, _, err := generateV2AuthToken(now.Add(-time.Minute))
	require.NoError(t, err)
	token.GetAuthDataSignature().GetEcdsaCompact().Bytes = randomBytes(64)
	_, err = validateToken(ctx, log, token, now)
	require.Error(t, err)
	require.Equal(t, err, ErrInvalidSignature)
}

func Test_SignatureMismatch(t *testing.T) {
	log, _ := zap.NewDevelopment()
	ctx := context.Background()
	now := time.Now()
	token1, _, err := generateV2AuthToken(now.Add(-time.Minute))
	require.NoError(t, err)
	token2, _, err := generateV2AuthToken(now.Add(-time.Minute))
	require.NoError(t, err)

	// Nominal Checks
	_, err = validateToken(ctx, log, token1, now)
	require.NoError(t, err)
	_, err = validateToken(ctx, log, token2, now)
	require.NoError(t, err)

	// Swap Signatures to check for valid but mismatched signatures
	token1.IdentityKey.Signature, token2.AuthDataSignature = token2.AuthDataSignature, token1.IdentityKey.Signature

	// Expect Errors as the derived walletAddr will not match the one supplied in AuthData
	_, err = validateToken(ctx, log, token1, now)
	require.Error(t, err)
	_, err = validateToken(ctx, log, token1, now)
	require.Error(t, err)
}

func Test_DecodeXmtpjsToken(t *testing.T) {
	_, err := decodeAuthToken("CpIBCOqy8+iqMBJECkIKQGsPVyMg1ZjfJXCu7+IxuRJ9/JrWvPIPsZ+GjRM+eQ8kLCfuOequP3GscERICC3qRk/l4eCW/kqM5fbd5/TcQHEaQwpBBAzYR20tIAxsD3cXSnrDBoZ8xr2FPgA9d4u1QzaNt/t0tpQ9G6i9ju35nCDTr35iMbZvbYfZJKYbed2ABOZ2xdMSNgoqMHg1NDY3ZUU5ZGU5MmE3ZDgzMGE3MTMxNTZkZjg4ZGQ2Qjg3MDQwZUY2EMCs9dTXv42GFxpGCkQKQBoIQ77Vi0SU3M5s1WsthNiJgI8Vx89cSzZvJiaIVNJYCJO2x55eEex5R4o55XFvoNrgGcRyDOGg3WMsTWlWOx8QAQ==")
	require.NoError(t, err)
}

func Test_DecodeInvalidToken(t *testing.T) {
	token, err := decodeAuthToken("aGk=")
	require.Error(t, err)
	require.Nil(t, token)
	require.Contains(t, err.Error(), "missing identity key")
}

func Test_DecodeEmptyToken(t *testing.T) {
	token, err := decodeAuthToken("")
	require.Error(t, err)
	require.Nil(t, token)
	require.Contains(t, err.Error(), "missing identity key")
}
