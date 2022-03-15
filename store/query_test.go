package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Blank import to register the sqlite3 driver

	"github.com/status-im/go-waku/tests"
	"github.com/status-im/go-waku/waku/persistence"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/stretchr/testify/require"
)

func NewMock() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		tests.Logger().Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db
}

func createIndex(digest []byte, receiverTime int64) *pb.Index {
	return &pb.Index{
		Digest:       digest,
		ReceiverTime: receiverTime,
		SenderTime:   1.0,
	}
}

func createAndFillDb(t *testing.T) *sql.DB {
	db := NewMock()
	option := persistence.WithDB(db)
	store, err := persistence.NewDBStore(tests.Logger(), option)
	require.NoError(t, err)

	res, err := store.GetAll()
	require.NoError(t, err)
	require.Empty(t, res)

	err = store.Put(
		createIndex([]byte("digest"), 1),
		"test",
		tests.CreateWakuMessage("test1", 1),
	)
	require.NoError(t, err)

	err = store.Put(
		createIndex([]byte("digest"), 2),
		"test",
		tests.CreateWakuMessage("test2", 2),
	)
	require.NoError(t, err)

	err = store.Put(
		createIndex([]byte("digest"), 3),
		"test",
		tests.CreateWakuMessage("test3", 3),
	)
	require.NoError(t, err)

	return db
}

func TestQuerySimple(t *testing.T) {
	db := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{
		ContentFilters: []*pb.ContentFilter{
			{
				ContentTopic: "test2",
			},
		},
	})

	require.NoError(t, err)
	require.Len(t, result.Messages, 1)
	require.Equal(t, result.Messages[0].ContentTopic, "test2")
	require.Equal(t, result.Messages[0].Payload, []byte{1, 2, 3})
}

func TestQueryMultipleTopics(t *testing.T) {
	db := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{
		ContentFilters: []*pb.ContentFilter{
			{
				ContentTopic: "test2",
			},
			{
				ContentTopic: "test3",
			},
		},
	})

	require.NoError(t, err)
	require.Len(t, result.Messages, 2)
}

func TestQueryNoTopics(t *testing.T) {
	db := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{})

	require.NoError(t, err)
	require.Len(t, result.Messages, 3)
}

func TestQueryStartTime(t *testing.T) {
	db := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{
		StartTime: 2,
	})

	require.NoError(t, err)
	require.Len(t, result.Messages, 2)
}

func TestQueryTimeWindow(t *testing.T) {
	db := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{
		StartTime: 2,
		EndTime:   3,
	})

	require.NoError(t, err)
	require.Len(t, result.Messages, 2)
	require.Equal(t, result.Messages[0].Timestamp, int64(2))
}

func TestQueryTimeAndTopic(t *testing.T) {
	db := createAndFillDb(t)

	resultNoHits, err := FindMessages(db, &pb.HistoryQuery{
		ContentFilters: []*pb.ContentFilter{
			{
				// Wrong content topic. Should exclude messages
				ContentTopic: "test2",
			},
		},
		StartTime: 3,
		EndTime:   4,
	})

	require.NoError(t, err)
	require.Equal(t, len(resultNoHits.Messages), 0)

	resultWithHits, err := FindMessages(db, &pb.HistoryQuery{
		ContentFilters: []*pb.ContentFilter{
			{
				ContentTopic: "test3",
			},
		},
		StartTime: 3,
		EndTime:   4,
	})
	require.NoError(t, err)
	require.Len(t, resultWithHits.Messages, 1)
}

func TestQueryPubsubTopic(t *testing.T) {
	db := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{
		PubsubTopic: "test",
	})

	require.NoError(t, err)
	require.Equal(t, len(result.Messages), 3)

	result, err = FindMessages(db, &pb.HistoryQuery{
		PubsubTopic: "foo",
	})
	require.NoError(t, err)
	require.Len(t, result.Messages, 0)
}
