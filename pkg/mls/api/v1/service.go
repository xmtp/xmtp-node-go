package api

import (
	"context"
	"encoding/hex"

	wakunode "github.com/waku-org/go-waku/waku/v2/node"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	proto "github.com/xmtp/proto/v3/go/mls/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store"
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

func (s *Service) SendGroupMessages(ctx context.Context, req *proto.SendGroupMessagesRequest) (res *emptypb.Empty, err error) {
	if err = validateSendGroupMessagesRequest(req); err != nil {
		return nil, err
	}

	validationResults, err := s.validationService.ValidateGroupMessages(ctx, req.Messages)
	if err != nil {
		// TODO: Separate validation errors from internal errors
		return nil, status.Errorf(codes.InvalidArgument, "invalid group message: %s", err)
	}

	for i, result := range validationResults {
		input := req.Messages[i]

		if err = requireReadyToSend(result.GroupId, input.GetV1().Data); err != nil {
			return nil, err
		}

		// TODO: Wrap this in a transaction so publishing is all or nothing
		decodedGroupId, err := hex.DecodeString(result.GroupId)
		if err != nil {
			return nil, err
		}
		msg, err := s.store.InsertGroupMessage(ctx, decodedGroupId, input.GetV1().Data)
		if err != nil {
			if store.IsAlreadyExistsError(err) {
				continue
			}
			return nil, status.Errorf(codes.Internal, "failed to insert message: %s", err)
		}

		wakuTopic := topic.BuildGroupTopic(decodedGroupId)
		_, err = s.waku.Relay().Publish(ctx, &wakupb.WakuMessage{
			ContentTopic: wakuTopic,
			Timestamp:    msg.CreatedAt.UnixNano(),
			Payload:      msg.Data,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to publish message: %s", err)
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) SendWelcomeMessages(ctx context.Context, req *proto.SendWelcomeMessagesRequest) (res *emptypb.Empty, err error) {
	if err = validateSendWelcomeMessagesRequest(req); err != nil {
		return nil, err
	}

	// TODO: Wrap this in a transaction so publishing is all or nothing
	for _, input := range req.Messages {
		msg, err := s.store.InsertWelcomeMessage(ctx, input.GetV1().InstallationId, input.GetV1().Data)
		if err != nil {
			if store.IsAlreadyExistsError(err) {
				continue
			}
			return nil, status.Errorf(codes.Internal, "failed to insert message: %s", err)
		}

		wakuTopic := topic.BuildWelcomeTopic(input.GetV1().InstallationId)
		_, err = s.waku.Relay().Publish(ctx, &wakupb.WakuMessage{
			ContentTopic: wakuTopic,
			Timestamp:    msg.CreatedAt.UnixNano(),
			Payload:      msg.Data,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to publish message: %s", err)
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) QueryGroupMessages(ctx context.Context, req *proto.QueryGroupMessagesRequest) (*proto.QueryGroupMessagesResponse, error) {
	return s.store.QueryGroupMessagesV1(ctx, req)
}

func (s *Service) QueryWelcomeMessages(ctx context.Context, req *proto.QueryWelcomeMessagesRequest) (*proto.QueryWelcomeMessagesResponse, error) {
	return s.store.QueryWelcomeMessagesV1(ctx, req)
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

func validateSendGroupMessagesRequest(req *proto.SendGroupMessagesRequest) error {
	if req == nil || len(req.Messages) == 0 {
		return status.Errorf(codes.InvalidArgument, "no group messages to send")
	}
	for _, input := range req.Messages {
		if input == nil || input.GetV1() == nil {
			return status.Errorf(codes.InvalidArgument, "invalid group message")
		}
	}
	return nil
}

func validateSendWelcomeMessagesRequest(req *proto.SendWelcomeMessagesRequest) error {
	if req == nil || len(req.Messages) == 0 {
		return status.Errorf(codes.InvalidArgument, "no welcome messages to send")
	}
	for _, input := range req.Messages {
		if input == nil || input.GetV1() == nil {
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
	if len(groupId) == 0 {
		return status.Errorf(codes.InvalidArgument, "group id is empty")
	}
	if len(message) == 0 {
		return status.Errorf(codes.InvalidArgument, "message is empty")
	}
	return nil
}
