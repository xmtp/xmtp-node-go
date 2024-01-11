package store

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	mlsv1 "github.com/xmtp/proto/v3/go/mls/api/v1"
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
	QueryGroupMessages(ctx context.Context, query *mlsv1.QueryGroupMessagesRequest) ([]*GroupMessage, error)
	InsertWelcomeMessage(ctx context.Context, installationId string, data []byte) (*WelcomeMessage, error)
	QueryWelcomeMessages(ctx context.Context, query *mlsv1.QueryWelcomeMessagesRequest) ([]*WelcomeMessage, error)
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

func (s *Store) QueryGroupMessages(ctx context.Context, req *mlsv1.QueryGroupMessagesRequest) ([]*GroupMessage, error) {
	msgs := make([]*GroupMessage, 0)

	err := s.db.NewSelect().
		Model(&msgs).
		Where("group_id = ?", req.GroupId).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return msgs, nil

	// messages := make([]*GroupMessage, 0)

	// if len(query.ContentTopics) == 0 {
	// 	return nil, errors.New("topic is required")
	// }

	// if len(query.ContentTopics) > 1 {
	// 	return nil, errors.New("multiple topics not supported")
	// }

	// q := s.db.NewSelect().
	// 	Model(&messages).
	// 	Where("topic = ?", query.ContentTopics[0])

	// direction := messagev1.SortDirection_SORT_DIRECTION_DESCENDING
	// if query.PagingInfo != nil && query.PagingInfo.Direction != messagev1.SortDirection_SORT_DIRECTION_UNSPECIFIED {
	// 	direction = query.PagingInfo.Direction
	// }
	// switch direction {
	// case messagev1.SortDirection_SORT_DIRECTION_DESCENDING:
	// 	q = q.Order("tid DESC")
	// case messagev1.SortDirection_SORT_DIRECTION_ASCENDING:
	// 	q = q.Order("tid ASC")
	// }

	// pageSize := maxPageSize
	// if query.PagingInfo != nil && query.PagingInfo.Limit > 0 && query.PagingInfo.Limit <= maxPageSize {
	// 	pageSize = int(query.PagingInfo.Limit)
	// }
	// q = q.Limit(pageSize)

	// if query.PagingInfo != nil && query.PagingInfo.GetCursor() != nil && query.PagingInfo.GetCursor().GetIndex() != nil {
	// 	index := query.PagingInfo.GetCursor().GetIndex()
	// 	if index.SenderTimeNs == 0 || len(index.Digest) == 0 {
	// 		return nil, errors.New("invalid cursor")
	// 	}
	// 	cursorTID := buildMessageTIDFromCursor(index)
	// 	if direction == messagev1.SortDirection_SORT_DIRECTION_ASCENDING {
	// 		q = q.Where("tid > ?", cursorTID)
	// 	} else {
	// 		q = q.Where("tid < ?", cursorTID)
	// 	}
	// } else {
	// 	if query.StartTimeNs > 0 {
	// 		q = q.Where("created_at >= ?", query.StartTimeNs)
	// 	}
	// 	if query.EndTimeNs > 0 {
	// 		q = q.Where("created_at <= ?", query.EndTimeNs)
	// 	}
	// }

	// err := q.Scan(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// envs := make([]*messagev1.Envelope, 0, len(messages))
	// for _, msg := range messages {
	// 	envs = append(envs, &messagev1.Envelope{
	// 		ContentTopic: msg.Topic,
	// 		TimestampNs:  uint64(msg.CreatedAt),
	// 		Message:      msg.Content,
	// 	})
	// }

	// pagingInfo := &messagev1.PagingInfo{Limit: 0, Cursor: nil, Direction: direction}
	// if len(envs) >= pageSize {
	// 	if len(envs) > 0 {
	// 		lastEnv := envs[len(envs)-1]
	// 		digest := buildMessageContentDigest(lastEnv.Message)
	// 		pagingInfo.Cursor = &messagev1.Cursor{
	// 			Cursor: &messagev1.Cursor_Index{
	// 				Index: &messagev1.IndexCursor{
	// 					Digest:       digest[:],
	// 					SenderTimeNs: lastEnv.TimestampNs,
	// 				},
	// 			},
	// 		}
	// 	}
	// }

	// return &messagev1.QueryResponse{
	// 	Envelopes:  envs,
	// 	PagingInfo: pagingInfo,
	// }, nil
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

func (s *Store) QueryWelcomeMessages(ctx context.Context, req *mlsv1.QueryWelcomeMessagesRequest) ([]*WelcomeMessage, error) {
	msgs := make([]*WelcomeMessage, 0)

	err := s.db.NewSelect().
		Model(&msgs).
		Where("installation_id = ?", req.InstallationId).
		Order("created_at DESC").
		Scan(ctx)
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
