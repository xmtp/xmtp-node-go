package authz

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/migrations/authz"
	"go.uber.org/zap"
)

const REFRESH_INTERVAL_MINUTES = 5

type Permission int64

const (
	Allowed     Permission = 0
	Denied      Permission = 1
	Unspecified Permission = 2
)

// Add a string function to use for logging
func (p Permission) String() string {
	switch p {
	case Allowed:
		return "allowed"
	case Denied:
		return "denied"
	case Unspecified:
		return "unspecified"
	}
	return "unknown"
}

// WalletAuthorizer interface
type WalletAuthorizer interface {
	Start() error
	Stop()
	IsAllowListed(walletAddress string) bool
	IsDenyListed(walletAddress string) bool
	GetPermissions(walletAddress string) Permission
}

type DatabaseWalletAuthorizer struct {
	db              *bun.DB
	log             *zap.Logger
	permissions     map[string]Permission
	refreshInterval time.Duration
	ctx             context.Context
	cancelFunc      context.CancelFunc
}

func NewDatabaseWalletAuthorizer(db *bun.DB, log *zap.Logger) *DatabaseWalletAuthorizer {
	result := new(DatabaseWalletAuthorizer)
	result.db = db
	result.log = log
	result.permissions = make(map[string]Permission)
	result.refreshInterval = time.Minute * REFRESH_INTERVAL_MINUTES

	return result
}

// Get the permissions for a wallet address
func (d *DatabaseWalletAuthorizer) GetPermissions(walletAddress string) Permission {
	permission, hasPermission := d.permissions[walletAddress]
	if !hasPermission {
		return Unspecified
	}
	return permission
}

// Check if the wallet address is explicitly allow listed
func (d *DatabaseWalletAuthorizer) IsAllowListed(walletAddress string) bool {
	return d.GetPermissions(walletAddress) == Allowed
}

// Check if the wallet address is explicitly deny listed
func (d *DatabaseWalletAuthorizer) IsDenyListed(walletAddress string) bool {
	return d.GetPermissions(walletAddress) == Denied
}

// Load the permissions and start listening
func (d *DatabaseWalletAuthorizer) Start(ctx context.Context) error {
	newCtx, cancel := context.WithCancel(ctx)
	d.ctx = newCtx
	d.cancelFunc = cancel
	err := d.migrate(ctx)
	if err != nil {
		return err
	}
	err = d.loadPermissions()
	if err != nil {
		return err
	}
	go d.listenForChanges()

	d.log.Info("Started DatabaseWalletAuthorizer")
	return nil
}

// Actually load latest permissions from the database
// Currently just loads absolutely everything, since n is going to be small enough for now.
// Should be possible to do incrementally if we need to using the created_at field
func (d *DatabaseWalletAuthorizer) loadPermissions() error {
	var wallets []WalletAddress
	query := d.db.NewSelect().Model(&wallets).Where("deleted_at IS NULL")
	if err := query.Scan(d.ctx); err != nil {
		return err
	}
	newPermissionMap := make(map[string]Permission)
	for _, wallet := range wallets {
		newPermissionMap[wallet.WalletAddress] = mapPermission(wallet.Permission)
	}
	d.permissions = newPermissionMap

	d.log.Info("Updated allow/deny lists from the database", zap.Int("num_values", len(newPermissionMap)))

	return nil
}

func (d *DatabaseWalletAuthorizer) migrate(ctx context.Context) error {
	migrator := migrate.NewMigrator(d.db, authz.Migrations)
	err := migrator.Init(ctx)
	if err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if group.IsZero() {
		d.log.Info("No new migrations to run")
	}

	return err
}

func mapPermission(permission string) Permission {
	switch permission {
	case "allow":
		return Allowed
	case "deny":
		return Denied
	default:
		return Unspecified
	}
}

func (d *DatabaseWalletAuthorizer) listenForChanges() {
	ticker := time.NewTicker(d.refreshInterval)

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			if err := d.loadPermissions(); err != nil {
				d.log.Error("Error reading permissions from DB", zap.Error(err))
			}
		}
	}
}

func (d *DatabaseWalletAuthorizer) Stop() {
	d.cancelFunc()
}
