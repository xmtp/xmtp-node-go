package authn

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"go.uber.org/zap"
)

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Reader.Read(b)
	return b
}

func Test_Nominal(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := logging.With(context.Background(), logger)
	now := time.Now()
	token, data, err := GenerateToken(now.Add(-time.Minute))
	require.NoError(t, err)
	walletAddr, err := validateToken(ctx, token, now)
	require.NoError(t, err)
	require.Equal(t, data.WalletAddr, string(walletAddr), "wallet address mismatch")
}

func Test_XmtpjsToken(t *testing.T) {
	// This token was captured from xmtp-js Authn test suite run.
	tokenBytes := "CpQBCJ-Sl5epMBJGCkQKQPW_UHrWPXGJfHgSmK8vuo9vfy358G7lCQT9ugdXX9W2QVtQz6PJb7aXtfrfsUIBMAQp23iMJdWMtQdsEfnkoPQQARpDCkEEkCxWmMaKJQQMllUYDNJsqPpG_nV-jMs3fuT7wX4oCZd7snbWF5_qI-XJbaCIjtTiDXBkHkih5p56YiEvhHuLURI2CioweEJCNTBEMjczMkZEOGQ0OTgyZUFDMzE0QmRBN2QzNTc5MDQyMDAzQTkQgLzBk6y_qYUXGkYKRApAo42fSDqHB4kY3SwClbGek7O4cxnleu52gHIWEjT3lrcI1RxlJK4KI69UtFd_YRTDzwLwZrbZLXXlYCWqJlev6xAB"
	tokenAddress := "0xBB50D2732FD8d4982eAC314BdA7d3579042003A9"
	tokenCreatedNs := int64(1660321909062000128)

	logger, _ := zap.NewDevelopment()
	ctx := logging.With(context.Background(), logger)
	now := time.Unix(0, tokenCreatedNs).Add(10 * time.Minute)
	token, err := DecodeToken(tokenBytes)
	require.NoError(t, err)
	walletAddr, err := validateToken(ctx, token, now)
	require.NoError(t, err)
	require.Equal(t, tokenAddress, string(walletAddr))
}

func Test_BadAuthSig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := logging.With(context.Background(), logger)
	now := time.Now()
	token, _, err := GenerateToken(now.Add(-time.Minute))
	require.NoError(t, err)
	token.GetAuthDataSignature().GetEcdsaCompact().Bytes = randomBytes(64)
	_, err = validateToken(ctx, token, now)
	require.Error(t, err)
	require.Equal(t, err, ErrInvalidSignature)
}

func Test_SignatureMismatch(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := logging.With(context.Background(), logger)
	now := time.Now()
	token1, _, err := GenerateToken(now.Add(-time.Minute))
	require.NoError(t, err)
	token2, _, err := GenerateToken(now.Add(-time.Minute))
	require.NoError(t, err)

	// Nominal Checks
	_, err = validateToken(ctx, token1, now)
	require.NoError(t, err)
	_, err = validateToken(ctx, token2, now)
	require.NoError(t, err)

	// Swap Signatures to check for valid but mismatched signatures
	token1.IdentityKey.Signature, token2.AuthDataSignature = token2.AuthDataSignature, token1.IdentityKey.Signature

	// Expect Errors as the derived walletAddr will not match the one supplied in AuthData
	_, err = validateToken(ctx, token1, now)
	require.Error(t, err)
	_, err = validateToken(ctx, token1, now)
	require.Error(t, err)
}

func Test_DecodeToken(t *testing.T) {
	encoded := "CpIBCNq0jcWqMBJECkIKQPZCm1Csbn4SjAE1jyn5AGaGXoOLMoVOTzjJxuSHlwO2fb3sdtfnezZnpF0YBcMmtKlVhSlEuw3vvYBtONWGK34aQwpBBISsYDcvnpmW5m1b6-rL0yfntnx6VeSMWm4OQ7fOXNOD7-1pbpWpEeTEZc6ww9WWfGg_9_juzP2TpvDKyv3pwx4SNgoqMHg5NDA5NTU5RGE0QzgzZmU3ZTc4YjBFOTE2MzBFZjAwNTQ2MDcyNzE4EMCN6uTjt_yFFxpGCkQKQJz75ZZhpGu15ZWruXSyNZFFaDf8JJSdAl8XYLDMLtg8F-85_ARF9Y9m2GoHn3c6QtbgUPe17HzXMAbHtTVd8MIQAQ"
	_, err := DecodeToken(encoded)
	require.NoError(t, err)
}
