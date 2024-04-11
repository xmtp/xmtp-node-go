package api

import (
	"context"

	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	api "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	api.UnimplementedIdentityApiServer

	log   *zap.Logger
	store mlsstore.MlsStore

	ctx       context.Context
	ctxCancel func()
}

func NewService(log *zap.Logger, store mlsstore.MlsStore) (s *Service, err error) {
	s = &Service{
		log:   log.Named("identity"),
		store: store,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	s.log.Info("Starting identity service")
	return s, nil
}

func (s *Service) Close() {
	s.log.Info("closing")

	if s.ctxCancel != nil {
		s.ctxCancel()
	}

	s.log.Info("closed")
}

func (s *Service) PublishIdentityUpdate(ctx context.Context, req *api.PublishIdentityUpdateRequest) (*api.PublishIdentityUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetIdentityUpdates(ctx context.Context, req *api.GetIdentityUpdatesRequest) (*api.GetIdentityUpdatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetInboxIds(ctx context.Context, req *api.GetInboxIdsRequest) (*api.GetInboxIdsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}
