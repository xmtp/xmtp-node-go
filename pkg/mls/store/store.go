package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/utils"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	migrations "github.com/xmtp/xmtp-node-go/pkg/migrations/mls"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/proto/mls/message_contents"
	"github.com/xmtp/xmtp-node-go/pkg/types"
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
	PublishIdentityUpdate(
		ctx context.Context,
		req *identity.PublishIdentityUpdateRequest,
		validationService mlsvalidate.MLSValidationService,
	) (*identity.PublishIdentityUpdateResponse, error)
	GetInboxLogs(
		ctx context.Context,
		req *identity.GetIdentityUpdatesRequest,
	) (*identity.GetIdentityUpdatesResponse, error)
	GetInboxIds(
		ctx context.Context,
		req *identity.GetInboxIdsRequest,
	) (*identity.GetInboxIdsResponse, error)
}

type MlsStore interface {
	IdentityStore

	CreateOrUpdateInstallation(ctx context.Context, installationId []byte, keyPackage []byte) error
	FetchKeyPackages(
		ctx context.Context,
		installationIds [][]byte,
	) ([]queries.FetchKeyPackagesRow, error)
	InsertGroupMessage(
		ctx context.Context,
		groupId []byte,
		data []byte,
		senderHmac []byte,
		shouldPush bool,
		isCommit bool,
	) (*queries.GroupMessage, error)
	InsertWelcomeMessage(
		ctx context.Context,
		installationId []byte,
		data []byte,
		hpkePublicKey []byte,
		algorithm types.WrapperAlgorithm,
		welcomeMetadata []byte,
	) (*queries.WelcomeMessage, error)
	QueryGroupMessagesV1(
		ctx context.Context,
		query *mlsv1.QueryGroupMessagesRequest,
	) (*mlsv1.QueryGroupMessagesResponse, error)
	QueryWelcomeMessagesV1(
		ctx context.Context,
		query *mlsv1.QueryWelcomeMessagesRequest,
	) (*mlsv1.QueryWelcomeMessagesResponse, error)
	QueryCommitLog(
		ctx context.Context,
		query *mlsv1.QueryCommitLogRequest,
	) (*mlsv1.QueryCommitLogResponse, error)
	InsertCommitLog(
		ctx context.Context,
		groupId []byte,
		encrypted_entry []byte,
	) (queries.CommitLog, error)
}

var _ MlsStore = (*Store)(nil)

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

func (s *Store) GetInboxIds(
	ctx context.Context,
	req *identity.GetInboxIdsRequest,
) (*identity.GetInboxIdsResponse, error) {
	addresses := []string{}
	for _, request := range req.Requests {
		addresses = append(addresses, request.GetIdentifier())
	}

	addressLogEntries, err := s.queries.GetAddressLogs(ctx, addresses)
	if err != nil {
		return nil, err
	}

	out := make([]*identity.GetInboxIdsResponse_Response, len(addresses))

	for index, address := range addresses {
		resp := identity.GetInboxIdsResponse_Response{}
		resp.Identifier = address

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

func (s *Store) PublishIdentityUpdate(
	ctx context.Context,
	req *identity.PublishIdentityUpdateRequest,
	validationService mlsvalidate.MLSValidationService,
) (*identity.PublishIdentityUpdateResponse, error) {
	newUpdate := req.GetIdentityUpdate()
	if newUpdate == nil {
		return nil, errors.New("IdentityUpdate is required")
	}

	if err := s.RunInRepeatableReadTx(ctx, 3, func(ctx context.Context, txQueries *queries.Queries) error {
		inboxId := newUpdate.GetInboxId()
		// We use a pg_advisory_lock to lock the inbox_id instead of SELECT FOR UPDATE
		// This allows the lock to be enforced even when there are no existing `inbox_log`s
		if err := txQueries.LockInboxLog(ctx, inboxId); err != nil {
			return err
		}

		log := s.log.With(zap.String("inbox_id", inboxId))
		inboxLogEntries, err := txQueries.GetAllInboxLogs(ctx, inboxId)
		if err != nil {
			return err
		}

		if len(inboxLogEntries) >= 256 {
			return errors.New("inbox log is full")
		}

		updates := make([]*associations.IdentityUpdate, 0, len(inboxLogEntries))
		for _, log := range inboxLogEntries {
			identityUpdate := &associations.IdentityUpdate{}
			if err := proto.Unmarshal(log.IdentityUpdateProto, identityUpdate); err != nil {
				return err
			}
			updates = append(updates, identityUpdate)
		}

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
			if address, ok := new_member.Kind.(*associations.MemberIdentifier_EthereumAddress); ok {
				_, err = txQueries.InsertAddressLog(ctx, queries.InsertAddressLogParams{
					Address:               address.EthereumAddress,
					InboxID:               inboxId,
					AssociationSequenceID: sql.NullInt64{Valid: true, Int64: sequence_id},
					RevocationSequenceID:  sql.NullInt64{Valid: false},
				})
				if err != nil {
					return err
				}
			}
		}

		for _, removed_member := range state.StateDiff.RemovedMembers {
			log.Info("Removed member", zap.Any("member", removed_member))
			if address, ok := removed_member.Kind.(*associations.MemberIdentifier_EthereumAddress); ok {
				err = txQueries.RevokeAddressFromLog(ctx, queries.RevokeAddressFromLogParams{
					Address:              address.EthereumAddress,
					InboxID:              inboxId,
					RevocationSequenceID: sql.NullInt64{Valid: true, Int64: sequence_id},
				})
				if err != nil {
					return err
				}
			}
		}

		err = txQueries.TouchInbox(ctx, inboxId)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &identity.PublishIdentityUpdateResponse{}, nil
}

func (s *Store) GetInboxLogs(
	ctx context.Context,
	batched_req *identity.GetIdentityUpdatesRequest,
) (*identity.GetIdentityUpdatesResponse, error) {
	reqs := batched_req.GetRequests()

	filters := make(queries.InboxLogFilterList, len(reqs))
	for i, req := range reqs {
		filters[i] = queries.InboxLogFilter{
			InboxId:    req.InboxId, // InboxLogFilters take inbox_id as text and decode it inside Postgres, since the filters are JSON
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
	resultMap := make(map[string][]queries.GetInboxLogFilteredRow)
	for _, result := range results {
		inboxId := result.InboxID
		resultMap[inboxId] = append(resultMap[inboxId], result)
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
func (s *Store) CreateOrUpdateInstallation(
	ctx context.Context,
	installationId []byte,
	keyPackage []byte,
) error {
	now := nowNs()

	return s.queries.CreateOrUpdateInstallation(ctx, queries.CreateOrUpdateInstallationParams{
		ID:         installationId,
		CreatedAt:  now,
		UpdatedAt:  now,
		KeyPackage: keyPackage,
	})
}

func (s *Store) FetchKeyPackages(
	ctx context.Context,
	installationIds [][]byte,
) ([]queries.FetchKeyPackagesRow, error) {
	return s.queries.FetchKeyPackages(ctx, installationIds)
}

func (s *Store) InsertGroupMessage(
	ctx context.Context,
	groupId []byte,
	data []byte,
	senderHmac []byte,
	shouldPush bool,
	isCommit bool,
) (*queries.GroupMessage, error) {
	dataHash := sha256.Sum256(append(groupId, data...))
	message, err := s.queries.InsertGroupMessage(ctx, queries.InsertGroupMessageParams{
		GroupID:         groupId,
		Data:            data,
		GroupIDDataHash: dataHash[:],
		IsCommit:        isCommit,
		SenderHmac:      senderHmac,
		ShouldPush:      shouldPush,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, NewAlreadyExistsError(err)
		}
		return nil, err
	}

	return &message, nil
}

func (s *Store) InsertWelcomeMessage(
	ctx context.Context,
	installationId []byte,
	data []byte,
	hpkePublicKey []byte,
	wrapperAlgorithm types.WrapperAlgorithm,
	welcomeMetadata []byte,
) (*queries.WelcomeMessage, error) {
	dataHash := sha256.Sum256(append(installationId, data...))
	message, err := s.queries.InsertWelcomeMessage(ctx, queries.InsertWelcomeMessageParams{
		InstallationKey:         installationId,
		Data:                    data,
		InstallationKeyDataHash: dataHash[:],
		HpkePublicKey:           hpkePublicKey,
		WrapperAlgorithm:        int16(wrapperAlgorithm),
		WelcomeMetadata:         welcomeMetadata,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, NewAlreadyExistsError(err)
		}
		return nil, err
	}

	return &message, nil
}

func (s *Store) QueryGroupMessagesV1(
	ctx context.Context,
	req *mlsv1.QueryGroupMessagesRequest,
) (*mlsv1.QueryGroupMessagesResponse, error) {
	if len(req.GroupId) == 0 {
		return nil, errors.New("group is required")
	}

	sortDesc := true
	var idCursor int64
	var err error
	var messages []queries.GroupMessage
	pageSize := int32(maxPageSize)

	if req.PagingInfo != nil &&
		req.PagingInfo.Direction == mlsv1.SortDirection_SORT_DIRECTION_ASCENDING {
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
			messages, err = s.queries.QueryGroupMessagesWithCursorDesc(
				ctx,
				queries.QueryGroupMessagesWithCursorDescParams{
					GroupID: req.GroupId,
					Cursor:  idCursor,
					Numrows: pageSize,
				},
			)
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
					Id:         uint64(msg.ID),
					CreatedNs:  uint64(msg.CreatedAt.UnixNano()),
					GroupId:    msg.GroupID,
					Data:       msg.Data,
					ShouldPush: msg.ShouldPush.Bool,
					SenderHmac: msg.SenderHmac,
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

func (s *Store) QueryWelcomeMessagesV1(
	ctx context.Context,
	req *mlsv1.QueryWelcomeMessagesRequest,
) (*mlsv1.QueryWelcomeMessagesResponse, error) {
	if len(req.InstallationKey) == 0 {
		return nil, errors.New("installation is required")
	}

	sortDesc := true
	direction := mlsv1.SortDirection_SORT_DIRECTION_DESCENDING
	pageSize := int32(maxPageSize)
	var idCursor int64
	var err error
	var messages []queries.WelcomeMessage

	if req.PagingInfo != nil &&
		req.PagingInfo.Direction == mlsv1.SortDirection_SORT_DIRECTION_ASCENDING {
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
			messages, err = s.queries.QueryWelcomeMessagesWithCursorDesc(
				ctx,
				queries.QueryWelcomeMessagesWithCursorDescParams{
					InstallationKey: req.InstallationKey,
					Cursor:          idCursor,
					Numrows:         pageSize,
				},
			)
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
					WrapperAlgorithm: types.WrapperAlgorithmToProto(
						types.WrapperAlgorithm(msg.WrapperAlgorithm),
					),
					WelcomeMetadata: msg.WelcomeMetadata,
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

func (s *Store) QueryCommitLog(
	ctx context.Context,
	req *mlsv1.QueryCommitLogRequest,
) (*mlsv1.QueryCommitLogResponse, error) {
	if len(req.GetGroupId()) == 0 {
		return nil, errors.New("group is required")
	}
	if req.GetPagingInfo().GetDirection() == mlsv1.SortDirection_SORT_DIRECTION_DESCENDING {
		return nil, errors.New("descending direction is not supported")
	}
	pageSize := int32(req.GetPagingInfo().GetLimit())
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	idCursor := int64(req.GetPagingInfo().GetIdCursor())

	entries, err := s.queries.QueryCommitLog(ctx, queries.QueryCommitLogParams{
		GroupID: req.GroupId,
		Cursor:  idCursor,
		Numrows: pageSize,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*message_contents.CommitLogEntry, len(entries))
	for idx, entry := range entries {
		out[idx] = &message_contents.CommitLogEntry{
			SequenceId:              uint64(entry.ID),
			EncryptedCommitLogEntry: entry.EncryptedEntry,
		}
	}

	pagingInfo := &mlsv1.PagingInfo{
		Limit:     uint32(pageSize),
		IdCursor:  0,
		Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
	}
	// It is strange behavior, but we must follow the semantics of other queries in v3 - only setting the IdCursor
	// to a non-zero value if we have a full page of results
	if len(entries) >= int(pageSize) {
		lastEntry := entries[len(entries)-1]
		pagingInfo.IdCursor = uint64(lastEntry.ID)
	}

	return &mlsv1.QueryCommitLogResponse{
		GroupId:          req.GroupId,
		CommitLogEntries: out,
		PagingInfo:       pagingInfo,
	}, nil
}

func (s *Store) InsertCommitLog(
	ctx context.Context,
	groupId []byte,
	encrypted_entry []byte,
) (queries.CommitLog, error) {
	return s.queries.InsertCommitLog(ctx, queries.InsertCommitLogParams{
		GroupID:        groupId,
		EncryptedEntry: encrypted_entry,
	})
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

func (s *Store) RunInRepeatableReadTx(
	ctx context.Context,
	numRetries int,
	fn func(ctx context.Context, txQueries *queries.Queries) error,
) error {
	var err error
	for i := 0; i < numRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = RunInTx(ctx, s.db.DB, &sql.TxOptions{Isolation: sql.LevelRepeatableRead}, fn)
			if err == nil {
				return nil
			}
			s.log.Warn("Error in tx", zap.Error(err))
			utils.RandomSleep(20)
		}
	}
	return err
}
