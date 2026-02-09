package api

import (
	"context"

	migrationv1 "github.com/xmtp/xmtp-node-go/pkg/proto/migration/api/v1"
	"go.uber.org/zap"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	migrationv1.UnimplementedD14NMigrationApiServer

	log           *zap.Logger
	d14nCutoverNs uint64

	ctx       context.Context
	ctxCancel func()
}

func NewService(log *zap.Logger, d14nCutoverNs uint64) *Service {
	s := &Service{
		log:           log.Named("migration/v1"),
		d14nCutoverNs: d14nCutoverNs,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	s.log.Info("Starting migration service", zap.Uint64("d14n_cutover_ns", d14nCutoverNs))
	return s
}

func (s *Service) Close() {
	s.log.Info("closing")

	if s.ctxCancel != nil {
		s.ctxCancel()
	}

	s.log.Info("closed")
}

func (s *Service) FetchD14NCutover(
	ctx context.Context,
	req *emptypb.Empty,
) (*migrationv1.FetchD14NCutoverResponse, error) {
	return &migrationv1.FetchD14NCutoverResponse{
		TimestampNs: s.d14nCutoverNs,
	}, nil
}
