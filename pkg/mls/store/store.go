package store

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	mlsv1 "github.com/xmtp/proto/v3/go/mls/api/v1"
	"github.com/xmtp/proto/v3/go/mls/message_contents"
	migrations "github.com/xmtp/xmtp-node-go/pkg/migrations/mls"
	"go.uber.org/zap"
)

const maxPageSize = 100

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
	InsertGroupMessage(ctx context.Context, groupId string, data []byte) (*GroupMessage, error)
	InsertWelcomeMessage(ctx context.Context, installationId string, data []byte) (*WelcomeMessage, error)
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
	createdAt := s.nowNs()

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
		UpdatedAt: s.nowNs(),

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
				InstallationId:     installation.ID,
				CredentialIdentity: installation.CredentialIdentity,
				TimestampNs:        uint64(installation.CreatedAt),
			})
		}
		if installation.RevokedAt != nil && *installation.RevokedAt > startTimeNs {
			out[installation.WalletAddress] = append(out[installation.WalletAddress], IdentityUpdate{
				Kind:           Revoke,
				InstallationId: installation.ID,
				TimestampNs:    uint64(*installation.RevokedAt),
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
		Set("revoked_at = ?", s.nowNs()).
		Where("id = ?", installationId).
		Where("revoked_at IS NULL").
		Exec(ctx)

	return err
}

func (s *Store) InsertGroupMessage(ctx context.Context, groupId string, data []byte) (*GroupMessage, error) {
	message := GroupMessage{
		Data: data,
	}

	var id uint64
	err := s.db.QueryRow("INSERT INTO group_messages (group_id, data) VALUES (?, ?) RETURNING id", groupId, data).Scan(&id)
	if err != nil {
		return nil, err
	}

	err = s.db.NewSelect().Model(&message).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (s *Store) InsertWelcomeMessage(ctx context.Context, installationId string, data []byte) (*WelcomeMessage, error) {
	message := WelcomeMessage{
		Data: data,
	}

	var id uint64
	err := s.db.QueryRow("INSERT INTO welcome_messages (installation_id, data) VALUES (?, ?) RETURNING id", installationId, data).Scan(&id)
	if err != nil {
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

	if req.GroupId == "" {
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

	if req.PagingInfo != nil && req.PagingInfo.Cursor != 0 {
		if direction == mlsv1.SortDirection_SORT_DIRECTION_ASCENDING {
			q = q.Where("id > ?", req.PagingInfo.Cursor)
		} else {
			q = q.Where("id < ?", req.PagingInfo.Cursor)
		}
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}

	messages := make([]*message_contents.GroupMessage, 0, len(msgs))
	for _, msg := range msgs {
		messages = append(messages, &message_contents.GroupMessage{
			Version: &message_contents.GroupMessage_V1_{
				V1: &message_contents.GroupMessage_V1{
					Id:        msg.Id,
					CreatedNs: uint64(msg.CreatedAt.UnixNano()),
					GroupId:   msg.GroupId,
					Data:      msg.Data,
				},
			},
		})
	}

	pagingInfo := &mlsv1.PagingInfo{Limit: 0, Cursor: 0, Direction: direction}
	if len(messages) >= pageSize {
		if len(messages) > 0 {
			lastMsg := msgs[len(messages)-1]
			pagingInfo.Cursor = lastMsg.Id
		}
	}

	return &mlsv1.QueryGroupMessagesResponse{
		Messages:   messages,
		PagingInfo: pagingInfo,
	}, nil
}

func (s *Store) QueryWelcomeMessagesV1(ctx context.Context, req *mlsv1.QueryWelcomeMessagesRequest) (*mlsv1.QueryWelcomeMessagesResponse, error) {
	msgs := make([]*WelcomeMessage, 0)

	if req.InstallationId == "" {
		return nil, errors.New("installation is required")
	}

	q := s.db.NewSelect().
		Model(&msgs).
		Where("installation_id = ?", req.InstallationId)

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

	if req.PagingInfo != nil && req.PagingInfo.Cursor != 0 {
		if direction == mlsv1.SortDirection_SORT_DIRECTION_ASCENDING {
			q = q.Where("id > ?", req.PagingInfo.Cursor)
		} else {
			q = q.Where("id < ?", req.PagingInfo.Cursor)
		}
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}

	messages := make([]*message_contents.WelcomeMessage, 0, len(msgs))
	for _, msg := range msgs {
		messages = append(messages, &message_contents.WelcomeMessage{
			Version: &message_contents.WelcomeMessage_V1_{
				V1: &message_contents.WelcomeMessage_V1{
					Id:        msg.Id,
					CreatedNs: uint64(msg.CreatedAt.UnixNano()),
					Data:      msg.Data,
				},
			},
		})
	}

	pagingInfo := &mlsv1.PagingInfo{Limit: 0, Cursor: 0, Direction: direction}
	if len(messages) >= pageSize {
		if len(messages) > 0 {
			lastMsg := msgs[len(messages)-1]
			pagingInfo.Cursor = lastMsg.Id
		}
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

func (s *Store) nowNs() int64 {
	return s.config.now().UnixNano()
}

type IdentityUpdateKind int

const (
	Create IdentityUpdateKind = iota
	Revoke
)

type IdentityUpdate struct {
	Kind               IdentityUpdateKind
	InstallationId     []byte
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
