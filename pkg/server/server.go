package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	"github.com/xmtp/xmtp-node-go/pkg/api"
	"github.com/xmtp/xmtp-node-go/pkg/crdt"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"go.uber.org/zap"
)

type Server struct {
	log     *zap.Logger
	metrics *metrics.Server
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	grpc    *api.Server
	crdt    *crdt.Node
}

// Create a new Server
func New(ctx context.Context, log *zap.Logger, options Options) (*Server, error) {
	s := &Server{
		log: log,
	}
	s.ctx, s.cancel = context.WithCancel(ctx)

	if options.Metrics.Enable {
		var err error
		s.metrics, err = metrics.NewMetricsServer(s.ctx, s.log, options.Metrics.Address, options.Metrics.Port)
		if err != nil {
			return nil, errors.Wrap(err, "initializing metrics server")
		}
	}

	if options.NodeKey == "" {
		nodeKey := os.Getenv("XMTP_NODE_KEY")
		if nodeKey != "" {
			options.NodeKey = nodeKey
		}
	}

	// Initialize CRDT ds node.
	crdt, err := crdt.NewNode(ctx, log, crdt.Options{
		NodeKey:            options.NodeKey,
		DataPath:           options.DataPath,
		P2PPort:            options.P2PPort,
		P2PPersistentPeers: options.P2PPersistentPeers,
	})
	if err != nil {
		return nil, errors.Wrap(err, "initializing crdt")
	}
	s.crdt = crdt

	// Initialize gRPC server.
	s.grpc, err = api.New(
		&api.Config{
			Options: options.API,
			Log:     s.log.Named("api"),
			CRDT:    s.crdt,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing grpc server")
	}
	return s, nil
}

func (s *Server) WaitForShutdown() {
	// Wait for a SIGINT or SIGTERM signal
	termChannel := make(chan os.Signal, 1)
	signal.Notify(termChannel, syscall.SIGINT, syscall.SIGTERM)
	<-termChannel
	s.Shutdown()
}

func (s *Server) Shutdown() {
	s.log.Info("shutting down...")

	if s.metrics != nil {
		if err := s.metrics.Stop(s.ctx); err != nil {
			s.log.Error("stopping metrics", zap.Error(err))
		}
	}

	// Close the gRPC s.
	if s.grpc != nil {
		s.grpc.Close()
	}

	// Close crdt.
	if s.crdt != nil {
		err := s.crdt.Close()
		if err != nil {
			s.log.Error("closing crdt", zap.Error(err))
		}
	}

	// Cancel outstanding goroutines
	s.cancel()
	s.wg.Wait()
	s.log.Info("shutdown complete")
}
