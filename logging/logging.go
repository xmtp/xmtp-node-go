// logging implements custom logging field types for commonly
// logged values like host ID or wallet address.
//
// implementation purposely does as little as possible at field creation time,
// and postpones any transformation to output time by relying on the generic
// zap types like zap.Stringer, zap.Array, zap.Object
//
package logging

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// List of multiaddrs

type multiaddrs []multiaddr.Multiaddr

func MultiAddrs(key string, addrs ...multiaddr.Multiaddr) zapcore.Field {
	return zap.Array(key, multiaddrs(addrs))
}

func (addrs multiaddrs) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, addr := range addrs {
		encoder.AppendString(addr.String())
	}
	return nil
}

// Host ID

type hostID peer.ID

func HostID(key string, id peer.ID) zapcore.Field {
	return zap.Stringer(key, hostID(id))
}

func (id hostID) String() string { return peer.Encode(peer.ID(id)) }

// Time - looks like Waku is using Microsecond Unix Time

type timestamp int64

func Time(key string, time int64) zapcore.Field {
	return zap.Stringer(key, timestamp(time))
}

func (t timestamp) String() string {
	return time.UnixMicro(int64(t)).Format(time.RFC3339)
}

// History Query Filters

type historyFilters []*pb.ContentFilter

func Filters(key string, filters []*pb.ContentFilter) zapcore.Field {
	return zap.Array(key, historyFilters(filters))
}

func (filters historyFilters) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, filter := range filters {
		encoder.AppendString(filter.ContentTopic)
	}
	return nil
}

// Wallet Address

func WalletAddress(address string) zapcore.Field {
	return zap.String("wallet_address", address)
}