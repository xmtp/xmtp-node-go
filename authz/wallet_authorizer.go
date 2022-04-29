package authz

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type Permission int64

const REFRESH_INTERVAL_MINUTES = 5

const (
	Allowed     Permission = 0
	Denied      Permission = 1
	Unspecified Permission = 2
)

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

type WalletAuthorizer interface {
	Start() error
	Stop() error
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
}

func NewDatabaseWalletAuthorizer(db *bun.DB, log *zap.Logger) *DatabaseWalletAuthorizer {
	result := new(DatabaseWalletAuthorizer)
	result.db = db
	result.log = log
	result.permissions = make(map[string]Permission)
	result.refreshInterval = time.Minute * REFRESH_INTERVAL_MINUTES

	return result
}

func (d *DatabaseWalletAuthorizer) GetPermissions(walletAddress string) Permission {
	permission, hasPermission := d.permissions[walletAddress]
	if !hasPermission {
		return Unspecified
	}
	return permission
}

func (d *DatabaseWalletAuthorizer) IsAllowListed(walletAddress string) bool {
	return d.GetPermissions(walletAddress) == Allowed
}

func (d *DatabaseWalletAuthorizer) IsDenyListed(walletAddress string) bool {
	return d.GetPermissions(walletAddress) == Denied
}

func (d *DatabaseWalletAuthorizer) Start(ctx context.Context) error {
	d.ctx = ctx
	err := d.loadPermissions()
	if err != nil {
		return err
	}
	go d.listenForChanges()
	return nil
}

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

func (d *DatabaseWalletAuthorizer) listenForChanges() {
	ticker := time.NewTicker(d.refreshInterval)

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			err := d.loadPermissions()
			if err != nil {
				d.log.Error("Error reading permissions from DB", zap.Error(err))
			}
		}
	}
}
