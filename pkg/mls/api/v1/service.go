package api

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	v1proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/subscriptions"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"github.com/xmtp/xmtp-node-go/pkg/types"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	pb "google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	// Defines the maximum number of requests we can support per batch.
	// Note: because the client must be aware of these limits, decreasing these values would be a breaking change.
	maxBatchInserts = 10
	maxBatchQueries = 20
)

type Service struct {
	mlsv1.UnimplementedMlsApiServer

	log               *zap.Logger
	writerStore       mlsstore.ReadWriteMlsStore
	readOnlyStore     mlsstore.ReadMlsStore
	validationService mlsvalidate.MLSValidationService

	dbWorker *dbWorker

	subDispatcher *subscriptions.SubscriptionDispatcher

	ctx       context.Context
	ctxCancel func()
}

func NewService(
	log *zap.Logger,
	writerStore mlsstore.ReadWriteMlsStore,
	readOnlyStore mlsstore.ReadMlsStore,
	subDispatcher *subscriptions.SubscriptionDispatcher,
	validationService mlsvalidate.MLSValidationService,
) (s *Service, err error) {
	s = &Service{
		log:               log.Named("mls/v1"),
		writerStore:       writerStore,
		readOnlyStore:     readOnlyStore,
		validationService: validationService,
		subDispatcher:     subDispatcher,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	if s.dbWorker, err = newDBWorker(s.ctx, log, s.readOnlyStore.Queries(), subDispatcher, DEFAULT_POLL_INTERVAL); err != nil {
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

	s.log.Info("closed")
}

/*
*
DEPRECATED: Use UploadKeyPackage instead
*
*/
func (s *Service) RegisterInstallation(
	ctx context.Context,
	req *mlsv1.RegisterInstallationRequest,
) (*mlsv1.RegisterInstallationResponse, error) {
	if err := validateRegisterInstallationRequest(req); err != nil {
		return nil, err
	}

	results, err := s.validationService.ValidateInboxIdKeyPackages(
		ctx,
		[][]byte{req.KeyPackage.KeyPackageTlsSerialized},
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identity: %s", err)
	}

	if len(results) != 1 {
		return nil, status.Errorf(codes.Internal, "unexpected number of results: %d", len(results))
	}

	installationKey := results[0].InstallationKey
	if err = s.writerStore.CreateOrUpdateInstallation(ctx, installationKey, req.KeyPackage.KeyPackageTlsSerialized); err != nil {
		return nil, err
	}
	return &mlsv1.RegisterInstallationResponse{
		InstallationKey: installationKey,
	}, nil
}

func (s *Service) FetchKeyPackages(
	ctx context.Context,
	req *mlsv1.FetchKeyPackagesRequest,
) (*mlsv1.FetchKeyPackagesResponse, error) {
	ids := req.InstallationKeys
	installations, err := s.writerStore.FetchKeyPackages(ctx, ids)
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
			return nil, status.Error(codes.Internal, "could not find key package for installation")
		}

		resPackages[idx] = &mlsv1.FetchKeyPackagesResponse_KeyPackage{
			KeyPackageTlsSerialized: installation.KeyPackage,
		}
	}

	return &mlsv1.FetchKeyPackagesResponse{
		KeyPackages: resPackages,
	}, nil
}

func (s *Service) UploadKeyPackage(
	ctx context.Context,
	req *mlsv1.UploadKeyPackageRequest,
) (res *emptypb.Empty, err error) {
	if err = validateUploadKeyPackageRequest(req); err != nil {
		return nil, err
	}
	// Extract the key packages from the request
	keyPackageBytes := req.KeyPackage.KeyPackageTlsSerialized

	validationResults, err := s.validationService.ValidateInboxIdKeyPackages(
		ctx,
		[][]byte{keyPackageBytes},
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identity: %s", err)
	}

	installationId := validationResults[0].InstallationKey

	if err = s.writerStore.CreateOrUpdateInstallation(ctx, installationId, keyPackageBytes); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert key packages: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) RevokeInstallation(
	ctx context.Context,
	req *mlsv1.RevokeInstallationRequest,
) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetIdentityUpdates(
	ctx context.Context,
	req *mlsv1.GetIdentityUpdatesRequest,
) (res *mlsv1.GetIdentityUpdatesResponse, err error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *Service) SendGroupMessages(
	ctx context.Context,
	req *mlsv1.SendGroupMessagesRequest,
) (res *emptypb.Empty, err error) {
	log := s.log.Named("send-group-messages")
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
			log.Warn("invalid group message", zap.Error(err))
			return nil, err
		}

		// TODO: Wrap this in a transaction so publishing is all or nothing
		decodedGroupId, err := hex.DecodeString(result.GroupId)
		if err != nil {
			log.Warn("invalid group id", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "invalid group id")
		}

		msgV1 := input.GetV1()
		msg, err := s.writerStore.InsertGroupMessage(
			ctx,
			decodedGroupId,
			msgV1.Data,
			msgV1.SenderHmac,
			msgV1.ShouldPush,
			result.IsCommit,
		)
		if err != nil {
			log.Warn("error inserting message", zap.Error(err))
			if mlsstore.IsAlreadyExistsError(err) {
				continue
			}
			return nil, status.Errorf(codes.Internal, "failed to insert message: %s", err)
		}

		metrics.EmitMLSSentGroupMessage(ctx, log, msg)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) SendWelcomeMessages(
	ctx context.Context,
	req *mlsv1.SendWelcomeMessagesRequest,
) (res *emptypb.Empty, err error) {
	log := s.log.Named("send-welcome-messages")
	if err = validateSendWelcomeMessagesRequest(req); err != nil {
		return nil, err
	}

	err = tracing.Wrap(
		ctx,
		log,
		"send-welcome-messages",
		func(ctx context.Context, log *zap.Logger, span tracing.Span) error {
			tracing.SpanTag(span, "message_count", len(req.Messages))

			g, ctx := errgroup.WithContext(ctx)

			for _, input := range req.Messages {
				input := input
				g.Go(func() error {
					insertSpan, insertCtx := tracer.StartSpanFromContext(
						ctx,
						"insert-welcome-message",
					)
					msg, err := s.writerStore.InsertWelcomeMessage(
						insertCtx,
						input.GetV1().InstallationKey,
						input.GetV1().Data,
						input.GetV1().HpkePublicKey,
						types.WrapperAlgorithmFromProto(input.GetV1().WrapperAlgorithm),
						input.GetV1().GetWelcomeMetadata(),
					)
					insertSpan.Finish(tracing.WithError(err))
					if err != nil {
						if mlsstore.IsAlreadyExistsError(err) {
							return nil
						}

						return status.Errorf(codes.Internal, "failed to insert message: %s", err)
					}

					metrics.EmitMLSSentWelcomeMessage(ctx, log, msg)

					return nil
				})
			}

			return g.Wait()
		})
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) QueryGroupMessages(
	ctx context.Context,
	req *mlsv1.QueryGroupMessagesRequest,
) (*mlsv1.QueryGroupMessagesResponse, error) {
	return s.writerStore.QueryGroupMessagesV1(ctx, req)
}

func (s *Service) QueryWelcomeMessages(
	ctx context.Context,
	req *mlsv1.QueryWelcomeMessagesRequest,
) (*mlsv1.QueryWelcomeMessagesResponse, error) {
	return s.writerStore.QueryWelcomeMessagesV1(ctx, req)
}

func (s *Service) SubscribeGroupMessages(
	req *mlsv1.SubscribeGroupMessagesRequest,
	stream mlsv1.MlsApi_SubscribeGroupMessagesServer,
) error {
	log := s.log.Named("subscribe-group-messages").With(zap.Int("filters", len(req.Filters)))
	log.Info("subscription started")
	// Send a header (any header) to fix an issue with Tonic based GRPC clients.
	// See: https://github.com/xmtp/libxmtp/pull/58
	_ = stream.SendHeader(metadata.Pairs("subscribed", "true"))

	var streamLock sync.Mutex
	highWaterMarks := make(map[string]uint64)

	sendToStream := func(msgs []*mlsv1.GroupMessage) error {
		streamLock.Lock()
		defer streamLock.Unlock()

		for _, msg := range msgs {
			groupID := string(msg.GetV1().GroupId)
			if highWaterMarks[groupID] < msg.GetV1().Id {
				highWaterMarks[groupID] = msg.GetV1().Id
			} else {
				continue
			}
			if err := stream.Send(msg); err != nil {
				log.Error("error streaming group message", zap.Error(err))
				return err
			}
		}

		return nil
	}

	errChan := make(chan error, len(req.Filters))

	fetchHistorical := func(filter *mlsv1.SubscribeGroupMessagesRequest_Filter) {
		pagingInfo := &mlsv1.PagingInfo{
			IdCursor:  filter.IdCursor,
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		}

		for {
			select {
			case <-stream.Context().Done():
				return
			case <-s.ctx.Done():
				return
			default:
			}

			if len(errChan) > 0 {
				return
			}

			resp, err := s.readOnlyStore.QueryGroupMessagesV1(
				stream.Context(),
				&mlsv1.QueryGroupMessagesRequest{
					GroupId:    filter.GroupId,
					PagingInfo: pagingInfo,
				},
			)
			if err != nil {
				if err == context.Canceled {
					return
				}
				log.Error("error querying for subscription cursor messages", zap.Error(err))
				errChan <- err
				return
			}

			if err = sendToStream(resp.Messages); err != nil {
				errChan <- err
				return
			}

			if len(resp.Messages) == 0 || resp.PagingInfo == nil || resp.PagingInfo.IdCursor == 0 {
				break
			}
			pagingInfo = resp.PagingInfo
		}
	}

	topicMap := make(map[string]bool)
	for _, filter := range req.Filters {
		topicMap[topic.BuildMLSV1GroupTopic(filter.GroupId)] = true
	}

	// Start the subscription before we start fetching historical messages
	sub := s.subDispatcher.Subscribe(topicMap)
	defer sub.Unsubscribe()

	var wg sync.WaitGroup
	for _, filter := range req.Filters {
		wg.Add(1)
		go func(filter *mlsv1.SubscribeGroupMessagesRequest_Filter) {
			defer wg.Done()
			fetchHistorical(filter)
		}(filter)
	}

	// Wait for all historical messages to be fetched before returning anything from the subscription
	wg.Wait()

	var (
		msg     *mlsv1.GroupMessage
		err     error
		env     *v1proto.Envelope
		subOpen bool
	)

	for {
		select {
		// We received an error when pulling the historical messages
		case err = <-errChan:
			return err
		// The stream is closed
		case <-stream.Context().Done():
			return nil
		// The service is shutting down
		case <-s.ctx.Done():
			return status.Errorf(codes.Unavailable, "service is shutting down")
		// We received a message from the subscription
		case env, subOpen = <-sub.MessagesCh:
			if !subOpen {
				return status.Errorf(codes.Aborted, "caller did not read all messages fast enough")
			}
			msg, err = getGroupMessageFromEnvelope(env)
			if err != nil {
				log.Error("error parsing message", zap.Error(err))
				continue
			}
			if err = sendToStream([]*mlsv1.GroupMessage{msg}); err != nil {
				return err
			}
		}
	}
}

func (s *Service) SubscribeWelcomeMessages(
	req *mlsv1.SubscribeWelcomeMessagesRequest,
	stream mlsv1.MlsApi_SubscribeWelcomeMessagesServer,
) error {
	log := s.log.Named("subscribe-welcome-messages").With(zap.Int("filters", len(req.Filters)))
	log.Info("subscription started")
	// Send a header (any header) to fix an issue with Tonic based GRPC clients.
	// See: https://github.com/xmtp/libxmtp/pull/58
	_ = stream.SendHeader(metadata.Pairs("subscribed", "true"))

	errChan := make(chan error, len(req.Filters))
	highWaterMarks := make(map[string]uint64)
	var streamLock sync.Mutex

	sendError := func(err error) {
		select {
		case errChan <- err:
		default:
		}
	}

	sendToStream := func(msgs []*mlsv1.WelcomeMessage) error {
		streamLock.Lock()
		defer streamLock.Unlock()

		for _, msg := range msgs {
			installationKey := string(msg.GetV1().InstallationKey)
			if highWaterMarks[installationKey] < msg.GetV1().Id {
				highWaterMarks[installationKey] = msg.GetV1().Id
			} else {
				continue
			}
			if err := stream.Send(msg); err != nil {
				log.Error("error streaming welcome message", zap.Error(err))
				return err
			}
		}

		return nil
	}

	fetchHistorical := func(filter *mlsv1.SubscribeWelcomeMessagesRequest_Filter) {
		pagingInfo := &mlsv1.PagingInfo{
			IdCursor:  filter.IdCursor,
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		}

		for {
			select {
			case <-stream.Context().Done():
				return
			case <-s.ctx.Done():
				return
			default:
			}

			if len(errChan) > 0 {
				return
			}

			resp, err := s.readOnlyStore.QueryWelcomeMessagesV1(
				stream.Context(),
				&mlsv1.QueryWelcomeMessagesRequest{
					InstallationKey: filter.InstallationKey,
					PagingInfo:      pagingInfo,
				},
			)
			if err != nil {
				if err == context.Canceled {
					return
				}
				log.Error("error querying for subscription cursor messages", zap.Error(err))
				sendError(err)
				return
			}

			if err = sendToStream(resp.Messages); err != nil {
				sendError(err)
				return
			}

			if len(resp.Messages) == 0 || resp.PagingInfo == nil || resp.PagingInfo.IdCursor == 0 {
				break
			}
			pagingInfo = resp.PagingInfo
		}
	}

	topicMap := make(map[string]bool)
	for _, filter := range req.Filters {
		topicMap[topic.BuildMLSV1WelcomeTopic(filter.InstallationKey)] = true
	}

	// Start the subscription before we start fetching historical messages
	sub := s.subDispatcher.Subscribe(topicMap)
	defer sub.Unsubscribe()

	var wg sync.WaitGroup
	for _, filter := range req.Filters {
		wg.Add(1)
		go func(filter *mlsv1.SubscribeWelcomeMessagesRequest_Filter) {
			defer wg.Done()
			fetchHistorical(filter)
		}(filter)
	}

	// Wait for all historical messages to be fetched before returning anything from the subscription
	wg.Wait()

	var (
		msg     *mlsv1.WelcomeMessage
		err     error
		env     *v1proto.Envelope
		subOpen bool
	)

	for {
		select {
		// We received an error when pulling the historical messages
		case err = <-errChan:
			return err
		// The stream is closed
		case <-stream.Context().Done():
			return nil
		// The service is shutting down
		case <-s.ctx.Done():
			return status.Errorf(codes.Unavailable, "service is shutting down")
		// We received a message from the subscription
		case env, subOpen = <-sub.MessagesCh:
			if !subOpen {
				return nil
			}
			msg, err = getWelcomeMessageFromEnvelope(env)
			if err != nil {
				log.Error("error parsing message", zap.Error(err))
				continue
			}
			if err = sendToStream([]*mlsv1.WelcomeMessage{msg}); err != nil {
				return err
			}
		}
	}
}

func (s *Service) BatchPublishCommitLog(
	ctx context.Context,
	req *mlsv1.BatchPublishCommitLogRequest,
) (*emptypb.Empty, error) {
	log := s.log.Named("batch-publish-commit-log")
	if err := validateBatchPublishCommitLogRequest(req); err != nil {
		return nil, err
	}

	for _, entry := range req.Requests {
		if entry == nil ||
			entry.GroupId == nil || len(entry.GroupId) == 0 ||
			entry.SerializedCommitLogEntry == nil || len(entry.SerializedCommitLogEntry) == 0 ||
			entry.Signature == nil {
			return nil, status.Error(codes.InvalidArgument, "invalid commit log entry")
		}
		serializedSignature, err := pb.Marshal(entry.Signature)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid signature")
		}
		inserted, err := s.writerStore.InsertCommitLog(
			ctx,
			entry.GroupId,
			entry.SerializedCommitLogEntry,
			serializedSignature,
		)
		if err != nil {
			log.Error("error inserting commit log", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to insert commit log: %s", err)
		}

		metrics.EmitMLSPublishedCommitLogEntry(ctx, log, &inserted)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) BatchQueryCommitLog(
	ctx context.Context,
	req *mlsv1.BatchQueryCommitLogRequest,
) (*mlsv1.BatchQueryCommitLogResponse, error) {
	log := s.log.Named("batch-query-commit-log")
	if err := validateBatchQueryCommitLogRequest(req); err != nil {
		return nil, err
	}

	out := make([]*mlsv1.QueryCommitLogResponse, len(req.Requests))
	for idx, request := range req.Requests {
		if request == nil || request.GroupId == nil || len(request.GroupId) == 0 {
			return nil, status.Error(codes.InvalidArgument, "invalid request")
		}
		response, err := s.writerStore.QueryCommitLog(ctx, request)
		if err != nil {
			log.Error("error querying commit log", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to query commit log: %s", err)
		}
		out[idx] = response
	}
	return &mlsv1.BatchQueryCommitLogResponse{
		Responses: out,
	}, nil
}

func validateSendGroupMessagesRequest(req *mlsv1.SendGroupMessagesRequest) error {
	if req == nil || len(req.Messages) == 0 {
		return status.Error(codes.InvalidArgument, "no group messages to send")
	}
	for _, input := range req.Messages {
		if input == nil || input.GetV1() == nil {
			return status.Error(codes.InvalidArgument, "invalid group message")
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
			return status.Error(codes.InvalidArgument, "invalid welcome message")
		}

		v1 := input.GetV1()
		if len(v1.Data) == 0 || len(v1.InstallationKey) == 0 || len(v1.HpkePublicKey) == 0 {
			return status.Error(codes.InvalidArgument, "invalid welcome message")
		}
	}
	return nil
}

func validateRegisterInstallationRequest(req *mlsv1.RegisterInstallationRequest) error {
	if req == nil || req.KeyPackage == nil {
		return status.Error(codes.InvalidArgument, "no key package")
	}
	return nil
}

func validateUploadKeyPackageRequest(req *mlsv1.UploadKeyPackageRequest) error {
	if req == nil || req.KeyPackage == nil {
		return status.Error(codes.InvalidArgument, "no key package")
	}
	return nil
}

func validateBatchPublishCommitLogRequest(req *mlsv1.BatchPublishCommitLogRequest) error {
	if req == nil || len(req.Requests) == 0 {
		return status.Error(codes.InvalidArgument, "no log entries to publish")
	}
	if len(req.Requests) > maxBatchInserts {
		return status.Errorf(
			codes.InvalidArgument,
			"cannot exceed %d inserts in single batch",
			maxBatchInserts,
		)
	}
	return nil
}

func validateBatchQueryCommitLogRequest(req *mlsv1.BatchQueryCommitLogRequest) error {
	if req == nil || len(req.Requests) == 0 {
		return status.Error(codes.InvalidArgument, "no requests to query")
	}
	if len(req.Requests) > maxBatchQueries {
		return status.Errorf(
			codes.InvalidArgument,
			"cannot exceed %d queries in single batch",
			maxBatchQueries,
		)
	}
	return nil
}

func requireReadyToSend(groupId string, message []byte) error {
	if len(groupId) == 0 {
		return status.Error(codes.InvalidArgument, "group id is empty")
	}
	if len(message) == 0 {
		return status.Error(codes.InvalidArgument, "message is empty")
	}
	return nil
}

func getGroupMessageFromEnvelope(env *v1proto.Envelope) (*mlsv1.GroupMessage, error) {
	var msg mlsv1.GroupMessage
	err := pb.Unmarshal(env.Message, &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func getWelcomeMessageFromEnvelope(env *v1proto.Envelope) (*mlsv1.WelcomeMessage, error) {
	var msg mlsv1.WelcomeMessage
	err := pb.Unmarshal(env.Message, &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}
