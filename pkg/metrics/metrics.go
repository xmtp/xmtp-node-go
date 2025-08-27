package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"

	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pires/go-proxyproto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
)

type Server struct {
	ctx  context.Context
	log  *zap.Logger
	http net.Listener
}

func NewMetricsServer(
	ctx context.Context,
	address string,
	port int,
	log *zap.Logger,
	reg *prometheus.Registry,
) (*Server, error) {
	s := &Server{
		ctx: ctx,
		log: log.Named("metrics"),
	}

	addr := fmt.Sprintf("%s:%d", address, port)
	httpListener, err := net.Listen("tcp", addr)
	s.http = &proxyproto.Listener{Listener: httpListener}
	if err != nil {
		return nil, err
	}
	registerCollectors(reg)
	reg.MustRegister(grpcPrometheus.DefaultServerMetrics)
	srv := http.Server{
		Addr:    addr,
		Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}),
	}

	go tracing.PanicWrap(ctx, "metrics server", func(_ context.Context) {
		s.log.Info("serving metrics http", zap.String("address", s.http.Addr().String()))
		err = srv.Serve(s.http)
		if err != nil {
			s.log.Error("serving http", zap.Error(err))
		}
	})

	return s, nil
}

func (s *Server) Close() error {
	return s.http.Close()
}

func registerCollectors(reg prometheus.Registerer) {
	cols := []prometheus.Collector{
		PeersByProto,
		BootstrapPeers,
		StoredMessages,
		apiRequests,
		apiRequestsDuration,
		subscribeTopicsLength,
		subscribedTopics,
		publishedEnvelopeSize,
		publishedEnvelopeCount,
		queryDuration,
		queryResultLength,
		ratelimiterBuckets,
		ratelimiterBucketsDeleted,
		mlsSentGroupMessageSize,
		mlsSentGroupMessageCount,
		mlsSentWelcomeMessageSize,
		mlsSentWelcomeMessageCount,
		mlsIdentityUpdateSize,
		mlsIdentityUpdateCount,
		mlsCommitLogEntrySize,
		mlsCommitLogEntryCount,
		mlsKeyPackageSize,
		mlsKeyPackageCount,
	}
	for _, col := range cols {
		reg.MustRegister(col)
	}
}
