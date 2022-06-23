package store

import (
	"context"
	"database/sql"
	"math/rand"
	"strings"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/status-im/go-waku/tests"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/driver/pgdriver"
)

func newTestDB(t *testing.T) (*sql.DB, func()) {
	dsn := "postgres://postgres:xmtp@localhost:5432?sslmode=disable"
	ctlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	dbName := randomStringLower(5)
	_, err := ctlDB.Exec("CREATE DATABASE " + dbName)
	require.NoError(t, err)

	dsn = "postgres://postgres:xmtp@localhost:5432/" + dbName + "?sslmode=disable"
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	return db, func() {
		db.Close()
		_, err := ctlDB.Exec("DROP DATABASE " + dbName)
		require.NoError(t, err)
		ctlDB.Close()
	}
}

func newTestStore(t *testing.T) (*XmtpStore, func()) {
	db, dbCleanup := newTestDB(t)
	dbStore, err := NewDBStore(utils.Logger(), WithDB(db))
	require.NoError(t, err)

	host := newTestPeer(t)
	host.Peerstore().AddAddr(host.ID(), tests.GetHostAddress(host), peerstore.PermanentAddrTTL)
	err = host.Peerstore().AddProtocols(host.ID(), string(store.StoreID_v20beta4))
	require.NoError(t, err)

	store := NewXmtpStore(host, db, dbStore, 0, utils.Logger())

	store.Start(context.Background())

	return store, func() {
		store.Stop()
		dbCleanup()
	}
}

func addStoreProtocol(t *testing.T, h1, h2 host.Host) {
	h1.Peerstore().AddAddr(h2.ID(), tests.GetHostAddress(h2), peerstore.PermanentAddrTTL)
	err := h1.Peerstore().AddProtocols(h2.ID(), string(store.StoreID_v20beta4))
	require.NoError(t, err)
}

func newTestEnvelope(t *testing.T, msg *pb.WakuMessage, pubSubTopic string) *protocol.Envelope {
	return protocol.NewEnvelope(msg, utils.GetUnixEpoch(), pubSubTopic)
}

func newTestPeer(t *testing.T) host.Host {
	host, err := libp2p.New(libp2p.DefaultTransports, libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	require.NoError(t, err)
	return host
}

func expectMessages(t *testing.T, s *XmtpStore, pubSubTopic string, msgs []*pb.WakuMessage) {
	res, err := s.FindMessages(&pb.HistoryQuery{
		PubsubTopic: pubSubTopic,
	})
	require.NoError(t, err)
	require.Empty(t, res.Error)
	require.Len(t, res.Messages, len(msgs))
	require.ElementsMatch(t, msgs, res.Messages)
}

func storeMessage(t *testing.T, s *XmtpStore, msg *pb.WakuMessage, pubSubTopic string) {
	err := s.storeMessage(newTestEnvelope(t, msg, pubSubTopic))
	require.NoError(t, err)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func randomStringLower(n int) string {
	return strings.ToLower(randomString(n))
}
