package api

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	mlsv1 "github.com/xmtp/proto/v3/go/mls/api/v1"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	pb "google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	mlsv1.UnimplementedMlsApiServer

	log               *zap.Logger
	store             mlsstore.MlsStore
	validationService mlsvalidate.MLSValidationService

	publishToWakuRelay func(context.Context, *wakupb.WakuMessage) error

	ns *server.Server
	nc *nats.Conn

	ctx       context.Context
	ctxCancel func()
}

func NewService(log *zap.Logger, store mlsstore.MlsStore, validationService mlsvalidate.MLSValidationService, publishToWakuRelay func(context.Context, *wakupb.WakuMessage) error) (s *Service, err error) {
	s = &Service{
		log:                log.Named("mls/v1"),
		store:              store,
		validationService:  validationService,
		publishToWakuRelay: publishToWakuRelay,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	// Initialize nats for subscriptions.
	s.ns, err = server.NewServer(&server.Options{
		Port: server.RANDOM_PORT,
	})
	if err != nil {
		return nil, err
	}
	go s.ns.Start()
	if !s.ns.ReadyForConnections(4 * time.Second) {
		return nil, errors.New("nats not ready")
	}
	s.nc, err = nats.Connect(s.ns.ClientURL())
	if err != nil {
		return nil, err
	}

	s.log.Info("Starting MLS service")
	return s, nil
}

func (s *Service) Close() {
	s.log.Info("closing")

	if s.ctxCancel != nil {
		s.ctxCancel()
	}

	if s.nc != nil {
		s.nc.Close()
	}
	if s.ns != nil {
		s.ns.Shutdown()
	}

	s.log.Info("closed")
}

func (s *Service) HandleIncomingWakuRelayMessage(wakuMsg *wakupb.WakuMessage) error {
	if topic.IsMLSV1Group(wakuMsg.ContentTopic) {
		var msg mlsv1.GroupMessage
		err := pb.Unmarshal(wakuMsg.Payload, &msg)
		if err != nil {
			return err
		}
		if msg.GetV1() == nil {
			return nil
		}
		err = s.nc.Publish(buildNatsSubjectForGroupMessages(msg.GetV1().GroupId), wakuMsg.Payload)
		if err != nil {
			return err
		}
	} else if topic.IsMLSV1Welcome(wakuMsg.ContentTopic) {
		var msg mlsv1.WelcomeMessage
		err := pb.Unmarshal(wakuMsg.Payload, &msg)
		if err != nil {
			return err
		}
		if msg.GetV1() == nil {
			return nil
		}
		err = s.nc.Publish(buildNatsSubjectForWelcomeMessages(msg.GetV1().InstallationKey), wakuMsg.Payload)
		if err != nil {
			return err
		}
	} else {
		s.log.Info("received unknown mls message type from waku relay", zap.String("topic", wakuMsg.ContentTopic))
	}

	return nil
}

func (s *Service) RegisterInstallation(ctx context.Context, req *mlsv1.RegisterInstallationRequest) (*mlsv1.RegisterInstallationResponse, error) {
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

	installationId := results[0].InstallationKey
	accountAddress := results[0].AccountAddress
	credentialIdentity := results[0].CredentialIdentity

	if err = s.store.CreateInstallation(ctx, installationId, accountAddress, credentialIdentity, req.KeyPackage.KeyPackageTlsSerialized, results[0].Expiration); err != nil {
		return nil, err
	}

	return &mlsv1.RegisterInstallationResponse{
		InstallationKey: installationId,
	}, nil
}

func (s *Service) FetchKeyPackages(ctx context.Context, req *mlsv1.FetchKeyPackagesRequest) (*mlsv1.FetchKeyPackagesResponse, error) {
	ids := req.InstallationKeys
	installations, err := s.store.FetchKeyPackages(ctx, ids)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch key packages: %s", err)
	}
	keyPackageMap := make(map[string]int)
	for idx, id := range ids {
		keyPackageMap[string(id)] = idx
	}

	resPackages := make([]*mlsv1.FetchKeyPackagesResponse_KeyPackage, len(ids))
	for _, installation := range installations {

		idx, ok := keyPackageMap[string(installation.ID)]
		if !ok {
			return nil, status.Errorf(codes.Internal, "could not find key package for installation")
		}

		resPackages[idx] = &mlsv1.FetchKeyPackagesResponse_KeyPackage{
			KeyPackageTlsSerialized: installation.KeyPackage,
		}
	}

	return &mlsv1.FetchKeyPackagesResponse{
		KeyPackages: resPackages,
	}, nil
}

func (s *Service) UploadKeyPackage(ctx context.Context, req *mlsv1.UploadKeyPackageRequest) (res *emptypb.Empty, err error) {
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
	installationId := validationResults[0].InstallationKey
	expiration := validationResults[0].Expiration

	if err = s.store.UpdateKeyPackage(ctx, installationId, keyPackageBytes, expiration); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert key packages: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) RevokeInstallation(ctx context.Context, req *mlsv1.RevokeInstallationRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetIdentityUpdates(ctx context.Context, req *mlsv1.GetIdentityUpdatesRequest) (res *mlsv1.GetIdentityUpdatesResponse, err error) {
	if err = validateGetIdentityUpdatesRequest(req); err != nil {
		return nil, err
	}

	accountAddresses := req.AccountAddresses
	updates, err := s.store.GetIdentityUpdates(ctx, req.AccountAddresses, int64(req.StartTimeNs))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get identity updates: %s", err)
	}

	resUpdates := make([]*mlsv1.GetIdentityUpdatesResponse_WalletUpdates, len(accountAddresses))
	for i, accountAddress := range accountAddresses {
		walletUpdates := updates[accountAddress]

		resUpdates[i] = &mlsv1.GetIdentityUpdatesResponse_WalletUpdates{
			Updates: []*mlsv1.GetIdentityUpdatesResponse_Update{},
		}

		for _, walletUpdate := range walletUpdates {
			resUpdates[i].Updates = append(resUpdates[i].Updates, buildIdentityUpdate(walletUpdate))
		}
	}

	return &mlsv1.GetIdentityUpdatesResponse{
		Updates: resUpdates,
	}, nil
}

func (s *Service) SendGroupMessages(ctx context.Context, req *mlsv1.SendGroupMessagesRequest) (res *emptypb.Empty, err error) {
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
			return nil, status.Error(codes.InvalidArgument, "invalid group id")
		}
		msg, err := s.store.InsertGroupMessage(ctx, decodedGroupId, input.GetV1().Data)
		if err != nil {
			if mlsstore.IsAlreadyExistsError(err) {
				continue
			}
			return nil, status.Errorf(codes.Internal, "failed to insert message: %s", err)
		}

		msgB, err := pb.Marshal(&mlsv1.GroupMessage{
			Version: &mlsv1.GroupMessage_V1_{
				V1: &mlsv1.GroupMessage_V1{
					Id:        msg.Id,
					CreatedNs: uint64(msg.CreatedAt.UnixNano()),
					GroupId:   msg.GroupId,
					Data:      msg.Data,
				},
			},
		})
		if err != nil {
			return nil, err
		}

		err = s.publishToWakuRelay(ctx, &wakupb.WakuMessage{
			ContentTopic: topic.BuildMLSV1GroupTopic(decodedGroupId),
			Timestamp:    msg.CreatedAt.UnixNano(),
			Payload:      msgB,
		})
		if err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) SendWelcomeMessages(ctx context.Context, req *mlsv1.SendWelcomeMessagesRequest) (res *emptypb.Empty, err error) {
	if err = validateSendWelcomeMessagesRequest(req); err != nil {
		return nil, err
	}

	// TODO: Wrap this in a transaction so publishing is all or nothing
	for _, input := range req.Messages {
		msg, err := s.store.InsertWelcomeMessage(ctx, input.GetV1().InstallationKey, input.GetV1().Data)
		if err != nil {
			if mlsstore.IsAlreadyExistsError(err) {
				continue
			}
			return nil, status.Errorf(codes.Internal, "failed to insert message: %s", err)
		}

		msgB, err := pb.Marshal(&mlsv1.WelcomeMessage{
			Version: &mlsv1.WelcomeMessage_V1_{
				V1: &mlsv1.WelcomeMessage_V1{
					Id:              msg.Id,
					CreatedNs:       uint64(msg.CreatedAt.UnixNano()),
					InstallationKey: msg.InstallationKey,
					Data:            msg.Data,
				},
			},
		})
		if err != nil {
			return nil, err
		}

		err = s.publishToWakuRelay(ctx, &wakupb.WakuMessage{
			ContentTopic: topic.BuildMLSV1WelcomeTopic(input.GetV1().InstallationKey),
			Timestamp:    msg.CreatedAt.UnixNano(),
			Payload:      msgB,
		})
		if err != nil {
			return nil, err
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) QueryGroupMessages(ctx context.Context, req *mlsv1.QueryGroupMessagesRequest) (*mlsv1.QueryGroupMessagesResponse, error) {
	return s.store.QueryGroupMessagesV1(ctx, req)
}

func (s *Service) QueryWelcomeMessages(ctx context.Context, req *mlsv1.QueryWelcomeMessagesRequest) (*mlsv1.QueryWelcomeMessagesResponse, error) {
	return s.store.QueryWelcomeMessagesV1(ctx, req)
}

func (s *Service) SubscribeGroupMessages(req *mlsv1.SubscribeGroupMessagesRequest, stream mlsv1.MlsApi_SubscribeGroupMessagesServer) error {
	log := s.log.Named("subscribe-group-messages").With(zap.Int("filters", len(req.Filters)))

	// Send a header (any header) to fix an issue with Tonic based GRPC clients.
	// See: https://github.com/xmtp/libxmtp/pull/58
	_ = stream.SendHeader(metadata.Pairs("subscribed", "true"))

	var streamLock sync.Mutex
	for _, filter := range req.Filters {
		natsSubject := buildNatsSubjectForGroupMessages(filter.GroupId)
		sub, err := s.nc.Subscribe(natsSubject, func(natsMsg *nats.Msg) {
			var msg mlsv1.GroupMessage
			err := pb.Unmarshal(natsMsg.Data, &msg)
			if err != nil {
				log.Error("parsing group message from bytes", zap.Error(err))
				return
			}
			func() {
				streamLock.Lock()
				defer streamLock.Unlock()
				err := stream.Send(&msg)
				if err != nil {
					log.Error("sending group message to subscribe", zap.Error(err))
				}
			}()
		})
		if err != nil {
			log.Error("error subscribing to group messages", zap.Error(err))
			return err
		}
		defer func() {
			_ = sub.Unsubscribe()
		}()
	}

	select {
	case <-stream.Context().Done():
		return nil
	case <-s.ctx.Done():
		return nil
	}
}

func (s *Service) SubscribeWelcomeMessages(req *mlsv1.SubscribeWelcomeMessagesRequest, stream mlsv1.MlsApi_SubscribeWelcomeMessagesServer) error {
	log := s.log.Named("subscribe-welcome-messages").With(zap.Int("filters", len(req.Filters)))

	// Send a header (any header) to fix an issue with Tonic based GRPC clients.
	// See: https://github.com/xmtp/libxmtp/pull/58
	_ = stream.SendHeader(metadata.Pairs("subscribed", "true"))

	var streamLock sync.Mutex
	for _, filter := range req.Filters {
		natsSubject := buildNatsSubjectForWelcomeMessages(filter.InstallationKey)
		sub, err := s.nc.Subscribe(natsSubject, func(natsMsg *nats.Msg) {
			var msg mlsv1.WelcomeMessage
			err := pb.Unmarshal(natsMsg.Data, &msg)
			if err != nil {
				log.Error("parsing welcome message from bytes", zap.Error(err))
				return
			}
			func() {
				streamLock.Lock()
				defer streamLock.Unlock()
				err := stream.Send(&msg)
				if err != nil {
					log.Error("sending welcome message to subscribe", zap.Error(err))
				}
			}()
		})
		if err != nil {
			log.Error("error subscribing to welcome messages", zap.Error(err))
			return err
		}
		defer func() {
			_ = sub.Unsubscribe()
		}()
	}

	select {
	case <-stream.Context().Done():
		return nil
	case <-s.ctx.Done():
		return nil
	}
}

func buildNatsSubjectForGroupMessages(groupId []byte) string {
	hasher := fnv.New64a()
	hasher.Write(groupId)
	return fmt.Sprintf("gm-%x", hasher.Sum64())
}

func buildNatsSubjectForWelcomeMessages(installationId []byte) string {
	hasher := fnv.New64a()
	hasher.Write(installationId)
	return fmt.Sprintf("wm-%x", hasher.Sum64())
}

func buildIdentityUpdate(update mlsstore.IdentityUpdate) *mlsv1.GetIdentityUpdatesResponse_Update {
	base := mlsv1.GetIdentityUpdatesResponse_Update{
		TimestampNs: update.TimestampNs,
	}
	switch update.Kind {
	case mlsstore.Create:
		base.Kind = &mlsv1.GetIdentityUpdatesResponse_Update_NewInstallation{
			NewInstallation: &mlsv1.GetIdentityUpdatesResponse_NewInstallationUpdate{
				InstallationKey:    update.InstallationKey,
				CredentialIdentity: update.CredentialIdentity,
			},
		}
	case mlsstore.Revoke:
		base.Kind = &mlsv1.GetIdentityUpdatesResponse_Update_RevokedInstallation{
			RevokedInstallation: &mlsv1.GetIdentityUpdatesResponse_RevokedInstallationUpdate{
				InstallationKey: update.InstallationKey,
			},
		}
	}

	return &base
}

func validateSendGroupMessagesRequest(req *mlsv1.SendGroupMessagesRequest) error {
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

func validateSendWelcomeMessagesRequest(req *mlsv1.SendWelcomeMessagesRequest) error {
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

func validateRegisterInstallationRequest(req *mlsv1.RegisterInstallationRequest) error {
	if req == nil || req.KeyPackage == nil {
		return status.Errorf(codes.InvalidArgument, "no key package")
	}
	return nil
}

func validateUploadKeyPackageRequest(req *mlsv1.UploadKeyPackageRequest) error {
	if req == nil || req.KeyPackage == nil {
		return status.Errorf(codes.InvalidArgument, "no key package")
	}
	return nil
}

func validateGetIdentityUpdatesRequest(req *mlsv1.GetIdentityUpdatesRequest) error {
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
