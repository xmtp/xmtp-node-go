package authz

import (
	"context"
	"sync"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/pkg/migrations/authz"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
)

const REFRESH_INTERVAL_SECONDS = 5

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

// WalletAllowLister maintains an allow list for wallets.
type WalletAllowLister interface {
	Start(ctx context.Context) error
	Stop()
	IsAllowListed(walletAddress string) bool
	IsDenyListed(walletAddress string) bool
	GetPermissions(walletAddress string) Permission
	Deny(ctx context.Context, WalletAddress string) error
	Allow(ctx context.Context, WalletAddress string) error
}

// DatabaseWalletAllowLister implements database backed allow list.
type DatabaseWalletAllowLister struct {
	db              *bun.DB
	log             *zap.Logger
	permissions     map[string]Permission
	permissionsLock sync.RWMutex
	refreshInterval time.Duration
	ctx             context.Context
	cancelFunc      context.CancelFunc
	wg              sync.WaitGroup
}

func NewDatabaseWalletAllowLister(db *bun.DB, log *zap.Logger) *DatabaseWalletAllowLister {
	result := new(DatabaseWalletAllowLister)
	result.db = db
	result.log = log.Named("wauthz")
	result.permissions = make(map[string]Permission)
	result.refreshInterval = time.Second * REFRESH_INTERVAL_SECONDS

	return result
}

// Get the permissions for a wallet address
func (d *DatabaseWalletAllowLister) GetPermissions(walletAddress string) Permission {
	d.permissionsLock.RLock()
	defer d.permissionsLock.RUnlock()

	permission, hasPermission := d.permissions[walletAddress]
	if !hasPermission {
		return Unspecified
	}
	return permission
}

// Check if the wallet address is explicitly allow listed
func (d *DatabaseWalletAllowLister) IsAllowListed(walletAddress string) bool {
	return d.GetPermissions(walletAddress) == Allowed
}

// Check if the wallet address is explicitly deny listed
func (d *DatabaseWalletAllowLister) IsDenyListed(walletAddress string) bool {
	return d.GetPermissions(walletAddress) == Denied
}

// Add an address to the deny list.
func (d *DatabaseWalletAllowLister) Deny(ctx context.Context, walletAddress string) error {
	return d.Apply(ctx, walletAddress, Denied)
}

// Add an address to the allow list.
func (d *DatabaseWalletAllowLister) Allow(ctx context.Context, walletAddress string) error {
	return d.Apply(ctx, walletAddress, Allowed)
}

func (d *DatabaseWalletAllowLister) Apply(ctx context.Context, walletAddress string, permission Permission) error {
	d.permissionsLock.Lock()
	defer d.permissionsLock.Unlock()

	wallet := WalletAddress{
		WalletAddress: walletAddress,
		Permission:    unmapPermission(permission),
	}
	_, err := d.db.NewInsert().Model(&wallet).Exec(ctx)
	if err != nil {
		return err
	}
	d.permissions[walletAddress] = mapPermission(wallet.Permission)
	return nil
}

// Load the permissions and start listening
func (d *DatabaseWalletAllowLister) Start(ctx context.Context) error {
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
	tracing.GoPanicWrap(ctx, &d.wg, "authz-change-listener", func(_ context.Context) { d.listenForChanges() })

	d.log.Info("started")
	return nil
}

// Actually load latest permissions from the database
// Currently just loads absolutely everything, since n is going to be small enough for now.
// Should be possible to do incrementally if we need to using the created_at field
func (d *DatabaseWalletAllowLister) loadPermissions() error {
	d.permissionsLock.Lock()
	defer d.permissionsLock.Unlock()

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

	d.log.Debug("Updated allow/deny lists from the database", zap.Int("num_values", len(newPermissionMap)))

	return nil
}

func (d *DatabaseWalletAllowLister) migrate(ctx context.Context) error {
	migrator := migrate.NewMigrator(d.db, authz.Migrations)
	err := migrator.Init(ctx)
	if err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		d.log.Info("No new migrations to run for DatabaseWalletAllowLister")
	}

	return nil
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

func unmapPermission(permission Permission) string {
	switch permission {
	case Allowed:
		return "allow"
	case Denied:
		return "deny"
	default:
		return "unspecified"
	}
}

func (d *DatabaseWalletAllowLister) listenForChanges() {
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

func (d *DatabaseWalletAllowLister) Stop() {
	d.cancelFunc()
	d.wg.Wait()
	d.log.Info("stopped")
}
