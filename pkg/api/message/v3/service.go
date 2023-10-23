package api

import (
	"context"

	wakunode "github.com/waku-org/go-waku/waku/v2/node"
	proto "github.com/xmtp/proto/v3/go/message_api/v3"
	"github.com/xmtp/xmtp-node-go/pkg/mlsstore"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	proto.UnimplementedMlsApiServer

	log          *zap.Logger
	waku         *wakunode.WakuNode
	messageStore *store.Store
	mlsStore     mlsstore.MlsStore

	ctx       context.Context
	ctxCancel func()
}

func NewService(node *wakunode.WakuNode, logger *zap.Logger, messageStore *store.Store, mlsStore mlsstore.MlsStore) (s *Service, err error) {
	s = &Service{
		log:          logger.Named("message/v3"),
		waku:         node,
		messageStore: messageStore,
		mlsStore:     mlsStore,
	}

	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	return s, nil
}

func (s *Service) Close() {
	if s.ctxCancel != nil {
		s.ctxCancel()
	}
}

func (s *Service) RegisterInstallation(ctx context.Context, req *proto.RegisterInstallationRequest) (*proto.RegisterInstallationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) ConsumeKeyPackages(ctx context.Context, req *proto.ConsumeKeyPackagesRequest) (*proto.ConsumeKeyPackagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) PublishToGroup(ctx context.Context, req *proto.PublishToGroupRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) PublishWelcomes(ctx context.Context, req *proto.PublishWelcomesRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) UploadKeyPackages(ctx context.Context, req *proto.UploadKeyPackagesRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) RevokeInstallation(ctx context.Context, req *proto.RevokeInstallationRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetIdentityUpdates(ctx context.Context, req *proto.GetIdentityUpdatesRequest) (*proto.GetIdentityUpdatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}
