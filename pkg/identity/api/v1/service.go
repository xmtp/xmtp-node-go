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
Algorithm:

Start transaction

 1. Insert the update into the inbox_log table
 2. Read the log for the inbox_id, ordering by sequence_id
 3. If the log has more than 256 entries, abort the transaction.
 3. Validate it sequentially. If failed, abort the transaction.
 4. For each affected address:
    a. Insert the update into the address_log table
    -- Note: There may be races across multiple inboxes. The address_log can use
    -- inbox_log_sequence_ID to establish ordering.
    b. Read the log for the address, ordering by *inbox_log_sequence_id*
    c. If the log has more than 256 entries, abort the transaction.
    c. Process the address log (without signature validation) and compute the final inbox ID
    d. Update the inserted entry with the result.

End transaction
*/
func (s *Service) PublishIdentityUpdate(ctx context.Context, req *api.PublishIdentityUpdateRequest) (*api.PublishIdentityUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetIdentityUpdates(ctx context.Context, req *api.GetIdentityUpdatesRequest) (*api.GetIdentityUpdatesResponse, error) {
	/*
		Algorithm for each request:
		1. Query the inbox_log table for the inbox_id, ordering by sequence_id
		2. Return all of the entries
	*/
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetInboxIds(ctx context.Context, req *api.GetInboxIdsRequest) (*api.GetInboxIdsResponse, error) {
	/*
		Algorithm for each request:
		1. Query the address_log table for the latest sequence_id on the address
			  -- Note: NOT inbox_log_sequence_id
		2. Return the value of the 'inbox_id' column
	*/
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}
