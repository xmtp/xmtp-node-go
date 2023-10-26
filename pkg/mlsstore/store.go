package mlsstore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
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
	CreateInstallation(ctx context.Context, installationId InstallationId, walletAddress string, lastResortKeyPackage []byte, credentialIdentity []byte) error
	InsertKeyPackages(ctx context.Context, keyPackages []*KeyPackage) error
	ConsumeKeyPackages(ctx context.Context, installationIds []InstallationId) ([]*KeyPackage, error)
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

// Creates the installation and last resort key package
func (s *Store) CreateInstallation(ctx context.Context, installationId InstallationId, walletAddress string, lastResortKeyPackage []byte, credentialIdentity []byte) error {
	createdAt := nowNs()

	installation := Installation{
		ID:                 installationId,
		WalletAddress:      walletAddress,
		CreatedAt:          createdAt,
		CredentialIdentity: credentialIdentity,
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
			Model(keyPackage).
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

func (s *Store) ConsumeKeyPackages(ctx context.Context, installationIds []InstallationId) ([]*KeyPackage, error) {
	tx, err := s.db.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	sql, args := buildSelectKeyPackagesQuery(installationIds)
	rows, err := tx.QueryContext(ctx, sql, args...)
	if err != nil {
		panic(err)
		return nil, err
	}
	var keyPackages []*KeyPackage
	if keyPackages, err = rowsToKeyPackages(rows); err != nil {
		panic(err)
		return nil, err
	}

	sql, args = buildConsumeKeyPackagesQuery(extractIds(keyPackages))
	fmt.Println(sql, args)
	if _, err = tx.ExecContext(ctx, sql, args...); err != nil {
		panic(err)
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return keyPackages, nil
}

func buildSelectKeyPackagesQuery(installationIDs []InstallationId) (string, []interface{}) {
	placeholders := make([]string, len(installationIDs))
	args := make([]interface{}, len(installationIDs))
	for i, id := range installationIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	return `
	SELECT DISTINCT ON(installation_id) 
	id, installation_id, created_at, consumed_at, not_consumed, is_last_resort, data
	FROM key_packages
	WHERE "installation_id" IN (` + strings.Join(placeholders, ",") + `)
	AND not_consumed = TRUE
	ORDER BY installation_id ASC, is_last_resort ASC, created_at ASC`, args
}

func rowsToKeyPackages(rows *sql.Rows) ([]*KeyPackage, error) {
	defer rows.Close()
	var result []*KeyPackage

	for rows.Next() {
		kp := &KeyPackage{}
		var installationId InstallationId
		if err := rows.Scan(&kp.ID, &installationId, &kp.CreatedAt, &kp.ConsumedAt, &kp.NotConsumed, &kp.IsLastResort, &kp.Data); err != nil {
			return nil, err
		}
		kp.InstallationId = installationId
		result = append(result, kp)
	}

	return result, nil
}

func buildConsumeKeyPackagesQuery(keyPackageIds []string) (string, []interface{}) {
	args := []interface{}{nowNs()}
	placeHolders := []string{}
	for i, id := range keyPackageIds {
		args = append(args, id)
		placeHolders = append(placeHolders, fmt.Sprintf("$%d", i+2))
	}

	return `UPDATE key_packages
	SET consumed_at = $1, not_consumed = FALSE
	WHERE is_last_resort = FALSE
	AND id in (` + strings.Join(placeHolders, ",") + `)`, args
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
	// Sort the updates by timestamp now that the full list is assembled
	for _, updates := range out {
		sort.Sort(updates)
	}

	return out, nil
}

func (s *Store) RevokeInstallation(ctx context.Context, installationId InstallationId) error {
	_, err := s.db.NewUpdate().
		Model(&Installation{}).
		Set("revoked_at = ?", nowNs()).
		Where("id = ?", installationId).
		Where("revoked_at IS NULL").
		Exec(ctx)

	return err
}

func NewKeyPackage(installationId InstallationId, data []byte, isLastResort bool) *KeyPackage {
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
