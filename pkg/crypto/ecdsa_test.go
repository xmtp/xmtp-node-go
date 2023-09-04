package crypto_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/crypto"
)

func TestHexToECDSA(t *testing.T) {
	tcs := []struct {
		name        string
		key         string
		expectedErr string
	}{
		{
			name: "length 64 valid",
			key:  "1df2b07b69f67be885148ef85ced1576048c7c33fb861f7cd8e9b62e13013330",
		},
		{
			name: "length 62 valid",
			key:  "f2b07b69f67be885148ef85ced1576048c7c33fb861f7cd8e9b62e13013330",
		},
		{
			name: "length 60 valid",
			key:  "b07b69f67be885148ef85ced1576048c7c33fb861f7cd8e9b62e13013330",
		},
		{
			name:        "length 58 valid",
			key:         "7b69f67be885148ef85ced1576048c7c33fb861f7cd8e9b62e13013330",
			expectedErr: "invalid length, need 256 bits",
		},
		{
			name:        "length 66 valid",
			key:         "c7c33fb07b69f67be885148ef85ced1576048c7c33fb861f7cd8e9b62e13013330",
			expectedErr: "invalid length, need 256 bits",
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key, err := crypto.HexToECDSA(tc.key)
			if tc.expectedErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedErr)
			} else {
				keyHex := hex.EncodeToString(key.D.Bytes())
				require.Equal(t, tc.key, keyHex)
			}
		})
	}
}
