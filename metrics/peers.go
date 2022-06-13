package metrics

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/xmtp/xmtp-node-go/logging"
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

func EmitPeersByProtocol(ctx context.Context, host host.Host) {
	byProtocol := map[string]int64{}
	ps := host.Peerstore()
	for _, peer := range ps.Peers() {
		protos, err := ps.GetProtocols(peer)
		if err != nil {
			continue
		}
		for _, proto := range protos {
			byProtocol[proto]++
		}
	}
	for proto, count := range byProtocol {
		if err := stats.RecordWithTags(ctx, []tag.Mutator{tag.Insert(TagProto, proto)}, PeersByProto.M(count)); err != nil {
			logging.From(ctx).Warn("recording metric", zap.String("metric", PeersByProto.Name()), zap.String("proto", proto), zap.Error(err))
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
	stats.Record(ctx, BootstrapPeers.M(float64(bootstrapPeersFound)/float64(len(bootstrapPeers))))
}
