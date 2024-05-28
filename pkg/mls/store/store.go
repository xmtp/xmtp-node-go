package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	migrations "github.com/xmtp/xmtp-node-go/pkg/migrations/mls"
	queries "github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const maxPageSize = 100

type Store struct {
	config  Config
	log     *zap.Logger
	db      *bun.DB
	queries *queries.Queries
}

type IdentityStore interface {
	PublishIdentityUpdate(ctx context.Context, req *identity.PublishIdentityUpdateRequest, validationService mlsvalidate.MLSValidationService) (*identity.PublishIdentityUpdateResponse, error)
	GetInboxLogs(ctx context.Context, req *identity.GetIdentityUpdatesRequest) (*identity.GetIdentityUpdatesResponse, error)
	GetInboxIds(ctx context.Context, req *identity.GetInboxIdsRequest) (*identity.GetInboxIdsResponse, error)
}

type MlsStore interface {
	IdentityStore

	CreateInstallation(ctx context.Context, installationId []byte, walletAddress string, credentialIdentity, keyPackage []byte, expiration uint64) error
	UpdateKeyPackage(ctx context.Context, installationId, keyPackage []byte, expiration uint64) error
	FetchKeyPackages(ctx context.Context, installationIds [][]byte) ([]queries.FetchKeyPackagesRow, error)
	GetIdentityUpdates(ctx context.Context, walletAddresses []string, startTimeNs int64) (map[string]IdentityUpdateList, error)
	InsertGroupMessage(ctx context.Context, groupId []byte, data []byte) (*queries.GroupMessage, error)
	InsertWelcomeMessage(ctx context.Context, installationId []byte, data []byte, hpkePublicKey []byte) (*queries.WelcomeMessage, error)
	QueryGroupMessagesV1(ctx context.Context, query *mlsv1.QueryGroupMessagesRequest) (*mlsv1.QueryGroupMessagesResponse, error)
	QueryWelcomeMessagesV1(ctx context.Context, query *mlsv1.QueryWelcomeMessagesRequest) (*mlsv1.QueryWelcomeMessagesResponse, error)
}

func New(ctx context.Context, config Config) (*Store, error) {
	if config.now == nil {
		config.now = func() time.Time {
			return time.Now().UTC()
		}
	}
	s := &Store{
		log:     config.Log.Named("mlsstore"),
		db:      config.DB,
		config:  config,
		queries: queries.New(config.DB.DB),
	}

	if err := s.migrate(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) GetInboxIds(ctx context.Context, req *identity.GetInboxIdsRequest) (*identity.GetInboxIdsResponse, error) {

	addresses := []string{}
	for _, request := range req.Requests {
		addresses = append(addresses, request.GetAddress())
	}

	addressLogEntries, err := s.queries.GetAddressLogs(ctx, addresses)
	if err != nil {
		return nil, err
	}

	out := make([]*identity.GetInboxIdsResponse_Response, len(addresses))

	for index, address := range addresses {
		resp := identity.GetInboxIdsResponse_Response{}
		resp.Address = address

		for _, logEntry := range addressLogEntries {
			if logEntry.Address == address {
				inboxId := logEntry.InboxID
				resp.InboxId = &inboxId
			}
		}
		out[index] = &resp
	}

	return &identity.GetInboxIdsResponse{
		Responses: out,
	}, nil
}

func (s *Store) PublishIdentityUpdate(ctx context.Context, req *identity.PublishIdentityUpdateRequest, validationService mlsvalidate.MLSValidationService) (*identity.PublishIdentityUpdateResponse, error) {
	newUpdate := req.GetIdentityUpdate()
	if newUpdate == nil {
		return nil, errors.New("IdentityUpdate is required")
	}

	if err := s.RunInRepeatableReadTx(ctx, 3, func(ctx context.Context, txQueries *queries.Queries) error {
		inboxId := newUpdate.GetInboxId()
		log := s.log.With(zap.String("inbox_id", inboxId))
		inboxLogEntries, err := txQueries.GetAllInboxLogs(ctx, inboxId)
		if err != nil {
			return err
		}

		if len(inboxLogEntries) >= 256 {
			return errors.New("inbox log is full")
		}

		updates := make([]*associations.IdentityUpdate, 0, len(inboxLogEntries)+1)
		for _, log := range inboxLogEntries {
			identityUpdate := &associations.IdentityUpdate{}
			if err := proto.Unmarshal(log.IdentityUpdateProto, identityUpdate); err != nil {
				return err
			}
			updates = append(updates, identityUpdate)
		}
		_ = append(updates, newUpdate)

		state, err := validationService.GetAssociationState(ctx, updates, []*associations.IdentityUpdate{newUpdate})
		if err != nil {
			return err
		}

		protoBytes, err := proto.Marshal(newUpdate)
		if err != nil {
			return err
		}

		sequence_id, err := txQueries.InsertInboxLog(ctx, queries.InsertInboxLogParams{
			InboxID:             inboxId,
			ServerTimestampNs:   nowNs(),
			IdentityUpdateProto: protoBytes,
		})

		log.Info("Inserted inbox log", zap.Any("sequence_id", sequence_id))

		if err != nil {
			return err
		}

		for _, new_member := range state.StateDiff.NewMembers {
			log.Info("New member", zap.Any("member", new_member))
			if address, ok := new_member.Kind.(*associations.MemberIdentifier_Address); ok {
				_, err = txQueries.InsertAddressLog(ctx, queries.InsertAddressLogParams{
					Address:               address.Address,
					InboxID:               state.AssociationState.InboxId,
					AssociationSequenceID: sequence_id,
					RevocationSequenceID:  sql.NullInt64{Valid: false},
				})
				if err != nil {
					return err
				}
			}
		}

		for _, removed_member := range state.StateDiff.RemovedMembers {
			log.Info("Removed member", zap.Any("member", removed_member))
			if address, ok := removed_member.Kind.(*associations.MemberIdentifier_Address); ok {
				err = txQueries.RevokeAddressFromLog(ctx, queries.RevokeAddressFromLogParams{
					Address:              address.Address,
					InboxID:              state.AssociationState.InboxId,
					RevocationSequenceID: sql.NullInt64{Valid: true, Int64: sequence_id},
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &identity.PublishIdentityUpdateResponse{}, nil
}

func (s *Store) GetInboxLogs(ctx context.Context, batched_req *identity.GetIdentityUpdatesRequest) (*identity.GetIdentityUpdatesResponse, error) {
	reqs := batched_req.GetRequests()

	filters := make(queries.InboxLogFilterList, len(reqs))
	for i, req := range reqs {
		filters[i] = queries.InboxLogFilter{
			InboxId:    req.InboxId,
			SequenceId: int64(req.SequenceId),
		}
	}
	filterBytes, err := filters.ToSql()
	if err != nil {
		return nil, err
	}

	results, err := s.queries.GetInboxLogFiltered(ctx, filterBytes)
	if err != nil {
		return nil, err
	}

	// Organize the results by inbox ID
	resultMap := make(map[string][]queries.InboxLog)
	for _, result := range results {
		resultMap[result.InboxID] = append(resultMap[result.InboxID], result)
	}

	resps := make([]*identity.GetIdentityUpdatesResponse_Response, len(reqs))
	for i, req := range reqs {
		logEntries := resultMap[req.InboxId]
		updates := make([]*identity.GetIdentityUpdatesResponse_IdentityUpdateLog, len(logEntries))
		for j, entry := range logEntries {
			identity_update := &associations.IdentityUpdate{}
			if err := proto.Unmarshal(entry.IdentityUpdateProto, identity_update); err != nil {
				return nil, err
			}
			updates[j] = &identity.GetIdentityUpdatesResponse_IdentityUpdateLog{
				SequenceId:        uint64(entry.SequenceID),
				ServerTimestampNs: uint64(entry.ServerTimestampNs),
				Update:            identity_update,
			}
		}
		resps[i] = &identity.GetIdentityUpdatesResponse_Response{
			InboxId: req.InboxId,
			Updates: updates,
		}
	}

	return &identity.GetIdentityUpdatesResponse{
		Responses: resps,
	}, nil
}

// Creates the installation and last resort key package
func (s *Store) CreateInstallation(ctx context.Context, installationId []byte, walletAddress string, credentialIdentity, keyPackage []byte, expiration uint64) error {
	createdAt := nowNs()

	return s.queries.CreateInstallation(ctx, queries.CreateInstallationParams{
		ID:                 installationId,
		WalletAddress:      walletAddress,
		CreatedAt:          createdAt,
		CredentialIdentity: credentialIdentity,
		KeyPackage:         keyPackage,
		Expiration:         int64(expiration),
	})
}

// Insert a new key package, ignoring any that may already exist
func (s *Store) UpdateKeyPackage(ctx context.Context, installationId, keyPackage []byte, expiration uint64) error {
	rowsUpdated, err := s.queries.UpdateKeyPackage(ctx, queries.UpdateKeyPackageParams{
		ID:         installationId,
		UpdatedAt:  nowNs(),
		KeyPackage: keyPackage,
		Expiration: int64(expiration),
	})

	if err != nil {
		return err
	}

	if rowsUpdated == 0 {
		return errors.New("installation id unknown")
	}

	return nil
}

func (s *Store) FetchKeyPackages(ctx context.Context, installationIds [][]byte) ([]queries.FetchKeyPackagesRow, error) {
	return s.queries.FetchKeyPackages(ctx, installationIds)
}

func (s *Store) GetIdentityUpdates(ctx context.Context, walletAddresses []string, startTimeNs int64) (map[string]IdentityUpdateList, error) {
	updated, err := s.queries.GetIdentityUpdates(ctx, queries.GetIdentityUpdatesParams{
		WalletAddresses: walletAddresses,
		StartTime:       startTimeNs,
	})
	if err != nil {
		return nil, err
	}

	// The returned list is only partially sorted
	out := make(map[string]IdentityUpdateList)
	for _, installation := range updated {
		if installation.CreatedAt > startTimeNs {
			out[installation.WalletAddress] = append(out[installation.WalletAddress], IdentityUpdate{
				Kind:               Create,
				InstallationKey:    installation.ID,
				CredentialIdentity: installation.CredentialIdentity,
				TimestampNs:        uint64(installation.CreatedAt),
			})
		}
		if installation.RevokedAt.Valid && installation.RevokedAt.Int64 > startTimeNs {
			out[installation.WalletAddress] = append(out[installation.WalletAddress], IdentityUpdate{
				Kind:            Revoke,
				InstallationKey: installation.ID,
				TimestampNs:     uint64(installation.RevokedAt.Int64),
			})
		}
	}
	// Sort the updates by timestamp now that the full list is assembled
	for _, updates := range out {
		sort.Sort(updates)
	}

	return out, nil
}

func (s *Store) RevokeInstallation(ctx context.Context, installationId []byte) error {
	return s.queries.RevokeInstallation(ctx, queries.RevokeInstallationParams{
		RevokedAt:      sql.NullInt64{Valid: true, Int64: nowNs()},
		InstallationID: installationId,
	})
}

func (s *Store) InsertGroupMessage(ctx context.Context, groupId []byte, data []byte) (*queries.GroupMessage, error) {
	dataHash := sha256.Sum256(append(groupId, data...))
	message, err := s.queries.InsertGroupMessage(ctx, queries.InsertGroupMessageParams{
		GroupID:         groupId,
		Data:            data,
		GroupIDDataHash: dataHash[:],
	})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, NewAlreadyExistsError(err)
		}
		return nil, err
	}

	return &message, nil
}

func (s *Store) InsertWelcomeMessage(ctx context.Context, installationId []byte, data []byte, hpkePublicKey []byte) (*queries.WelcomeMessage, error) {
	dataHash := sha256.Sum256(append(installationId, data...))
	message, err := s.queries.InsertWelcomeMessage(ctx, queries.InsertWelcomeMessageParams{
		InstallationKey:         installationId,
		Data:                    data,
		InstallationKeyDataHash: dataHash[:],
		HpkePublicKey:           hpkePublicKey,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, NewAlreadyExistsError(err)
		}
		return nil, err
	}

	return &message, nil
}

func (s *Store) QueryGroupMessagesV1(ctx context.Context, req *mlsv1.QueryGroupMessagesRequest) (*mlsv1.QueryGroupMessagesResponse, error) {
	if len(req.GroupId) == 0 {
		return nil, errors.New("group is required")
	}

	sortDesc := true
	var idCursor int64
	var err error
	var messages []queries.GroupMessage
	pageSize := int32(maxPageSize)

	if req.PagingInfo != nil && req.PagingInfo.Direction == mlsv1.SortDirection_SORT_DIRECTION_ASCENDING {
		sortDesc = false
	}

	if req.PagingInfo != nil && req.PagingInfo.Limit > 0 && req.PagingInfo.Limit <= maxPageSize {
		pageSize = int32(req.PagingInfo.Limit)
	}

	if req.PagingInfo != nil && req.PagingInfo.IdCursor != 0 {
		idCursor = int64(req.PagingInfo.IdCursor)
	}

	if idCursor > 0 {
		if sortDesc {
			messages, err = s.queries.QueryGroupMessagesWithCursorDesc(ctx, queries.QueryGroupMessagesWithCursorDescParams{
				GroupID: req.GroupId,
				Cursor:  idCursor,
				Numrows: pageSize,
			})
		} else {
			messages, err = s.queries.QueryGroupMessagesWithCursorAsc(ctx, queries.QueryGroupMessagesWithCursorAscParams{
				GroupID: req.GroupId,
				Cursor:  idCursor,
				Numrows: pageSize,
			})
		}
	} else {
		messages, err = s.queries.QueryGroupMessages(ctx, queries.QueryGroupMessagesParams{
			GroupID:  req.GroupId,
			Numrows:  pageSize,
			SortDesc: sortDesc,
		})
	}

	if err != nil {
		return nil, err
	}

	out := make([]*mlsv1.GroupMessage, len(messages))
	for idx, msg := range messages {
		out[idx] = &mlsv1.GroupMessage{
			Version: &mlsv1.GroupMessage_V1_{
				V1: &mlsv1.GroupMessage_V1{
					Id:        uint64(msg.ID),
					CreatedNs: uint64(msg.CreatedAt.UnixNano()),
					GroupId:   msg.GroupID,
					Data:      msg.Data,
				},
			},
		}
	}

	direction := mlsv1.SortDirection_SORT_DIRECTION_ASCENDING
	if sortDesc {
		direction = mlsv1.SortDirection_SORT_DIRECTION_DESCENDING
	}

	pagingInfo := &mlsv1.PagingInfo{Limit: uint32(pageSize), IdCursor: 0, Direction: direction}
	if len(messages) >= int(pageSize) {
		lastMsg := messages[len(messages)-1]
		pagingInfo.IdCursor = uint64(lastMsg.ID)
	}

	return &mlsv1.QueryGroupMessagesResponse{
		Messages:   out,
		PagingInfo: pagingInfo,
	}, nil
}

func (s *Store) QueryWelcomeMessagesV1(ctx context.Context, req *mlsv1.QueryWelcomeMessagesRequest) (*mlsv1.QueryWelcomeMessagesResponse, error) {
	if len(req.InstallationKey) == 0 {
		return nil, errors.New("installation is required")
	}

	sortDesc := true
	direction := mlsv1.SortDirection_SORT_DIRECTION_DESCENDING
	pageSize := int32(maxPageSize)
	var idCursor int64
	var err error
	var messages []queries.WelcomeMessage

	if req.PagingInfo != nil && req.PagingInfo.Direction == mlsv1.SortDirection_SORT_DIRECTION_ASCENDING {
		sortDesc = false
		direction = mlsv1.SortDirection_SORT_DIRECTION_ASCENDING
	}

	if req.PagingInfo != nil && req.PagingInfo.Limit > 0 && req.PagingInfo.Limit <= maxPageSize {
		pageSize = int32(req.PagingInfo.Limit)
	}

	if req.PagingInfo != nil && req.PagingInfo.IdCursor != 0 {
		idCursor = int64(req.PagingInfo.IdCursor)
	}

	if idCursor > 0 {
		if sortDesc {
			messages, err = s.queries.QueryWelcomeMessagesWithCursorDesc(ctx, queries.QueryWelcomeMessagesWithCursorDescParams{
				InstallationKey: req.InstallationKey,
				Cursor:          idCursor,
				Numrows:         pageSize,
			})
		} else {
			messages, err = s.queries.QueryWelcomeMessagesWithCursorAsc(ctx, queries.QueryWelcomeMessagesWithCursorAscParams{
				InstallationKey: req.InstallationKey,
				Cursor:          idCursor,
				Numrows:         pageSize,
			})
		}
	} else {
		messages, err = s.queries.QueryWelcomeMessages(ctx, queries.QueryWelcomeMessagesParams{
			InstallationKey: req.InstallationKey,
			Numrows:         pageSize,
			SortDesc:        sortDesc,
		})
	}

	if err != nil {
		return nil, err
	}

	out := make([]*mlsv1.WelcomeMessage, len(messages))
	for idx, msg := range messages {
		out[idx] = &mlsv1.WelcomeMessage{
			Version: &mlsv1.WelcomeMessage_V1_{
				V1: &mlsv1.WelcomeMessage_V1{
					Id:              uint64(msg.ID),
					CreatedNs:       uint64(msg.CreatedAt.UnixNano()),
					Data:            msg.Data,
					InstallationKey: msg.InstallationKey,
					HpkePublicKey:   msg.HpkePublicKey,
				},
			},
		}
	}

	pagingInfo := &mlsv1.PagingInfo{Limit: uint32(pageSize), IdCursor: 0, Direction: direction}
	if len(messages) >= int(pageSize) {
		lastMsg := messages[len(messages)-1]
		pagingInfo.IdCursor = uint64(lastMsg.ID)
	}

	return &mlsv1.QueryWelcomeMessagesResponse{
		Messages:   out,
		PagingInfo: pagingInfo,
	}, nil
}

func (s *Store) migrate(ctx context.Context) error {
	migrator := migrate.NewMigrator(s.db, migrations.Migrations)
	err := migrator.Init(ctx)
	if err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}

	if group.IsZero() {
		s.log.Info("No new migrations to run")
	}

	return nil
}

func nowNs() int64 {
	return time.Now().UTC().UnixNano()
}

type IdentityUpdateKind int

const (
	Create IdentityUpdateKind = iota
	Revoke
)

type IdentityUpdate struct {
	Kind               IdentityUpdateKind
	InstallationKey    []byte
	CredentialIdentity []byte
	TimestampNs        uint64
}

// Add the required methods to make a valid sort.Sort interface
type IdentityUpdateList []IdentityUpdate

func (a IdentityUpdateList) Len() int {
	return len(a)
}

func (a IdentityUpdateList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a IdentityUpdateList) Less(i, j int) bool {
	return a[i].TimestampNs < a[j].TimestampNs
}

type AlreadyExistsError struct {
	Err error
}

func (e *AlreadyExistsError) Error() string {
	return e.Err.Error()
}

func NewAlreadyExistsError(err error) *AlreadyExistsError {
	return &AlreadyExistsError{err}
}

func IsAlreadyExistsError(err error) bool {
	_, ok := err.(*AlreadyExistsError)
	return ok
}

func (s *Store) RunInTx(
	ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, txQueries *queries.Queries) error,
) error {
	tx, err := s.db.DB.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	var done bool

	defer func() {
		if !done {
			_ = tx.Rollback()
		}
	}()

	if err := fn(ctx, s.queries.WithTx(tx)); err != nil {
		return err
	}

	done = true
	return tx.Commit()
}

func (s *Store) RunInRepeatableReadTx(ctx context.Context, numRetries int, fn func(ctx context.Context, txQueries *queries.Queries) error) error {
	var err error
	for i := 0; i < numRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = s.RunInTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead}, fn)
			if err == nil {
				return nil
			}
			s.log.Warn("Error in serializable tx", zap.Error(err))
		}
	}
	return err
}
