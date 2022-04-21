package logging

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type multiaddrs []multiaddr.Multiaddr

func (addrs multiaddrs) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, addr := range addrs {
		encoder.AppendString(addr.String())
	}
	return nil
}

func MultiAddrs(key string, addrs ...multiaddr.Multiaddr) zapcore.Field {
	return zap.Array(key, multiaddrs(addrs))
}

type hostID peer.ID

func (id hostID) String() string { return peer.Encode(peer.ID(id)) }

func HostID(key string, id peer.ID) zapcore.Field {
	return zap.Stringer(key, hostID(id))
}
