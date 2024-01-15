package store

import (
	"context"
	"crypto/sha256"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	migrations "github.com/xmtp/xmtp-node-go/pkg/migrations/mls"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"go.uber.org/zap"
)

const MaxQueryPageSize = 100

type Store struct {
	config Config
	log    *zap.Logger
	db     *bun.DB
}

type MlsStore interface {
	CreateInstallation(ctx context.Context, installationId []byte, walletAddress string, credentialIdentity, keyPackage []byte, expiration uint64) error
	UpdateKeyPackage(ctx context.Context, installationId, keyPackage []byte, expiration uint64) error
	FetchKeyPackages(ctx context.Context, installationIds [][]byte) ([]*Installation, error)
	GetIdentityUpdates(ctx context.Context, walletAddresses []string, startTimeNs int64) (map[string]IdentityUpdateList, error)
	InsertGroupMessage(ctx context.Context, groupId []byte, data []byte) (*GroupMessage, error)
	InsertWelcomeMessage(ctx context.Context, installationId []byte, data []byte) (*WelcomeMessage, error)
	GetLatestGroupMessage(ctx context.Context, groupId []byte) (*GroupMessage, error)
	GetLatestWelcomeMessage(ctx context.Context, installationKey []byte) (*WelcomeMessage, error)
	QueryGroupMessagesV1(ctx context.Context, query *mlsv1.QueryGroupMessagesRequest) ([]*GroupMessage, error)
	QueryWelcomeMessagesV1(ctx context.Context, query *mlsv1.QueryWelcomeMessagesRequest) ([]*WelcomeMessage, error)
}

func New(ctx context.Context, config Config) (*Store, error) {
	if config.now == nil {
		config.now = func() time.Time {
			return time.Now().UTC()
		}
	}
	s := &Store{
		log:    config.Log.Named("mlsstore"),
		db:     config.DB,
		config: config,
	}

	if err := s.migrate(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

// Creates the installation and last resort key package
func (s *Store) CreateInstallation(ctx context.Context, installationId []byte, walletAddress string, credentialIdentity, keyPackage []byte, expiration uint64) error {
	createdAt := nowNs()

	installation := Installation{
		ID:                 installationId,
		WalletAddress:      walletAddress,
		CreatedAt:          createdAt,
		UpdatedAt:          createdAt,
		CredentialIdentity: credentialIdentity,

		KeyPackage: keyPackage,
		Expiration: expiration,
	}

	_, err := s.db.NewInsert().
		Model(&installation).
		Ignore().
		Exec(ctx)
	return err
}

// Insert a new key package, ignoring any that may already exist
func (s *Store) UpdateKeyPackage(ctx context.Context, installationId, keyPackage []byte, expiration uint64) error {
	installation := Installation{
		ID:        installationId,
		UpdatedAt: nowNs(),

		KeyPackage: keyPackage,
		Expiration: expiration,
	}

	res, err := s.db.NewUpdate().
		Model(&installation).
		OmitZero().
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("installation id unknown")
	}
	return nil
}

func (s *Store) FetchKeyPackages(ctx context.Context, installationIds [][]byte) ([]*Installation, error) {
	installations := make([]*Installation, 0)

	err := s.db.NewSelect().
		Model(&installations).
		Where("ID IN (?)", bun.In(installationIds)).
		Scan(ctx, &installations)
	if err != nil {
		return nil, err
	}

	return installations, nil
}

func (s *Store) GetIdentityUpdates(ctx context.Context, walletAddresses []string, startTimeNs int64) (map[string]IdentityUpdateList, error) {
	updated := make([]*Installation, 0)
	// Find all installations that were changed since the startTimeNs
	err := s.db.NewSelect().
		Model(&updated).
		Where("wallet_address IN (?)", bun.In(walletAddresses)).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("created_at > ?", startTimeNs).WhereOr("revoked_at > ?", startTimeNs)
		}).
		Order("created_at ASC").
		Scan(ctx)

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
		if installation.RevokedAt != nil && *installation.RevokedAt > startTimeNs {
			out[installation.WalletAddress] = append(out[installation.WalletAddress], IdentityUpdate{
				Kind:            Revoke,
				InstallationKey: installation.ID,
				TimestampNs:     uint64(*installation.RevokedAt),
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
	_, err := s.db.NewUpdate().
		Model(&Installation{}).
		Set("revoked_at = ?", nowNs()).
		Where("id = ?", installationId).
		Where("revoked_at IS NULL").
		Exec(ctx)

	return err
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

func (s *Store) InsertWelcomeMessage(ctx context.Context, installationId []byte, data []byte) (*WelcomeMessage, error) {
	message := WelcomeMessage{
		Data: data,
	}

	var id uint64
	err := s.db.QueryRow("INSERT INTO welcome_messages (installation_key, data, installation_key_data_hash) VALUES (?, ?, ?) RETURNING id", installationId, data, sha256.Sum256(append(installationId, data...))).Scan(&id)
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

func (s *Store) GetLatestGroupMessage(ctx context.Context, groupId []byte) (*GroupMessage, error) {
	var msg GroupMessage
	err := s.db.NewSelect().Model(&msg).Where("group_id = ?", groupId).Order("id DESC").Limit(1).Scan(ctx)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, NewNotFoundError(err)
		}
		return nil, err
	}
	return &msg, nil
}

func (s *Store) GetLatestWelcomeMessage(ctx context.Context, installationKey []byte) (*WelcomeMessage, error) {
	var msg WelcomeMessage
	err := s.db.NewSelect().Model(&msg).Where("installation_key = ?", installationKey).Order("id DESC").Limit(1).Scan(ctx)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, NewNotFoundError(err)
		}
		return nil, err
	}
	return &msg, nil
}

func (s *Store) QueryGroupMessagesV1(ctx context.Context, req *mlsv1.QueryGroupMessagesRequest) ([]*GroupMessage, error) {
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

	pageSize := MaxQueryPageSize
	if req.PagingInfo != nil && req.PagingInfo.Limit > 0 && req.PagingInfo.Limit <= MaxQueryPageSize {
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

	return msgs, nil
}

func (s *Store) QueryWelcomeMessagesV1(ctx context.Context, req *mlsv1.QueryWelcomeMessagesRequest) ([]*WelcomeMessage, error) {
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

	pageSize := MaxQueryPageSize
	if req.PagingInfo != nil && req.PagingInfo.Limit > 0 && req.PagingInfo.Limit <= MaxQueryPageSize {
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

	return msgs, nil
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

type NotFoundError struct {
	Err error
}

func (e *NotFoundError) Error() string {
	return e.Err.Error()
}

func NewNotFoundError(err error) *NotFoundError {
	return &NotFoundError{err}
}

func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}
