package authz

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.uber.org/zap"
)

func newDB() *bun.DB {
	dsn, hasDsn := os.LookupEnv("AUTHZ_POSTGRES_CONNECTION_STRING")
	if !hasDsn {
		dsn = "postgres://postgres:xmtp@localhost:6543/postgres?sslmode=disable"
	}
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	return db
}

func newAllowLister(t *testing.T) *DatabaseWalletAllowLister {
	db := newDB()
	logger, _ := zap.NewDevelopment()
	allowLister := NewDatabaseWalletAllowLister(db, logger)

	return allowLister
}

func fillDb(t *testing.T, db *bun.DB) (wallets []WalletAddress) {
	_, err := db.Exec("truncate table authz_addresses;")
	require.NoError(t, err)

	allowWallet := WalletAddress{
		WalletAddress: "0x1234",
		Permission:    "allow",
	}
	denyWallet := WalletAddress{
		WalletAddress: "0x5678",
		Permission:    "deny",
	}
	wallets = []WalletAddress{allowWallet, denyWallet}
	_, err = db.NewInsert().Model(&wallets).Exec(context.Background())

	require.NoError(t, err)

	return
}

func TestPermissionCheck(t *testing.T) {
	allowLister := newAllowLister(t)
	err := allowLister.migrate(context.Background())
	require.NoError(t, err)
	wallets := fillDb(t, allowLister.db)
	err = allowLister.Start(context.Background())
	require.NoError(t, err)

	defer allowLister.Stop()

	for _, wallet := range wallets {
		expectedValue := wallet.Permission
		isAllowed := allowLister.IsAllowListed(wallet.WalletAddress)
		isDenied := allowLister.IsDenyListed(wallet.WalletAddress)
		permission := allowLister.GetPermissions(wallet.WalletAddress)
		if expectedValue == "allow" {
			require.Equal(t, isAllowed, true)
			require.Equal(t, isDenied, false)
			require.Equal(t, permission, Allowed)
		} else {
			require.Equal(t, isAllowed, false)
			require.Equal(t, isDenied, true)
			require.Equal(t, permission, Denied)
		}
	}
}

func TestDelete(t *testing.T) {
	allowLister := newAllowLister(t)
	allowLister.refreshInterval = 100 * time.Millisecond
	wallets := fillDb(t, allowLister.db)
	allowedWallet := wallets[0]

	err := allowLister.Start(context.Background())
	require.NoError(t, err)
	defer allowLister.Stop()

	require.Equal(t, allowLister.IsAllowListed(allowedWallet.WalletAddress), true)

	// Delete the allowed wallet record
	require.NotNil(t, allowedWallet.ID)
	updateModel := WalletAddress{ID: allowedWallet.ID}
	now := time.Now().UTC()
	_, err = allowLister.db.NewUpdate().
		Model(&updateModel).Set("deleted_at = ?", now).
		Where("id = ?", allowedWallet.ID).
		Exec(context.Background())

	require.NoError(t, err)
	// Sleep to wait for the refresh to happen behind the scenes
	time.Sleep(200 * time.Millisecond)
	require.Equal(t, allowLister.IsAllowListed(allowedWallet.WalletAddress), false)
}

func TestUnknownWallet(t *testing.T) {
	allowLister := newAllowLister(t)
	err := allowLister.Start(context.Background())
	require.NoError(t, err)
	defer allowLister.Stop()

	unknownWalletAddress := "0xfoo"

	require.Equal(t, allowLister.GetPermissions(unknownWalletAddress), Unspecified)
	require.Equal(t, allowLister.IsAllowListed(unknownWalletAddress), false)
	require.Equal(t, allowLister.IsDenyListed(unknownWalletAddress), false)
}
