package metrics

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var TagProto, _ = tag.NewKey("protocol")

var PeersByProto = stats.Int64("peers_by_proto", "Count of peers by protocol", stats.UnitDimensionless)
var PeersByProtoView = &view.View{
	Name:        "xmtp_peers_by_proto",
	Measure:     PeersByProto,
	Description: "Current number of peers by protocol",
	Aggregation: view.LastValue(),
	TagKeys:     []tag.Key{TagProto},
}

var BootstrapPeers = stats.Float64("bootstrap_peers", "Percentage of bootstrap peers connected", stats.UnitDimensionless)
var BootstrapPeersView = &view.View{
	Name:        "xmtp_bootstrap_peers",
	Measure:     BootstrapPeers,
	Description: "Percentage of bootstrap peers connected",
	Aggregation: view.LastValue(),
}

func EmitPeersByProtocol(ctx context.Context, log *zap.Logger, host host.Host) {
	byProtocol := map[string]int64{}
	ps := host.Peerstore()
	for _, peer := range ps.Peers() {
		protocols, err := ps.GetProtocols(peer)
		if err != nil {
			continue
		}
		for _, proto := range protocol.ConvertToStrings(protocols) {
			byProtocol[proto]++
		}
	}
	for proto, count := range byProtocol {
		mutators := []tag.Mutator{tag.Insert(TagProto, proto)}
		err := recordWithTags(ctx, mutators, PeersByProto.M(count))
		if err != nil {
			log.Warn("recording metric", zap.String("metric", PeersByProto.Name()), zap.String("proto", proto), zap.Error(err))
		}
	}
}

func EmitBootstrapPeersConnected(ctx context.Context, host host.Host, bootstrapPeers map[peer.ID]bool) {
	var bootstrapPeersFound int
	for _, peer := range host.Network().Peers() {
		if bootstrapPeers[peer] {
			bootstrapPeersFound++
		}
	}
	record(ctx, BootstrapPeers.M(float64(bootstrapPeersFound)/float64(len(bootstrapPeers))))
}
