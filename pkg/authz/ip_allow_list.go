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

const REFRESH_INTERVAL_SECONDS = 60

// DatabaseAllowList implements database backed allow list.
type DatabaseAllowList struct {
	db              *bun.DB
	log             *zap.Logger
	permissions     map[string]Permission
	permissionsLock sync.RWMutex
	refreshInterval time.Duration
	ctx             context.Context
	cancelFunc      context.CancelFunc
	wg              sync.WaitGroup
}

func NewDatabaseAllowList(
	ctx context.Context,
	db *bun.DB,
	log *zap.Logger,
) (*DatabaseAllowList, error) {
	ctx, cancel := context.WithCancel(ctx)
	allowLister := DatabaseAllowList{
		db:              db,
		log:             log.Named("allowlist"),
		permissions:     make(map[string]Permission),
		refreshInterval: time.Second * REFRESH_INTERVAL_SECONDS,
		ctx:             ctx,
		cancelFunc:      cancel,
		wg:              sync.WaitGroup{},
	}

	if err := allowLister.start(); err != nil {
		return nil, err
	}

	return &allowLister, nil
}

// Get the permissions for a IP address
func (d *DatabaseAllowList) GetPermission(ipAddress string) Permission {
	d.permissionsLock.RLock()
	defer d.permissionsLock.RUnlock()

	permission, hasPermission := d.permissions[ipAddress]
	if !hasPermission {
		return Unspecified
	}
	return permission
}

// Load the permissions and start listening
func (d *DatabaseAllowList) start() error {
	err := d.migrate(d.ctx)
	if err != nil {
		return err
	}
	err = d.loadPermissions()
	if err != nil {
		return err
	}

	tracing.GoPanicWrap(
		d.ctx,
		&d.wg,
		"authz-change-listener",
		func(_ context.Context) { d.listenForChanges() },
	)

	d.log.Info("started")
	return nil
}

// Actually load latest permissions from the database
// Currently just loads absolutely everything, since n is going to be small enough for now.
// Should be possible to do incrementally if we need to using the created_at field
func (d *DatabaseAllowList) loadPermissions() error {
	d.permissionsLock.Lock()
	defer d.permissionsLock.Unlock()

	var ips []IPAddress
	query := d.db.NewSelect().Model(&ips).Where("deleted_at IS NULL")
	if err := query.Scan(d.ctx); err != nil {
		return err
	}

	newPermissionMap := make(map[string]Permission)
	for _, ip := range ips {
		permission := permissionFromString(ip.Permission)
		if permission == Unspecified {
			d.log.Warn(
				"Unknown permission in DB",
				zap.String("ip", ip.IPAddress),
				zap.String("permission", ip.Permission),
			)
		}
		newPermissionMap[ip.IPAddress] = permission
	}
	d.permissions = newPermissionMap

	d.log.Debug(
		"Updated allow/deny lists from the database",
		zap.Int("num_values", len(newPermissionMap)),
	)

	return nil
}

func (d *DatabaseAllowList) migrate(ctx context.Context) error {
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
		d.log.Info("No new migrations to run for DatabaseAllowList")
	}

	return nil
}

func (d *DatabaseAllowList) listenForChanges() {
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

func (d *DatabaseAllowList) Stop() {
	d.cancelFunc()
	d.wg.Wait()
	d.log.Info("stopped")
}
