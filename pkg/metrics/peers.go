package metrics

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/prometheus/client_golang/prometheus"
)

var PeersByProto = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "xmtp_peers_by_proto",
		Help: "Count of peers by protocol",
	},
	[]string{"protocol"},
)

var BootstrapPeers = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "xmtp_bootstrap_peers",
		Help: "Percentage of bootstrap peers connected",
	},
)

func EmitPeersByProtocol(ctx context.Context, host host.Host) {
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
		PeersByProto.WithLabelValues(proto).Set(float64(count))
	}
}

func EmitBootstrapPeersConnected(ctx context.Context, host host.Host, bootstrapPeers map[peer.ID]bool) {
	var bootstrapPeersFound int
	for _, peer := range host.Network().Peers() {
		if bootstrapPeers[peer] {
			bootstrapPeersFound++
		}
	}
	BootstrapPeers.Set(float64(bootstrapPeersFound) / float64(len(bootstrapPeers)))
}
