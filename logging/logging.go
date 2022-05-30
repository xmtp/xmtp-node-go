// logging implements custom logging field types for commonly
// logged values like host ID or wallet address.
//
// implementation purposely does as little as possible at field creation time,
// and postpones any transformation to output time by relying on the generic
// zap types like zap.Stringer, zap.Array, zap.Object
//
package logging

import (
	"github.com/status-im/go-waku/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Re-export relevant go-waku helpers

	MultiAddrs = logging.MultiAddrs
	HostID     = logging.HostID
	Time       = logging.Time
	Filters    = logging.Filters
	PagingInfo = logging.PagingInfo
	HexBytes   = logging.HexBytes
	TCPAddr    = logging.TCPAddr
	UDPAddr    = logging.UDPAddr
)

// Wallet Address

func WalletAddress(address string) zapcore.Field {
	return zap.String("wallet_address", address)
}
