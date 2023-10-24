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
			AND not_consumed = TRUE
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
			Set("not_consumed = FALSE").
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

func NewKeyPackage(installationId string, data []byte, isLastResort bool) *KeyPackage {
	return &KeyPackage{
		ID:             buildKeyPackageId(data),
		InstallationId: installationId,
		CreatedAt:      nowNs(),
		IsLastResort:   isLastResort,
		NotConsumed:    true,
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
