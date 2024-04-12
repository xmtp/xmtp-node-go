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

func (s *Service) PublishIdentityUpdate(ctx context.Context, req *api.PublishIdentityUpdateRequest) (*api.PublishIdentityUpdateResponse, error) {
	// Algorithm:
	// Note - the inbox_log table has global ordering via a serial sequence_ID, and an index by inbox_ID
	// 1. Append update to DB under inbox_id with commit_status UNCOMMITTED
	// 2. ProcessLog(inbox_id):
	//  	- Read the log for the inbox_id
	//  	- Validate log sequentially
	//		- Update UNCOMMITTED rows to either VALIDATED or delete them based on the validation result
	//  	- For each row that is VALIDATED:
	//			- Add it to the relevant address log.
	//				- Note: There may be races between multiple ProcessLog() calls on the same inbox, or across multiple inboxes.
	//			  	  The address log can use a unique index on inbox_log_sequence_ID to prevent duplicate updates and establish ordering.
	//          - Process the address log and cache the XID into a third table (address_lookup_cache)
	//				- Note: To prevent new data overwriting old data, the address_lookup_cache stores the inbox_log_sequence_id, and we do
	//				  an atomic update WHERE new_sequence_id > old_sequence_id
	// 			- Update the row in the inbox_id table to COMMITTED
	// 3. Return success from the API if the original identity update was COMMITTED, else return error
	//
	// If the server goes down in the middle of processing an update, subsequent ProcessLog() calls will pick up where the previous one left off.
	// The client is expected to retry with the same payload, and the server is expected to deduplicate the update.

	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetIdentityUpdates(ctx context.Context, req *api.GetIdentityUpdatesRequest) (*api.GetIdentityUpdatesResponse, error) {
	// Algorithm:
	// 1. Query the relevant inbox_log tables, filtering to COMMITTED rows
	// 2. Return the updates in the response
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *Service) GetInboxIds(ctx context.Context, req *api.GetInboxIdsRequest) (*api.GetInboxIdsResponse, error) {
	// Algorithm:
	// 1. Query the address_lookup_cache for each address
	// 2. Return the result
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}
