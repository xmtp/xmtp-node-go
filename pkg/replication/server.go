package replication

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/xmtp/xmtp-node-go/pkg/replication/api"
	"go.uber.org/zap"
)

type Server struct {
	options   Options
	log       *zap.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	apiServer *api.ApiServer
}

func New(ctx context.Context, log *zap.Logger, options Options) (*Server, error) {
	var err error

	s := &Server{
		options: options,
		log:     log,
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.apiServer, err = api.NewAPIServer(ctx, log, options.API.Port)
	if err != nil {
		return nil, err
	}
	log.Info("Replication server started", zap.Int("port", options.API.Port))
	return s, nil
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
