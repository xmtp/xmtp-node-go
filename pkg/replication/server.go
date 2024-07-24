package replication

import (
	"context"
	"crypto/ecdsa"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/xmtp/xmtp-node-go/pkg/replication/api"
	"github.com/xmtp/xmtp-node-go/pkg/replication/registry"
	"go.uber.org/zap"
)

type Server struct {
	options      Options
	log          *zap.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	apiServer    *api.ApiServer
	nodeRegistry registry.NodeRegistry
	privateKey   *ecdsa.PrivateKey
}

func New(ctx context.Context, log *zap.Logger, options Options, nodeRegistry registry.NodeRegistry) (*Server, error) {
	var err error
	s := &Server{
		options:      options,
		log:          log,
		nodeRegistry: nodeRegistry,
	}
	s.privateKey, err = parsePrivateKey(options.PrivateKeyString)
	if err != nil {
		return nil, err
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.apiServer, err = api.NewAPIServer(ctx, log, options.API.Port)
	if err != nil {
		return nil, err
	}
	log.Info("Replication server started", zap.Int("port", options.API.Port))
	return s, nil
}

func (s *Server) Addr() net.Addr {
	return s.apiServer.Addr()
}

func (s *Server) WaitForShutdown() {
	termChannel := make(chan os.Signal, 1)
	signal.Notify(termChannel, syscall.SIGINT, syscall.SIGTERM)
	<-termChannel
	s.Shutdown()
}

func (s *Server) Shutdown() {
	s.cancel()
}

func parsePrivateKey(privateKeyString string) (*ecdsa.PrivateKey, error) {
	return crypto.HexToECDSA(privateKeyString)
}
