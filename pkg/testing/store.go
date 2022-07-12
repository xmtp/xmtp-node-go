package testing

import (
	"database/sql"
	"testing"

	"github.com/status-im/go-waku/waku/v2/node"
	wakustore "github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/driver/pgdriver"
)

const (
	localTestDBDSNPrefix = "postgres://postgres:xmtp@localhost:5432"
	localTestDBDSNSuffix = "?sslmode=disable"
)

func NewDB(t *testing.T) (*sql.DB, string, func()) {
	dsn := localTestDBDSNPrefix + localTestDBDSNSuffix
	ctlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	dbName := "test_" + RandomStringLower(5)
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

func ConnectStoreNode(t *testing.T, n1, n2 *node.WakuNode) {
	Connect(t, n1, n2, string(wakustore.StoreID_v20beta4))
}
