package authz

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/migrations/authz"
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

func createDB(t *testing.T) (db *bun.DB) {
	db = newDB()
	migrator := migrate.NewMigrator(db, authz.Migrations)
	ctx := context.Background()
	err := migrator.Init(ctx)
	require.NoError(t, err)
	_, err = migrator.Migrate(ctx)
	require.NoError(t, err)
	_, err = db.Exec("TRUNCATE TABLE authz_addresses")
	require.NoError(t, err)

	return
}

func fillDb(t *testing.T, db *bun.DB) (wallets []WalletAddress) {
	allowWallet := WalletAddress{
		WalletAddress: "0x1234",
		Permission:    "allow",
	}
	denyWallet := WalletAddress{
		WalletAddress: "0x5678",
		Permission:    "deny",
	}
	wallets = []WalletAddress{allowWallet, denyWallet}
	_, err := db.NewInsert().Model(&wallets).Exec(context.Background())

	require.NoError(t, err)

	return
}

func TestPermissionCheck(t *testing.T) {
	db := createDB(t)
	wallets := fillDb(t, db)
	logger, _ := zap.NewDevelopment()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	authorizer := NewDatabaseWalletAuthorizer(db, logger)
	err := authorizer.Start(ctx)
	require.NoError(t, err)

	for _, wallet := range wallets {
		expectedValue := wallet.Permission
		isAllowed := authorizer.IsAllowListed(wallet.WalletAddress)
		isDenied := authorizer.IsDenyListed(wallet.WalletAddress)
		permission := authorizer.GetPermissions(wallet.WalletAddress)
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

func TestUnknownWallet(t *testing.T) {
	db := createDB(t)
	logger, _ := zap.NewDevelopment()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	authorizer := NewDatabaseWalletAuthorizer(db, logger)
	err := authorizer.Start(ctx)
	require.NoError(t, err)

	unknownWalletAddress := "0xfoo"

	require.Equal(t, authorizer.GetPermissions(unknownWalletAddress), Unspecified)
	require.Equal(t, authorizer.IsAllowListed(unknownWalletAddress), false)
	require.Equal(t, authorizer.IsDenyListed(unknownWalletAddress), false)
}
