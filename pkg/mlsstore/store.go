package mlsstore

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	mlsMigrations "github.com/xmtp/xmtp-node-go/pkg/migrations/mls"
	"go.uber.org/zap"
)

// TODO(snormore): implements metrics + cleaner

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
		Set("revoked_at = ?", nowNs()).
		Where("id = ?", installationId).
		Where("revoked_at IS NULL").
		Exec(ctx)

	return err
}

func (s *Store) InsertMessage(ctx context.Context, topic string, content []byte) (*messagev1.Envelope, error) {
	createdAt := nowNs()

	message := Message{
		TID:       topic + "",
		Topic:     topic,
		CreatedAt: createdAt,
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
		TimestampNs:  uint64(createdAt),
		Message:      content,
	}, nil
}

func (s *Store) QueryMessages(ctx context.Context, query *messagev1.QueryRequest) (*messagev1.QueryResponse, error) {
	messages := make([]*Message, 0)

	err := s.db.NewSelect().
		Model(&messages).
		Where("topic IN (?)", bun.In(query.ContentTopics)).
		// WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
		// 	return q.Where("created_at > ?", startTimeNs).WhereOr("revoked_at > ?", startTimeNs)
		// }).
		Order("created_at ASC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	// TODO: move api to api/mls/v1
	// TODO: update db schema to support querying by multiple topics, so tid should not have topic in it, I guess
	// TODO: finish implementing all of this (query request variations, pagination, etc)

	envs := make([]*messagev1.Envelope, 0, len(messages))
	for _, msg := range messages {
		envs = append(envs, &messagev1.Envelope{
			ContentTopic: msg.Topic,
			TimestampNs:  uint64(msg.CreatedAt),
			Message:      msg.Content,
		})
	}
	return &messagev1.QueryResponse{
		Envelopes: envs,
	}, nil
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
