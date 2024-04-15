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
    a. Insert or update the record with (address, inbox_id) into
    the address_log table. Update the sequence_id if it is
    higher

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
		1. Query the address_log table for the largest association_sequence_id
		   for the address where revocation_sequence_id is lower or NULL
		2. Return the value of the 'inbox_id' column
	*/
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}
