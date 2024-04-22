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
	return s.store.PublishIdentityUpdate(ctx, req)
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
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}
