package mlsstore

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	mlsMigrations "github.com/xmtp/xmtp-node-go/pkg/migrations/mls"
	"go.uber.org/zap"
)

// TODO(snormore): implements metrics + cleaner

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
	InsertMessage(ctx context.Context, contentTopic string, data []byte) (*messagev1.Envelope, error)
	QueryMessages(ctx context.Context, query *messagev1.QueryRequest) (*messagev1.QueryResponse, error)
	NewKeyPackage(installationId []byte, data []byte, isLastResort bool) *KeyPackage
}

func New(ctx context.Context, config Config) (*Store, error) {
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

func (s *Store) InsertMessage(ctx context.Context, topic string, content []byte) (*messagev1.Envelope, error) {
	createdAt := s.config.now()

	message := Message{
		Topic:     topic,
		TID:       buildMessageTID(createdAt, content),
		CreatedAt: createdAt.UnixNano(),
		Content:   content,
	}

	_, err := s.db.NewInsert().
		Model(&message).
		Ignore().
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return &messagev1.Envelope{
		ContentTopic: topic,
		TimestampNs:  uint64(message.CreatedAt),
		Message:      content,
	}, nil
}

func (s *Store) QueryMessages(ctx context.Context, query *messagev1.QueryRequest) (*messagev1.QueryResponse, error) {
	messages := make([]*Message, 0)

	if len(query.ContentTopics) == 0 {
		return nil, errors.New("topic is required")
	}

	if len(query.ContentTopics) > 1 {
		return nil, errors.New("multiple topics not supported")
	}

	q := s.db.NewSelect().
		Model(&messages).
		Where("topic = ?", query.ContentTopics[0])

	direction := messagev1.SortDirection_SORT_DIRECTION_DESCENDING
	if query.PagingInfo != nil && query.PagingInfo.Direction != messagev1.SortDirection_SORT_DIRECTION_UNSPECIFIED {
		direction = query.PagingInfo.Direction
	}
	switch direction {
	case messagev1.SortDirection_SORT_DIRECTION_DESCENDING:
		q = q.Order("tid DESC")
	case messagev1.SortDirection_SORT_DIRECTION_ASCENDING:
		q = q.Order("tid ASC")
	}

	pageSize := maxPageSize
	if query.PagingInfo != nil && query.PagingInfo.Limit > 0 && query.PagingInfo.Limit <= maxPageSize {
		pageSize = int(query.PagingInfo.Limit)
	}
	q = q.Limit(pageSize)

	if query.PagingInfo != nil && query.PagingInfo.GetCursor() != nil && query.PagingInfo.GetCursor().GetIndex() != nil {
		index := query.PagingInfo.GetCursor().GetIndex()
		if index.SenderTimeNs == 0 || len(index.Digest) == 0 {
			return nil, errors.New("invalid cursor")
		}
		cursorTID := buildMessageTIDFromCursor(index)
		if direction == messagev1.SortDirection_SORT_DIRECTION_ASCENDING {
			q = q.Where("tid > ?", cursorTID)
		} else {
			q = q.Where("tid < ?", cursorTID)
		}
	} else {
		if query.StartTimeNs > 0 {
			q = q.Where("created_at >= ?", query.StartTimeNs)
		}
		if query.EndTimeNs > 0 {
			q = q.Where("created_at <= ?", query.EndTimeNs)
		}
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}

	envs := make([]*messagev1.Envelope, 0, len(messages))
	for _, msg := range messages {
		envs = append(envs, &messagev1.Envelope{
			ContentTopic: msg.Topic,
			TimestampNs:  uint64(msg.CreatedAt),
			Message:      msg.Content,
		})
	}

	pagingInfo := &messagev1.PagingInfo{Limit: 0, Cursor: nil, Direction: direction}
	if len(envs) >= pageSize {
		if len(envs) > 0 {
			lastEnv := envs[len(envs)-1]
			digest := buildMessageContentDigest(lastEnv.Message)
			pagingInfo.Cursor = &messagev1.Cursor{
				Cursor: &messagev1.Cursor_Index{
					Index: &messagev1.IndexCursor{
						Digest:       digest[:],
						SenderTimeNs: lastEnv.TimestampNs,
					},
				},
			}
		}
	}

	return &messagev1.QueryResponse{
		Envelopes:  envs,
		PagingInfo: pagingInfo,
	}, nil
}

func buildMessageTID(createdAt time.Time, content []byte) string {
	return fmt.Sprintf("%s%x", buildMessageTIDPrefix(createdAt), buildMessageContentDigest(content))
}

func buildMessageTIDFromCursor(index *messagev1.IndexCursor) string {
	return fmt.Sprintf("%s%x", buildMessageTIDPrefix(time.Unix(0, int64(index.SenderTimeNs)).UTC()), index.Digest)
}

func buildMessageTIDPrefix(createdAt time.Time) string {
	return fmt.Sprintf("%s.%09d_", createdAt.Format(time.RFC3339), createdAt.Nanosecond())
}

func buildMessageContentDigest(content []byte) []byte {
	digest := sha256.Sum256(content)
	return digest[:]
}

func (s *Store) migrate(ctx context.Context) error {
	migrator := migrate.NewMigrator(s.db, mlsMigrations.Migrations)
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
