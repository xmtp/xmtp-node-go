package mlsstore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/pkg/migrations/messages"
	"go.uber.org/zap"
)

type Store struct {
	config Config
	log    *zap.Logger
	db     *bun.DB
}

type MlsStore interface {
	CreateInstallation(ctx context.Context, installationId string, walletAddress string, lastResortKeyPackage []byte) error
	InsertKeyPackages(ctx context.Context, keyPackages []*KeyPackage) error
	ConsumeKeyPackages(ctx context.Context, installationIds []string) ([]*KeyPackage, error)
	GetIdentityUpdates(ctx context.Context, walletAddresses []string, startTimeNs int64) (map[string]IdentityUpdateList, error)
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

func (s *Store) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// Creates the installation and last resort key package
func (s *Store) CreateInstallation(ctx context.Context, installationId string, walletAddress string, lastResortKeyPackage []byte) error {
	createdAt := nowNs()

	installation := Installation{
		ID:            installationId,
		WalletAddress: walletAddress,
		CreatedAt:     createdAt,
	}

	keyPackage := NewKeyPackage(installationId, lastResortKeyPackage, true)

	return s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().
			Model(&installation).
			Ignore().
			Exec(ctx)

		if err != nil {
			return err
		}

		_, err = tx.NewInsert().
			Model(&keyPackage).
			Ignore().
			Exec(ctx)

		if err != nil {
			return err
		}

		return nil
	})
}

// Insert a batch of key packages, ignoring any that may already exist
func (s *Store) InsertKeyPackages(ctx context.Context, keyPackages []*KeyPackage) error {
	_, err := s.db.NewInsert().Model(&keyPackages).Ignore().Exec(ctx)
	return err
}

func (s *Store) ConsumeKeyPackages(ctx context.Context, installationIds []string) ([]*KeyPackage, error) {
	keyPackages := make([]*KeyPackage, 0)
	err := s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		err := tx.NewRaw(`
			SELECT DISTINCT ON(installation_id) * FROM key_packages
			WHERE "installation_id" IN (?)
			AND "consumed_at" IS NULL
			ORDER BY installation_id ASC, is_last_resort ASC, created_at ASC
			`,
			bun.In(installationIds)).
			Scan(ctx, &keyPackages)

		if err != nil {
			return err
		}

		if len(keyPackages) < len(installationIds) {
			return errors.New("key packages not found")
		}

		_, err = tx.NewUpdate().
			Table("key_packages").
			Set("consumed_at = ?", nowNs()).
			Where("is_last_resort = FALSE").
			Where("id IN (?)", bun.In(extractIds(keyPackages))).
			Exec(ctx)

		return err
	})

	if err != nil {
		return nil, err
	}

	return keyPackages, nil
}

type IdentityUpdateKind int

const (
	Create IdentityUpdateKind = iota
	Revoke
)

type IdentityUpdate struct {
	Kind           IdentityUpdateKind
	InstallationId string
	TimestampNs    uint64
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

func (s *Store) GetIdentityUpdates(ctx context.Context, walletAddresses []string, startTimeNs int64) (map[string]IdentityUpdateList, error) {
	updated := make([]*Installation, 0)
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

	out := make(map[string]IdentityUpdateList)
	for _, installation := range updated {
		if installation.CreatedAt > startTimeNs {
			out[installation.WalletAddress] = append(out[installation.WalletAddress], IdentityUpdate{
				Kind:           Create,
				InstallationId: installation.ID,
				TimestampNs:    uint64(installation.CreatedAt),
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

	return out, nil
}

func (s *Store) RevokeInstallation(ctx context.Context, installationId string) error {
	_, err := s.db.NewUpdate().
		Model(&Installation{}).
		Set("revoked_at = ?", nowNs()).
		Where("id = ?", installationId).
		Where("revoked_at IS NULL").
		Exec(ctx)

	return err
}

func NewKeyPackage(installationId string, data []byte, isLastResort bool) KeyPackage {
	return KeyPackage{
		ID:             buildKeyPackageId(data),
		InstallationId: installationId,
		CreatedAt:      nowNs(),
		IsLastResort:   isLastResort,
		Data:           data,
	}
}

func extractIds(keyPackages []*KeyPackage) []string {
	out := make([]string, len(keyPackages))
	for i, keyPackage := range keyPackages {
		out[i] = keyPackage.ID
	}
	return out
}

func (s *Store) migrate(ctx context.Context) error {
	migrator := migrate.NewMigrator(s.db, messages.Migrations)
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

func buildKeyPackageId(keyPackageData []byte) string {
	digest := sha256.Sum256(keyPackageData)
	return hex.EncodeToString(digest[:])
}
