package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Blank import to register the sqlite3 driver

	"github.com/status-im/go-waku/tests"
	"github.com/status-im/go-waku/waku/persistence"
	"github.com/status-im/go-waku/waku/v2/protocol"
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

func buildIndex(msg *pb.WakuMessage, topic string) *pb.Index {
	idx, _ := computeIndex(protocol.NewEnvelope(msg, topic))
	return idx
}

func createStore(t *testing.T, db *sql.DB) *persistence.DBStore {
	option := persistence.WithDB(db)
	store, err := persistence.NewDBStore(tests.Logger(), option)
	require.NoError(t, err)
	return store
}

func createAndFillDb(t *testing.T) *sql.DB {
	db := NewMock()
	store := createStore(t, db)
	res, err := store.GetAll()
	require.NoError(t, err)
	require.Empty(t, res)

	msg1 := tests.CreateWakuMessage("test1", 1)
	msg2 := tests.CreateWakuMessage("test2", 2)
	msg3 := tests.CreateWakuMessage("test3", 3)
	topic := "test"

	err = store.Put(
		buildIndex(msg1, topic),
		topic,
		msg1,
	)
	require.NoError(t, err)

	err = store.Put(
		buildIndex(msg2, topic),
		topic,
		msg2,
	)
	require.NoError(t, err)

	err = store.Put(
		buildIndex(msg3, topic),
		topic,
		msg3,
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
	db := createAndFillDb(t)

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
	db := createAndFillDb(t)

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
	db := createAndFillDb(t)
	result, err := FindMessages(db, &pb.HistoryQuery{
		PagingInfo: &pb.PagingInfo{
			Cursor: &pb.Index{
				SenderTime:  int64(2),
				Digest:      []byte("digest2"),
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
	db := createAndFillDb(t)
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
	db := createAndFillDb(t)
	result, err := FindMessages(db, &pb.HistoryQuery{
		PagingInfo: &pb.PagingInfo{
			Cursor: &pb.Index{
				SenderTime:  int64(1),
				Digest:      []byte("digest2"),
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
	db := createAndFillDb(t)

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
	db := createAndFillDb(t)
	store := createStore(t, db)
	additionalMessage := tests.CreateWakuMessage("test1", 2)
	err := store.Put(
		buildIndex(additionalMessage, "test"),
		"test",
		additionalMessage,
	)
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
	db := createAndFillDb(t)
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
	// Not sure why the existing behaviour includes the previous cursor on the last page.
	// Would be more sensible to just return a null cursor. But I'm replicating the behaviour anyways
	require.Equal(t, response.PagingInfo.Cursor.SenderTime, idx.SenderTime)
}
