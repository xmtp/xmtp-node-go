package api

import (
	"context"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/xmtp/xmtp-node-go/pkg/envelopes"
	identityTypes "github.com/xmtp/xmtp-node-go/pkg/identity/types"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	api "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	v1proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	pb "google.golang.org/protobuf/proto"
)

type Service struct {
	api.UnimplementedIdentityApiServer

	log               *zap.Logger
	store             mlsstore.MlsStore
	validationService mlsvalidate.MLSValidationService

	ctx                context.Context
	nc                 *nats.Conn
	publishToWakuRelay func(context.Context, *wakupb.WakuMessage) error
	ctxCancel          func()
}

func NewService(
	log *zap.Logger,
	store mlsstore.MlsStore,
	validationService mlsvalidate.MLSValidationService,
	natsServer *server.Server,
	publishToWakuRelay func(context.Context, *wakupb.WakuMessage) error,
) (s *Service, err error) {
	s = &Service{
		log:                log.Named("identity"),
		store:              store,
		validationService:  validationService,
		publishToWakuRelay: publishToWakuRelay,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	s.nc, err = nats.Connect(natsServer.ClientURL())
	if err != nil {
		return nil, err
	}

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

/*
Properties we want on the inbox log:
 1. Updates come in and are assigned sequence numbers in some order
 2. Updates are not visible to API consumers until they have been
    validated and the address_log table has been updated
 3. If you read once, and then read again:
    a. The second read must have all of the updates from the first read
    b. New updates from the second read *cannot* have a lower sequence
    number than the latest sequence number from the first read
    c. This only applies to reads *on a given inbox*. Ordering between
    different inboxes does not matter.

For the address log, strict ordering/strong consistency is not required
across inboxes. Resolving eventually to any non-revoked inbox_id is
acceptable.

Algorithm for PublishIdentityUpdate:

Start transaction (SERIALIZABLE isolation level)

 1. Read the log for the inbox_id, ordering by sequence_id
    - Use FOR UPDATE to block other transactions on the same inbox_id
 2. If the log has 256 or more entries, abort the transaction.
 3. Concatenate the update in-memory and validate it sequentially. If
    failed, abort the transaction.
 4. Insert the update into the inbox_log table
 5. For each affected address:
    a. Insert or update the record with (address, inbox_id) into
    the address_log table, updating the relevant sequence_id (it should
    always be higher)

End transaction
*/
func (s *Service) PublishIdentityUpdate(ctx context.Context, req *api.PublishIdentityUpdateRequest) (*api.PublishIdentityUpdateResponse, error) {
	res, err := s.store.PublishIdentityUpdate(ctx, req, s.validationService)
	if err != nil {
		return nil, err
	}

	if err = s.PublishAssociationChangesEvent(ctx, res); err != nil {
		s.log.Error("error publishing association changes event", zap.Error(err))
		// Don't return the erro here because the transaction has already been committed
	}

	return &api.PublishIdentityUpdateResponse{}, nil
}

func (s *Service) GetIdentityUpdates(ctx context.Context, req *api.GetIdentityUpdatesRequest) (*api.GetIdentityUpdatesResponse, error) {
	/*
		Algorithm for each request:
		1. Query the inbox_log table for the inbox_id, ordering by sequence_id
		2. Return all of the entries
	*/
	return s.store.GetInboxLogs(ctx, req)
}

func (s *Service) GetInboxIds(ctx context.Context, req *api.GetInboxIdsRequest) (*api.GetInboxIdsResponse, error) {
	/*
		Algorithm for each request:
		1. Query the address_log table for the largest association_sequence_id
		   for the address where revocation_sequence_id is lower or NULL
		2. Return the value of the 'inbox_id' column
	*/
	return s.store.GetInboxIds(ctx, req)
}

func (s *Service) VerifySmartContractWalletSignatures(ctx context.Context, req *identity.VerifySmartContractWalletSignaturesRequest) (*identity.VerifySmartContractWalletSignaturesResponse, error) {
	return s.validationService.VerifySmartContractWalletSignatures(ctx, req)
}

func (s *Service) SubscribeAssociationChanges(req *identity.SubscribeAssociationChangesRequest, stream identity.IdentityApi_SubscribeAssociationChangesServer) error {
	log := s.log.Named("subscribe-association-changes")
	log.Info("subscription started")

	_ = stream.SendHeader(metadata.Pairs("subscribed", "true"))

	natsSubject := buildNatsSubjectForAssociationChanges()
	sub, err := s.nc.Subscribe(natsSubject, func(natsMsg *nats.Msg) {
		msg, err := getAssociationChangedMessageFromNats(natsMsg)
		if err != nil {
			log.Error("parsing message", zap.Error(err))
		}
		if err = stream.Send(msg); err != nil {
			log.Warn("sending message to stream", zap.Error(err))
		}
	})

	if err != nil {
		log.Error("error subscribing to nats", zap.Error(err))
		return err
	}

	defer func() {
		_ = sub.Unsubscribe()
	}()

	select {
	case <-stream.Context().Done():
		return nil
	case <-s.ctx.Done():
		return nil
	}
}

func (s *Service) PublishAssociationChangesEvent(ctx context.Context, identityUpdateResult *identityTypes.PublishIdentityUpdateResult) error {
	protoEvents := identityUpdateResult.GetChanges()
	if len(protoEvents) == 0 {
		return nil
	}

	for _, protoEvent := range protoEvents {
		msgBytes, err := pb.Marshal(protoEvent)
		if err != nil {
			return err
		}

		if err = s.publishToWakuRelay(ctx, &wakupb.WakuMessage{
			ContentTopic: topic.AssociationChangedTopic,
			Timestamp:    int64(identityUpdateResult.TimestampNs),
			Payload:      msgBytes,
		}); err != nil {
			return err
		}
	}

	return nil
}

func buildNatsSubjectForAssociationChanges() string {
	return envelopes.BuildNatsSubject(topic.BuildAssociationChangedTopic())
}

func getAssociationChangedMessageFromNats(natsMsg *nats.Msg) (*identity.SubscribeAssociationChangesResponse, error) {
	var env v1proto.Envelope
	err := pb.Unmarshal(natsMsg.Data, &env)
	if err != nil {
		return nil, err
	}

	var msg identity.SubscribeAssociationChangesResponse
	err = pb.Unmarshal(env.Message, &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}
