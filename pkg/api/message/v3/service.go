package api

import (
	"context"

	wakunode "github.com/waku-org/go-waku/waku/v2/node"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	proto "github.com/xmtp/proto/v3/go/message_api/v3"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"github.com/xmtp/xmtp-node-go/pkg/mlsstore"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
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
}

func NewService(node *wakunode.WakuNode, logger *zap.Logger, messageStore *store.Store, mlsStore mlsstore.MlsStore, validationService mlsvalidate.MLSValidationService) (s *Service, err error) {
	s = &Service{
		log:               logger.Named("message/v3"),
		waku:              node,
		messageStore:      messageStore,
		mlsStore:          mlsStore,
		validationService: validationService,
	}

	s.log.Info("Starting MLS service")
	return s, nil
}

func (s *Service) RegisterInstallation(ctx context.Context, req *proto.RegisterInstallationRequest) (*proto.RegisterInstallationResponse, error) {
	if err := validateRegisterInstallationRequest(req); err != nil {
		return nil, err
	}

	results, err := s.validationService.ValidateKeyPackages(ctx, [][]byte{req.LastResortKeyPackage.KeyPackageTlsSerialized})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identity: %s", err)
	}
	if len(results) != 1 {
		return nil, status.Errorf(codes.Internal, "unexpected number of results: %d", len(results))
	}

	installationId := results[0].InstallationId
	walletAddress := results[0].WalletAddress

	if err = s.mlsStore.CreateInstallation(ctx, installationId, walletAddress, req.LastResortKeyPackage.KeyPackageTlsSerialized); err != nil {
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

func (s *Service) PublishToGroup(ctx context.Context, req *proto.PublishToGroupRequest) (res *emptypb.Empty, err error) {
	if err = validatePublishToGroupRequest(req); err != nil {
		return nil, err
	}

	messages := make([][]byte, len(req.Messages))
	for i, message := range req.Messages {
		v1 := message.GetV1()
		if v1 == nil {
			return nil, status.Errorf(codes.InvalidArgument, "message must be v1")
		}
		messages[i] = v1.MlsMessageTlsSerialized
	}

	validationResults, err := s.validationService.ValidateGroupMessages(ctx, messages)
	if err != nil {
		// TODO: Separate validation errors from internal errors
		return nil, status.Errorf(codes.InvalidArgument, "invalid group message: %s", err)
	}

	for i, result := range validationResults {
		message := messages[i]

		if err = isReadyToSend(result.GroupId, message); err != nil {
			return nil, err
		}

		// TODO: Wrap this in a transaction so publishing is all or nothing
		if err = s.publishMessage(ctx, topic.BuildGroupTopic(result.GroupId), message); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to publish message: %s", err)
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) publishMessage(ctx context.Context, contentTopic string, message []byte) error {
	log := s.log.Named("publish-mls").With(zap.String("content_topic", contentTopic))
	env, err := s.messageStore.InsertMlsMessage(ctx, contentTopic, message)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to insert message: %s", err)
	}

	if _, err = s.waku.Relay().Publish(ctx, &wakupb.WakuMessage{
		ContentTopic: contentTopic,
		Timestamp:    int64(env.TimestampNs),
		Payload:      message,
	}); err != nil {
		return status.Errorf(codes.Internal, "failed to publish message: %s", err)
	}

	metrics.EmitPublishedEnvelope(ctx, log, env)

	return nil
}

func (s *Service) PublishWelcomes(ctx context.Context, req *proto.PublishWelcomesRequest) (res *emptypb.Empty, err error) {
	if err = validatePublishWelcomesRequest(req); err != nil {
		return nil, err
	}

	// TODO: Wrap this in a transaction so publishing is all or nothing
	for _, welcome := range req.WelcomeMessages {
		contentTopic := topic.BuildWelcomeTopic(welcome.InstallationId)
		if err = s.publishMessage(ctx, contentTopic, welcome.WelcomeMessage.GetV1().Ciphertext); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to publish welcome message: %s", err)
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UploadKeyPackages(ctx context.Context, req *proto.UploadKeyPackagesRequest) (res *emptypb.Empty, err error) {
	if err = validateUploadKeyPackagesRequest(req); err != nil {
		return nil, err
	}
	// Extract the key packages from the request
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

	if err = s.mlsStore.InsertKeyPackages(ctx, keyPackageModels); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert key packages: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) RevokeInstallation(ctx context.Context, req *proto.RevokeInstallationRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetIdentityUpdates(ctx context.Context, req *proto.GetIdentityUpdatesRequest) (res *proto.GetIdentityUpdatesResponse, err error) {
	if err = validateGetIdentityUpdatesRequest(req); err != nil {
		return nil, err
	}

	walletAddresses := req.WalletAddresses
	updates, err := s.mlsStore.GetIdentityUpdates(ctx, req.WalletAddresses, int64(req.StartTimeNs))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get identity updates: %s", err)
	}

	resUpdates := make([]*proto.GetIdentityUpdatesResponse_WalletUpdates, len(walletAddresses))
	for i, walletAddress := range walletAddresses {
		walletUpdates := updates[walletAddress]

		resUpdates[i] = &proto.GetIdentityUpdatesResponse_WalletUpdates{
			Updates: []*proto.GetIdentityUpdatesResponse_Update{},
		}

		for _, walletUpdate := range walletUpdates {
			resUpdates[i].Updates = append(resUpdates[i].Updates, buildIdentityUpdate(walletUpdate))
		}
	}

	return &proto.GetIdentityUpdatesResponse{
		Updates: resUpdates,
	}, nil
}

func buildIdentityUpdate(update mlsstore.IdentityUpdate) *proto.GetIdentityUpdatesResponse_Update {
	base := proto.GetIdentityUpdatesResponse_Update{
		TimestampNs: update.TimestampNs,
	}
	switch update.Kind {
	case mlsstore.Create:
		base.Kind = &proto.GetIdentityUpdatesResponse_Update_NewInstallation{
			NewInstallation: &proto.GetIdentityUpdatesResponse_NewInstallationUpdate{
				InstallationId: update.InstallationId,
			},
		}
	case mlsstore.Revoke:
		base.Kind = &proto.GetIdentityUpdatesResponse_Update_RevokedInstallation{
			RevokedInstallation: &proto.GetIdentityUpdatesResponse_RevokedInstallationUpdate{
				InstallationId: update.InstallationId,
			},
		}
	}

	return &base
}

func validatePublishToGroupRequest(req *proto.PublishToGroupRequest) error {
	if req == nil || len(req.Messages) == 0 {
		return status.Errorf(codes.InvalidArgument, "no messages to publish")
	}
	return nil
}

func validatePublishWelcomesRequest(req *proto.PublishWelcomesRequest) error {
	if req == nil || len(req.WelcomeMessages) == 0 {
		return status.Errorf(codes.InvalidArgument, "no welcome messages to publish")
	}
	for _, welcome := range req.WelcomeMessages {
		if welcome == nil || welcome.WelcomeMessage == nil {
			return status.Errorf(codes.InvalidArgument, "invalid welcome message")
		}
		ciphertext := welcome.WelcomeMessage.GetV1().Ciphertext
		if len(ciphertext) == 0 {
			return status.Errorf(codes.InvalidArgument, "invalid welcome message")
		}
	}
	return nil
}

func validateRegisterInstallationRequest(req *proto.RegisterInstallationRequest) error {
	if req == nil || req.LastResortKeyPackage == nil {
		return status.Errorf(codes.InvalidArgument, "no last resort key package")
	}
	return nil
}

func validateUploadKeyPackagesRequest(req *proto.UploadKeyPackagesRequest) error {
	if req == nil || len(req.KeyPackages) == 0 {
		return status.Errorf(codes.InvalidArgument, "no key packages to upload")
	}
	return nil
}

func validateGetIdentityUpdatesRequest(req *proto.GetIdentityUpdatesRequest) error {
	if req == nil || len(req.WalletAddresses) == 0 {
		return status.Errorf(codes.InvalidArgument, "no wallet addresses to get updates for")
	}
	return nil
}

func isReadyToSend(groupId string, message []byte) error {
	if groupId == "" {
		return status.Errorf(codes.InvalidArgument, "group id is empty")
	}
	if len(message) == 0 {
		return status.Errorf(codes.InvalidArgument, "message is empty")
	}
	return nil
}
