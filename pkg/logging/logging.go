// logging implements custom logging field types for commonly
// logged values like host ID or wallet address.
//
// implementation purposely does as little as possible at field creation time,
// and postpones any transformation to output time by relying on the generic
// zap types like zap.Stringer, zap.Array, zap.Object
package logging

import (
	"fmt"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type timestamp int64

func (t timestamp) String() string {
	return time.Unix(0, int64(t)).Format(time.RFC3339)
}

func Time(key string, time int64) zapcore.Field {
	return zap.Stringer(key, timestamp(time))
}

// WalletAddress creates a field for a wallet address.
func WalletAddress(address string) zapcore.Field {
	return zap.String("wallet_address", address)
}

func WalletAddressLabelled(label string, address types.WalletAddr) zapcore.Field {
	return zap.String(label, string(address))
}

type valueType struct{ val interface{} }

func ValueType(key string, val interface{}) zap.Field {
	return zap.Stringer(key, valueType{val})
}

func (vt valueType) String() string {
	return fmt.Sprintf("%T", vt.val)
}
