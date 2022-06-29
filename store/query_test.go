package store

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Blank import to register the sqlite3 driver
	"github.com/status-im/go-waku/tests"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/driver/pgdriver"
)

const TOPIC = "test"

func NewMock() *sql.DB {
	dsn, hasDsn := os.LookupEnv("MESSAGE_POSTGRES_CONNECTION_STRING")
	if !hasDsn {
		dsn = "postgres://postgres:xmtp@localhost:5432/postgres?sslmode=disable"
	}
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	return db
}

func buildIndex(msg *pb.WakuMessage, topic string) *pb.Index {
	return protocol.NewEnvelope(msg, utils.GetUnixEpoch(), topic).Index()
}

func createStore(t *testing.T, db *sql.DB) *DBStore {
	option := WithDB(db)
	store, err := NewDBStore(utils.Logger(), option)
	require.NoError(t, err)
	return store
}

func createAndFillDb(t *testing.T) (*sql.DB, []*pb.WakuMessage) {
	db := NewMock()
	store := createStore(t, db)
	_, err := db.Exec("TRUNCATE TABLE message;")
	require.NoError(t, err)

	msg1 := tests.CreateWakuMessage("test1", 1)
	msg2 := tests.CreateWakuMessage("test2", 2)
	msg3 := tests.CreateWakuMessage("test3", 3)

	err = store.Put(protocol.NewEnvelope(msg1, utils.GetUnixEpoch(), TOPIC))
	require.NoError(t, err)

	err = store.Put(protocol.NewEnvelope(msg2, utils.GetUnixEpoch(), TOPIC))
	require.NoError(t, err)

	err = store.Put(protocol.NewEnvelope(msg3, utils.GetUnixEpoch(), TOPIC))
	require.NoError(t, err)

	return db, []*pb.WakuMessage{msg1, msg2, msg3}
}

func TestQuerySimple(t *testing.T) {
	db, _ := createAndFillDb(t)

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
	db, _ := createAndFillDb(t)

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
	db, _ := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{})

	require.NoError(t, err)
	require.Len(t, result.Messages, 3)
}

func TestQueryStartTime(t *testing.T) {
	db, _ := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{
		StartTime: 2,
	})

	require.NoError(t, err)
	require.Len(t, result.Messages, 2)
}

func TestQueryTimeWindow(t *testing.T) {
	db, _ := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{
		StartTime: 2,
		EndTime:   3,
	})

	require.NoError(t, err)
	require.Len(t, result.Messages, 2)
	require.Equal(t, result.Messages[0].Timestamp, int64(3))
}

func TestQueryTimeAndTopic(t *testing.T) {
	db, _ := createAndFillDb(t)

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
	require.Len(t, resultNoHits.Messages, 0)

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
	db, _ := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{
		PubsubTopic: "test",
	})

	require.NoError(t, err)
	require.Len(t, result.Messages, 3)

	result, err = FindMessages(db, &pb.HistoryQuery{
		PubsubTopic: "foo",
	})
	require.NoError(t, err)
	require.Len(t, result.Messages, 0)
}

func TestDirectionSingleField(t *testing.T) {
	db, _ := createAndFillDb(t)

	result, err := FindMessages(db, &pb.HistoryQuery{
		PagingInfo: &pb.PagingInfo{
			Direction: pb.PagingInfo_FORWARD,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Messages, 3)
	require.Equal(t, result.Messages[0].Timestamp, int64(1))

	result, err = FindMessages(db, &pb.HistoryQuery{
		PagingInfo: &pb.PagingInfo{
			Direction: pb.PagingInfo_BACKWARD,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Messages, 3)
	require.Equal(t, result.Messages[0].Timestamp, int64(3))
}

func TestCursorTimestamp(t *testing.T) {
	db, msgs := createAndFillDb(t)
	idx := buildIndex(msgs[1], TOPIC)
	result, err := FindMessages(db, &pb.HistoryQuery{
		PagingInfo: &pb.PagingInfo{
			Cursor: &pb.Index{
				SenderTime:  int64(2),
				Digest:      idx.Digest,
				PubsubTopic: "test",
			},
			Direction: pb.PagingInfo_FORWARD,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Messages, 1)
	require.Equal(t, result.Messages[0].Timestamp, int64(3))
}

func TestCursorTimestampBackwards(t *testing.T) {
	db, _ := createAndFillDb(t)
	// Create a copy of the second message to use as the index
	msg2 := tests.CreateWakuMessage("test2", 2)
	idx := buildIndex(msg2, "test")

	result, err := FindMessages(db, &pb.HistoryQuery{
		PagingInfo: &pb.PagingInfo{
			Cursor:    idx,
			Direction: pb.PagingInfo_BACKWARD,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Messages, 1)
	require.Equal(t, result.Messages[0].Timestamp, int64(1))
}

func TestCursorTimestampTie(t *testing.T) {
	db, msgs := createAndFillDb(t)
	idx := buildIndex(msgs[0], TOPIC)
	result, err := FindMessages(db, &pb.HistoryQuery{
		PagingInfo: &pb.PagingInfo{
			Cursor: &pb.Index{
				SenderTime:  int64(1),
				Digest:      idx.Digest,
				PubsubTopic: "test",
			},
			Direction: pb.PagingInfo_FORWARD,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Messages, 2)
	require.Equal(t, result.Messages[0].Timestamp, int64(2))
}

func TestPagingInfoGeneration(t *testing.T) {
	db, _ := createAndFillDb(t)

	results := []*pb.WakuMessage{}
	var cursor *pb.Index
	for {
		response, err := FindMessages(db, &pb.HistoryQuery{
			PagingInfo: &pb.PagingInfo{
				PageSize:  1,
				Cursor:    cursor,
				Direction: pb.PagingInfo_FORWARD,
			},
		})
		require.NoError(t, err)
		require.Len(t, response.Messages, 1)
		results = append(results, response.Messages...)
		if len(results) == 3 {
			break
		}
		// Use this cursor to continue pagination
		cursor = response.PagingInfo.Cursor
	}
	require.Len(t, results, 3)
	require.Equal(t, results[0].Timestamp, int64(1))
	require.Equal(t, results[2].Timestamp, int64(3))
}

func TestPagingInfoWithFilter(t *testing.T) {
	db, _ := createAndFillDb(t)
	store := createStore(t, db)
	additionalMessage := tests.CreateWakuMessage("test1", 2)
	err := store.Put(protocol.NewEnvelope(additionalMessage, utils.GetUnixEpoch(), "test"))
	require.NoError(t, err)

	results := []*pb.WakuMessage{}
	var cursor *pb.Index
	for {
		response, err := FindMessages(db, &pb.HistoryQuery{
			ContentFilters: []*pb.ContentFilter{{
				ContentTopic: "test1",
			}},
			PagingInfo: &pb.PagingInfo{
				PageSize:  1,
				Cursor:    cursor,
				Direction: pb.PagingInfo_FORWARD,
			},
		})
		require.NoError(t, err)
		require.Len(t, response.Messages, 1)
		require.Equal(t, response.Messages[0].ContentTopic, "test1")
		results = append(results, response.Messages...)
		if len(results) == 2 {
			break
		}
		// Use this cursor to continue pagination
		cursor = response.PagingInfo.Cursor
	}
	require.Len(t, results, 2)
	require.Equal(t, results[0].Timestamp, int64(1))
	require.Equal(t, results[1].Timestamp, int64(2))
}

func TestLastPage(t *testing.T) {
	db, _ := createAndFillDb(t)
	msg2 := tests.CreateWakuMessage("test2", 2)
	idx := buildIndex(msg2, "test")

	response, err := FindMessages(db, &pb.HistoryQuery{
		PagingInfo: &pb.PagingInfo{
			PageSize:  10,
			Cursor:    idx,
			Direction: pb.PagingInfo_FORWARD,
		},
	})
	require.NoError(t, err)
	require.Len(t, response.Messages, 1)
	require.Equal(t, response.PagingInfo.PageSize, uint64(0))
	require.Nil(t, response.PagingInfo.Cursor)
}

func TestPageSizeOne(t *testing.T) {
	db, _ := createAndFillDb(t)
	var cursor *pb.Index
	loops := 0
	for {
		response, err := FindMessages(db, &pb.HistoryQuery{
			ContentFilters: []*pb.ContentFilter{
				{
					ContentTopic: "test1",
				},
			},
			PagingInfo: &pb.PagingInfo{
				PageSize:  1,
				Cursor:    cursor,
				Direction: pb.PagingInfo_FORWARD,
			},
		})
		cursor = response.PagingInfo.Cursor
		require.NoError(t, err)
		if loops == 1 {
			require.Len(t, response.Messages, 0)
			require.Equal(t, response.PagingInfo.PageSize, uint64(0))
			break
		}
		loops++
	}
}
