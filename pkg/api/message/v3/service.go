package api

import (
	"context"

	wakunode "github.com/waku-org/go-waku/waku/v2/node"
	proto "github.com/xmtp/proto/v3/go/message_api/v3"
	"github.com/xmtp/xmtp-node-go/pkg/mlsstore"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	proto.UnimplementedMlsApiServer

	log               *zap.Logger
	waku              *wakunode.WakuNode
	messageStore      *store.Store
	mlsStore          mlsstore.MlsStore
	validationService mlsvalidate.MLSValidationService

	ctx       context.Context
	ctxCancel func()
}

func NewService(node *wakunode.WakuNode, logger *zap.Logger, messageStore *store.Store, mlsStore mlsstore.MlsStore, validationService mlsvalidate.MLSValidationService) (s *Service, err error) {
	s = &Service{
		log:               logger.Named("message/v3"),
		waku:              node,
		messageStore:      messageStore,
		mlsStore:          mlsStore,
		validationService: validationService,
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
	results, err := s.validationService.ValidateKeyPackages(ctx, [][]byte{req.LastResortKeyPackage.KeyPackageTlsSerialized})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identity: %s", err)
	}
	if len(results) != 1 {
		return nil, status.Errorf(codes.Internal, "unexpected number of results: %d", len(results))
	}

	installationId := results[0].InstallationId
	walletAddress := results[0].WalletAddress

	err = s.mlsStore.CreateInstallation(ctx, installationId, walletAddress, req.LastResortKeyPackage.KeyPackageTlsSerialized)
	if err != nil {
		return nil, err
	}

	return &proto.RegisterInstallationResponse{
		InstallationId: installationId,
	}, nil
}

func (s *Service) ConsumeKeyPackages(ctx context.Context, req *proto.ConsumeKeyPackagesRequest) (*proto.ConsumeKeyPackagesResponse, error) {
	ids := req.InstallationIds
	keyPackages, err := s.mlsStore.ConsumeKeyPackages(ctx, ids)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to consume key packages: %s", err)
	}
	keyPackageMap := make(map[string]int)
	for idx, id := range ids {
		keyPackageMap[id] = idx
	}

	resPackages := make([]*proto.ConsumeKeyPackagesResponse_KeyPackage, len(keyPackages))
	for _, keyPackage := range keyPackages {

		idx, ok := keyPackageMap[keyPackage.InstallationId]
		if !ok {
			return nil, status.Errorf(codes.Internal, "could not find key package for installation")
		}

		resPackages[idx] = &proto.ConsumeKeyPackagesResponse_KeyPackage{
			KeyPackageTlsSerialized: keyPackage.Data,
		}
	}

	return &proto.ConsumeKeyPackagesResponse{
		KeyPackages: resPackages,
	}, nil
}

func (s *Service) PublishToGroup(ctx context.Context, req *proto.PublishToGroupRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) PublishWelcomes(ctx context.Context, req *proto.PublishWelcomesRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) UploadKeyPackages(ctx context.Context, req *proto.UploadKeyPackagesRequest) (*emptypb.Empty, error) {
	keyPackageBytes := make([][]byte, len(req.KeyPackages))
	for i, keyPackage := range req.KeyPackages {
		keyPackageBytes[i] = keyPackage.KeyPackageTlsSerialized
	}
	validationResults, err := s.validationService.ValidateKeyPackages(ctx, keyPackageBytes)
	if err != nil {
		// TODO: Differentiate between validation errors and internal errors
		return nil, status.Errorf(codes.InvalidArgument, "invalid identity: %s", err)
	}

	keyPackageModels := make([]*mlsstore.KeyPackage, len(validationResults))
	for i, validationResult := range validationResults {
		kp := mlsstore.NewKeyPackage(validationResult.InstallationId, keyPackageBytes[i], false)
		keyPackageModels[i] = kp
	}
	err = s.mlsStore.InsertKeyPackages(ctx, keyPackageModels)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert key packages: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) RevokeInstallation(ctx context.Context, req *proto.RevokeInstallationRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetIdentityUpdates(ctx context.Context, req *proto.GetIdentityUpdatesRequest) (*proto.GetIdentityUpdatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}
