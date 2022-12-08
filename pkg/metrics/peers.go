package metrics

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
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

func EmitBootstrapPeersConnected(ctx context.Context, host host.Host, bootstrapPeers map[peer.ID]bool) {
	var bootstrapPeersFound int
	for _, peer := range host.Network().Peers() {
		if bootstrapPeers[peer] {
			bootstrapPeersFound++
		}
	}
	record(ctx, BootstrapPeers.M(float64(bootstrapPeersFound)/float64(len(bootstrapPeers))))
}
