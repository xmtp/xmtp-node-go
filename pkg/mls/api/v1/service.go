package api

import (
	"context"

	wakunode "github.com/waku-org/go-waku/waku/v2/node"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	proto "github.com/xmtp/proto/v3/go/mls/api/v1"
	"github.com/xmtp/proto/v3/go/mls/message_contents"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
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
	store             mlsstore.MlsStore
	validationService mlsvalidate.MLSValidationService
}

func NewService(node *wakunode.WakuNode, logger *zap.Logger, store mlsstore.MlsStore, validationService mlsvalidate.MLSValidationService) (s *Service, err error) {
	s = &Service{
		log:               logger.Named("mls/v1"),
		waku:              node,
		store:             store,
		validationService: validationService,
	}

	s.log.Info("Starting MLS service")
	return s, nil
}

func (s *Service) RegisterInstallation(ctx context.Context, req *proto.RegisterInstallationRequest) (*proto.RegisterInstallationResponse, error) {
	if err := validateRegisterInstallationRequest(req); err != nil {
		return nil, err
	}

	results, err := s.validationService.ValidateKeyPackages(ctx, [][]byte{req.KeyPackage.KeyPackageTlsSerialized})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identity: %s", err)
	}
	if len(results) != 1 {
		return nil, status.Errorf(codes.Internal, "unexpected number of results: %d", len(results))
	}

	installationId := results[0].InstallationId
	accountAddress := results[0].AccountAddress
	credentialIdentity := results[0].CredentialIdentity

	if err = s.store.CreateInstallation(ctx, installationId, accountAddress, credentialIdentity, req.KeyPackage.KeyPackageTlsSerialized, results[0].Expiration); err != nil {
		return nil, err
	}

	return &proto.RegisterInstallationResponse{
		InstallationId: installationId,
	}, nil
}

func (s *Service) FetchKeyPackages(ctx context.Context, req *proto.FetchKeyPackagesRequest) (*proto.FetchKeyPackagesResponse, error) {
	ids := req.InstallationIds
	installations, err := s.store.FetchKeyPackages(ctx, ids)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch key packages: %s", err)
	}
	keyPackageMap := make(map[string]int)
	for idx, id := range ids {
		keyPackageMap[string(id)] = idx
	}

	resPackages := make([]*proto.FetchKeyPackagesResponse_KeyPackage, len(ids))
	for _, installation := range installations {

		idx, ok := keyPackageMap[string(installation.ID)]
		if !ok {
			return nil, status.Errorf(codes.Internal, "could not find key package for installation")
		}

		resPackages[idx] = &proto.FetchKeyPackagesResponse_KeyPackage{
			KeyPackageTlsSerialized: installation.KeyPackage,
		}
	}

	return &proto.FetchKeyPackagesResponse{
		KeyPackages: resPackages,
	}, nil
}

func (s *Service) UploadKeyPackage(ctx context.Context, req *proto.UploadKeyPackageRequest) (res *emptypb.Empty, err error) {
	if err = validateUploadKeyPackageRequest(req); err != nil {
		return nil, err
	}
	// Extract the key packages from the request
	keyPackageBytes := req.KeyPackage.KeyPackageTlsSerialized

	validationResults, err := s.validationService.ValidateKeyPackages(ctx, [][]byte{keyPackageBytes})
	if err != nil {
		// TODO: Differentiate between validation errors and internal errors
		return nil, status.Errorf(codes.InvalidArgument, "invalid identity: %s", err)
	}
	installationId := validationResults[0].InstallationId
	expiration := validationResults[0].Expiration

	if err = s.store.UpdateKeyPackage(ctx, installationId, keyPackageBytes, expiration); err != nil {
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

	accountAddresses := req.AccountAddresses
	updates, err := s.store.GetIdentityUpdates(ctx, req.AccountAddresses, int64(req.StartTimeNs))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get identity updates: %s", err)
	}

	resUpdates := make([]*proto.GetIdentityUpdatesResponse_WalletUpdates, len(accountAddresses))
	for i, accountAddress := range accountAddresses {
		walletUpdates := updates[accountAddress]

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

func (s *Service) PublishGroupMessages(ctx context.Context, req *proto.PublishGroupMessagesRequest) (res *emptypb.Empty, err error) {
	if err = validatePublishGroupMessagesRequest(req); err != nil {
		return nil, err
	}

	messages := make([][]byte, len(req.Messages))
	for i, message := range req.Messages {
		v1 := message.GetV1()
		if v1 == nil {
			return nil, status.Errorf(codes.InvalidArgument, "message must be v1")
		}
		messages[i] = v1.Data
	}

	validationResults, err := s.validationService.ValidateGroupMessages(ctx, messages)
	if err != nil {
		// TODO: Separate validation errors from internal errors
		return nil, status.Errorf(codes.InvalidArgument, "invalid group message: %s", err)
	}

	for i, result := range validationResults {
		data := messages[i]

		if err = requireReadyToSend(result.GroupId, data); err != nil {
			return nil, err
		}

		// TODO: Wrap this in a transaction so publishing is all or nothing
		wakuTopic := topic.BuildGroupTopic(result.GroupId)
		msg, err := s.store.InsertGroupMessage(ctx, result.GroupId, data)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to insert message: %s", err)
		}

		_, err = s.waku.Relay().Publish(ctx, &wakupb.WakuMessage{
			ContentTopic: wakuTopic,
			Timestamp:    msg.CreatedAt.UnixNano(),
			Payload:      data,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to publish message: %s", err)
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) PublishWelcomeMessages(ctx context.Context, req *proto.PublishWelcomeMessagesRequest) (res *emptypb.Empty, err error) {
	if err = validatePublishWelcomeMessagesRequest(req); err != nil {
		return nil, err
	}

	// TODO: Wrap this in a transaction so publishing is all or nothing
	for _, welcome := range req.WelcomeMessages {
		wakuTopic := topic.BuildWelcomeTopic(welcome.InstallationId)
		data := welcome.WelcomeMessage.GetV1().Data

		msg, err := s.store.InsertWelcomeMessage(ctx, string(welcome.InstallationId), data)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to insert message: %s", err)
		}

		_, err = s.waku.Relay().Publish(ctx, &wakupb.WakuMessage{
			ContentTopic: wakuTopic,
			Timestamp:    msg.CreatedAt.UnixNano(),
			Payload:      data,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to publish message: %s", err)
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) QueryGroupMessages(ctx context.Context, req *proto.QueryGroupMessagesRequest) (*proto.QueryGroupMessagesResponse, error) {
	msgs, err := s.store.QueryGroupMessages(ctx, req)
	if err != nil {
		return nil, err
	}

	messages := make([]*message_contents.GroupMessage, 0, len(msgs))
	for _, msg := range msgs {
		messages = append(messages, &message_contents.GroupMessage{
			Version: &message_contents.GroupMessage_V1_{
				V1: &message_contents.GroupMessage_V1{
					Id:        msg.Id,
					CreatedNs: uint64(msg.CreatedAt.UnixNano()),
					GroupId:   msg.GroupId,
					Data:      msg.Data,
				},
			},
		})
	}

	// TODO(snormore): build and return paging info

	return &proto.QueryGroupMessagesResponse{
		Messages:   messages,
		PagingInfo: nil,
	}, nil
}

func (s *Service) QueryWelcomeMessages(ctx context.Context, req *proto.QueryWelcomeMessagesRequest) (*proto.QueryWelcomeMessagesResponse, error) {
	msgs, err := s.store.QueryWelcomeMessages(ctx, req)
	if err != nil {
		return nil, err
	}

	messages := make([]*message_contents.WelcomeMessage, 0, len(msgs))
	for _, msg := range msgs {
		messages = append(messages, &message_contents.WelcomeMessage{
			Version: &message_contents.WelcomeMessage_V1_{
				V1: &message_contents.WelcomeMessage_V1{
					Id:             msg.Id,
					CreatedNs:      uint64(msg.CreatedAt.UnixNano()),
					InstallationId: msg.InstallationId,
					Data:           msg.Data,
				},
			},
		})
	}

	// TODO(snormore): build and return paging info

	return &proto.QueryWelcomeMessagesResponse{
		Messages:   messages,
		PagingInfo: nil,
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
				InstallationId:     update.InstallationId,
				CredentialIdentity: update.CredentialIdentity,
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

func validatePublishGroupMessagesRequest(req *proto.PublishGroupMessagesRequest) error {
	if req == nil || len(req.Messages) == 0 {
		return status.Errorf(codes.InvalidArgument, "no messages to publish")
	}
	return nil
}

func validatePublishWelcomeMessagesRequest(req *proto.PublishWelcomeMessagesRequest) error {
	if req == nil || len(req.WelcomeMessages) == 0 {
		return status.Errorf(codes.InvalidArgument, "no welcome messages to publish")
	}
	for _, welcome := range req.WelcomeMessages {
		if welcome == nil || welcome.WelcomeMessage == nil {
			return status.Errorf(codes.InvalidArgument, "invalid welcome message")
		}

		v1 := welcome.WelcomeMessage.GetV1()
		if v1 == nil {
			return status.Errorf(codes.InvalidArgument, "invalid welcome message")
		}
	}
	return nil
}

func validateRegisterInstallationRequest(req *proto.RegisterInstallationRequest) error {
	if req == nil || req.KeyPackage == nil {
		return status.Errorf(codes.InvalidArgument, "no key package")
	}
	return nil
}

func validateUploadKeyPackageRequest(req *proto.UploadKeyPackageRequest) error {
	if req == nil || req.KeyPackage == nil {
		return status.Errorf(codes.InvalidArgument, "no key package")
	}
	return nil
}

func validateGetIdentityUpdatesRequest(req *proto.GetIdentityUpdatesRequest) error {
	if req == nil || len(req.AccountAddresses) == 0 {
		return status.Errorf(codes.InvalidArgument, "no wallet addresses to get updates for")
	}
	return nil
}

func requireReadyToSend(groupId string, message []byte) error {
	if groupId == "" {
		return status.Errorf(codes.InvalidArgument, "group id is empty")
	}
	if len(message) == 0 {
		return status.Errorf(codes.InvalidArgument, "message is empty")
	}
	return nil
}
