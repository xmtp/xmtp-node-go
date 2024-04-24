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
	PublishIdentityUpdate(ctx context.Context, req *identity.PublishIdentityUpdateRequest) (*identity.PublishIdentityUpdateResponse, error)
	GetInboxLogs(ctx context.Context, req *identity.GetIdentityUpdatesRequest) (*identity.GetIdentityUpdatesResponse, error)
	GetInboxIds(ctx context.Context, req *identity.GetInboxIdsRequest) (*identity.GetInboxIdsResponse, error)
}

type MlsStore interface {
	IdentityStore

	CreateInstallation(ctx context.Context, installationId []byte, walletAddress string, credentialIdentity, keyPackage []byte, expiration uint64) error
	UpdateKeyPackage(ctx context.Context, installationId, keyPackage []byte, expiration uint64) error
	FetchKeyPackages(ctx context.Context, installationIds [][]byte) ([]queries.FetchKeyPackagesRow, error)
	GetIdentityUpdates(ctx context.Context, walletAddresses []string, startTimeNs int64) (map[string]IdentityUpdateList, error)
	InsertGroupMessage(ctx context.Context, groupId []byte, data []byte) (*GroupMessage, error)
	InsertWelcomeMessage(ctx context.Context, installationId []byte, data []byte, hpkePublicKey []byte) (*WelcomeMessage, error)
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

		for _, log_entry := range addressLogEntries {
			if log_entry.Address == address {
				resp.InboxId = &log_entry.InboxID
			}
		}
		out[index] = &resp
	}

	return &identity.GetInboxIdsResponse{
		Responses: out,
	}, nil
}

func (s *Store) PublishIdentityUpdate(ctx context.Context, req *identity.PublishIdentityUpdateRequest) (*identity.PublishIdentityUpdateResponse, error) {
	new_update := req.GetIdentityUpdate()
	if new_update == nil {
		return nil, errors.New("IdentityUpdate is required")
	}

	if err := s.RunInTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable}, func(ctx context.Context, txQueries *queries.Queries) error {
		inboxLogEntries, err := txQueries.GetAllInboxLogs(ctx, new_update.GetInboxId())
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
		_ = append(updates, new_update)

		// TODO: Validate the updates, and abort transaction if failed

		protoBytes, err := proto.Marshal(new_update)
		if err != nil {
			return err
		}

		_, err = txQueries.InsertInboxLog(ctx, queries.InsertInboxLogParams{
			InboxID:             new_update.GetInboxId(),
			ServerTimestampNs:   nowNs(),
			IdentityUpdateProto: protoBytes,
		})

		if err != nil {
			return err
		}
		// TODO: Insert or update the address_log table using sequence_id

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

type GroupedAddressLogEntry struct {
	Address                  string
	InboxId                  string
	MaxAssociationSequenceId uint64
}

func (s *Store) GetInboxIds(ctx context.Context, req *identity.GetInboxIdsRequest) (*identity.GetInboxIdsResponse, error) {

	addresses := []string{}
	for _, request := range req.Requests {
		addresses = append(addresses, request.GetAddress())
	}

	db_log_entries := make([]*GroupedAddressLogEntry, 0)

	// maybe restrict select to only fields in groupby?
	err := s.db.NewSelect().
		Model((*AddressLogEntry)(nil)).
		ColumnExpr("address, inbox_id, MAX(association_sequence_id) as max_association_sequence_id").
		Where("address IN (?)", bun.In(addresses)).
		Where("revocation_sequence_id IS NULL OR revocation_sequence_id < association_sequence_id").
		Group("address", "inbox_id").
		OrderExpr("MAX(association_sequence_id) ASC").
		Scan(ctx, &db_log_entries)

	if err != nil {
		return nil, err
	}

	out := make([]*identity.GetInboxIdsResponse_Response, len(addresses))

	for index, address := range addresses {
		resp := identity.GetInboxIdsResponse_Response{}
		resp.Address = address

		for _, log_entry := range db_log_entries {
			if log_entry.Address == address {
				resp.InboxId = &log_entry.InboxId
			}
		}
		out[index] = &resp
	}

	return &identity.GetInboxIdsResponse{
		Responses: out,
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

func (s *Store) InsertGroupMessage(ctx context.Context, groupId []byte, data []byte) (*GroupMessage, error) {
	message := GroupMessage{
		Data: data,
	}

	var id uint64
	err := s.db.QueryRow("INSERT INTO group_messages (group_id, data, group_id_data_hash) VALUES (?, ?, ?) RETURNING id", groupId, data, sha256.Sum256(append(groupId, data...))).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, NewAlreadyExistsError(err)
		}
		return nil, err
	}

	err = s.db.NewSelect().Model(&message).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (s *Store) InsertWelcomeMessage(ctx context.Context, installationId []byte, data []byte, hpkePublicKey []byte) (*WelcomeMessage, error) {
	message := WelcomeMessage{
		Data: data,
	}

	var id uint64
	err := s.db.QueryRow("INSERT INTO welcome_messages (installation_key, data, installation_key_data_hash, hpke_public_key) VALUES (?, ?, ?, ?) RETURNING id", installationId, data, sha256.Sum256(append(installationId, data...)), hpkePublicKey).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, NewAlreadyExistsError(err)
		}
		return nil, err
	}

	err = s.db.NewSelect().Model(&message).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (s *Store) QueryGroupMessagesV1(ctx context.Context, req *mlsv1.QueryGroupMessagesRequest) (*mlsv1.QueryGroupMessagesResponse, error) {
	msgs := make([]*GroupMessage, 0)

	if len(req.GroupId) == 0 {
		return nil, errors.New("group is required")
	}

	q := s.db.NewSelect().
		Model(&msgs).
		Where("group_id = ?", req.GroupId)

	direction := mlsv1.SortDirection_SORT_DIRECTION_DESCENDING
	if req.PagingInfo != nil && req.PagingInfo.Direction != mlsv1.SortDirection_SORT_DIRECTION_UNSPECIFIED {
		direction = req.PagingInfo.Direction
	}
	switch direction {
	case mlsv1.SortDirection_SORT_DIRECTION_DESCENDING:
		q = q.Order("id DESC")
	case mlsv1.SortDirection_SORT_DIRECTION_ASCENDING:
		q = q.Order("id ASC")
	}

	pageSize := maxPageSize
	if req.PagingInfo != nil && req.PagingInfo.Limit > 0 && req.PagingInfo.Limit <= maxPageSize {
		pageSize = int(req.PagingInfo.Limit)
	}
	q = q.Limit(pageSize)

	if req.PagingInfo != nil && req.PagingInfo.IdCursor != 0 {
		if direction == mlsv1.SortDirection_SORT_DIRECTION_ASCENDING {
			q = q.Where("id > ?", req.PagingInfo.IdCursor)
		} else {
			q = q.Where("id < ?", req.PagingInfo.IdCursor)
		}
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}

	messages := make([]*mlsv1.GroupMessage, 0, len(msgs))
	for _, msg := range msgs {
		messages = append(messages, &mlsv1.GroupMessage{
			Version: &mlsv1.GroupMessage_V1_{
				V1: &mlsv1.GroupMessage_V1{
					Id:        msg.Id,
					CreatedNs: uint64(msg.CreatedAt.UnixNano()),
					GroupId:   msg.GroupId,
					Data:      msg.Data,
				},
			},
		})
	}

	pagingInfo := &mlsv1.PagingInfo{Limit: uint32(pageSize), IdCursor: 0, Direction: direction}
	if len(messages) >= pageSize {
		lastMsg := msgs[len(messages)-1]
		pagingInfo.IdCursor = lastMsg.Id
	}

	return &mlsv1.QueryGroupMessagesResponse{
		Messages:   messages,
		PagingInfo: pagingInfo,
	}, nil
}

func (s *Store) QueryWelcomeMessagesV1(ctx context.Context, req *mlsv1.QueryWelcomeMessagesRequest) (*mlsv1.QueryWelcomeMessagesResponse, error) {
	msgs := make([]*WelcomeMessage, 0)

	if len(req.InstallationKey) == 0 {
		return nil, errors.New("installation is required")
	}

	q := s.db.NewSelect().
		Model(&msgs).
		Where("installation_key = ?", req.InstallationKey)

	direction := mlsv1.SortDirection_SORT_DIRECTION_DESCENDING
	if req.PagingInfo != nil && req.PagingInfo.Direction != mlsv1.SortDirection_SORT_DIRECTION_UNSPECIFIED {
		direction = req.PagingInfo.Direction
	}
	switch direction {
	case mlsv1.SortDirection_SORT_DIRECTION_DESCENDING:
		q = q.Order("id DESC")
	case mlsv1.SortDirection_SORT_DIRECTION_ASCENDING:
		q = q.Order("id ASC")
	}

	pageSize := maxPageSize
	if req.PagingInfo != nil && req.PagingInfo.Limit > 0 && req.PagingInfo.Limit <= maxPageSize {
		pageSize = int(req.PagingInfo.Limit)
	}
	q = q.Limit(pageSize)

	if req.PagingInfo != nil && req.PagingInfo.IdCursor != 0 {
		if direction == mlsv1.SortDirection_SORT_DIRECTION_ASCENDING {
			q = q.Where("id > ?", req.PagingInfo.IdCursor)
		} else {
			q = q.Where("id < ?", req.PagingInfo.IdCursor)
		}
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}

	messages := make([]*mlsv1.WelcomeMessage, 0, len(msgs))
	for _, msg := range msgs {
		messages = append(messages, &mlsv1.WelcomeMessage{
			Version: &mlsv1.WelcomeMessage_V1_{
				V1: &mlsv1.WelcomeMessage_V1{
					Id:              msg.Id,
					CreatedNs:       uint64(msg.CreatedAt.UnixNano()),
					Data:            msg.Data,
					InstallationKey: msg.InstallationKey,
					HpkePublicKey:   msg.HpkePublicKey,
				},
			},
		})
	}

	pagingInfo := &mlsv1.PagingInfo{Limit: uint32(pageSize), IdCursor: 0, Direction: direction}
	if len(messages) >= pageSize {
		lastMsg := msgs[len(messages)-1]
		pagingInfo.IdCursor = lastMsg.Id
	}

	return &mlsv1.QueryWelcomeMessagesResponse{
		Messages:   messages,
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
