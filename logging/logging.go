// logging implements custom logging field types for commonly
// logged values like host ID or wallet address.
//
// implementation purposely does as little as possible at field creation time,
// and postpones any transformation to output time by relying on the generic
// zap types like zap.Stringer, zap.Array, zap.Object
//
package logging

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
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
