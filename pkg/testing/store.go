package testing

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/pkg/migrations/authz"
	"github.com/xmtp/xmtp-node-go/pkg/migrations/mls"
)

const (
	localTestDBDSNPrefix = "postgres://postgres:xmtp@localhost:15432"
	localTestDBDSNSuffix = "?sslmode=disable"
)

func NewDB(t *testing.T) (*sql.DB, string, func()) {
	dsn := localTestDBDSNPrefix + localTestDBDSNSuffix
	ctlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	dbName := "test_" + RandomStringLower(12)
	_, err := ctlDB.Exec("CREATE DATABASE " + dbName)
	require.NoError(t, err)

	dsn = localTestDBDSNPrefix + "/" + dbName + localTestDBDSNSuffix
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	return db, dsn, func() {
		db.Close()
		_, err = ctlDB.Exec("DROP DATABASE " + dbName)
		require.NoError(t, err)
		ctlDB.Close()
	}
}

func NewAuthzDB(t *testing.T) (*bun.DB, string, func()) {
	db, dsn, cleanup := NewDB(t)
	bunDB := bun.NewDB(db, pgdialect.New())

	ctx := context.Background()
	migrator := migrate.NewMigrator(bunDB, authz.Migrations)
	err := migrator.Init(ctx)
	require.NoError(t, err)
	_, err = migrator.Migrate(ctx)
	require.NoError(t, err)

	return bunDB, dsn, cleanup
}

func NewMLSDB(t *testing.T) (*bun.DB, string, func()) {
	db, dsn, cleanup := NewDB(t)
	bunDB := bun.NewDB(db, pgdialect.New())

	ctx := context.Background()
	migrator := migrate.NewMigrator(bunDB, mls.Migrations)
	err := migrator.Init(ctx)
	require.NoError(t, err)
	_, err = migrator.Migrate(ctx)
	require.NoError(t, err)

	return bunDB, dsn, cleanup
}
