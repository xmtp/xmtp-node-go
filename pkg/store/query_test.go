package store

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Blank import to register the sqlite3 driver
	"github.com/stretchr/testify/require"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

const TOPIC = "test"

func createAndFillDb(t *testing.T) (*Store, func(), []*messagev1.Envelope) {
	store, cleanup := newTestStore(t)

	ts := time.Now().UTC().UnixNano()
	msg1 := test.NewEnvelope("test1", 1, "")
	msg2 := test.NewEnvelope("test2", 2, "")
	msg3 := test.NewEnvelope("test3", 3, "")

	err := store.insertMessage(msg1, ts)
	require.NoError(t, err)

	err = store.insertMessage(msg2, ts)
	require.NoError(t, err)

	err = store.insertMessage(msg3, ts)
	require.NoError(t, err)

	return store, cleanup, []*messagev1.Envelope{msg1, msg2, msg3}
}

func TestQuerySimple(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()

	result, err := store.Query(&messagev1.QueryRequest{
		ContentTopics: []string{"test2"},
	})

	require.NoError(t, err)
	require.Len(t, result.Envelopes, 1)
	require.Equal(t, result.Envelopes[0].ContentTopic, "test2")
	require.Equal(t, result.Envelopes[0].Message, []byte{})
}

func TestQueryMultipleTopics(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()

	result, err := store.Query(&messagev1.QueryRequest{
		ContentTopics: []string{"test2", "test3"},
	})

	require.NoError(t, err)
	require.Len(t, result.Envelopes, 2)
}

func TestQueryNoTopics(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()

	result, err := store.Query(&messagev1.QueryRequest{})

	require.NoError(t, err)
	require.Len(t, result.Envelopes, 3)
}

func TestQueryStartTime(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()

	result, err := store.Query(&messagev1.QueryRequest{
		StartTimeNs: 2,
	})

	require.NoError(t, err)
	require.Len(t, result.Envelopes, 2)
}

func TestQueryTimeWindow(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()

	result, err := store.Query(&messagev1.QueryRequest{
		StartTimeNs: 2,
		EndTimeNs:   3,
	})

	require.NoError(t, err)
	require.Len(t, result.Envelopes, 2)
	require.Equal(t, result.Envelopes[0].TimestampNs, uint64(3))
}

func TestQueryTimeAndTopic(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()

	resultNoHits, err := store.Query(&messagev1.QueryRequest{
		ContentTopics: []string{"test2"},
		StartTimeNs:   3,
		EndTimeNs:     4,
	})

	require.NoError(t, err)
	require.Len(t, resultNoHits.Envelopes, 0)

	resultWithHits, err := store.Query(&messagev1.QueryRequest{
		ContentTopics: []string{"test3"},
		StartTimeNs:   3,
		EndTimeNs:     4,
	})
	require.NoError(t, err)
	require.Len(t, resultWithHits.Envelopes, 1)
}

func TestDirectionSingleField(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()

	result, err := store.Query(&messagev1.QueryRequest{
		PagingInfo: &messagev1.PagingInfo{
			Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Envelopes, 3)
	require.Equal(t, result.Envelopes[0].TimestampNs, uint64(1))

	result, err = store.Query(&messagev1.QueryRequest{
		PagingInfo: &messagev1.PagingInfo{
			Direction: messagev1.SortDirection_SORT_DIRECTION_DESCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Envelopes, 3)
	require.Equal(t, result.Envelopes[0].TimestampNs, uint64(3))
}

func TestCursorTimestamp(t *testing.T) {
	store, cleanup, msgs := createAndFillDb(t)
	defer cleanup()
	result, err := store.Query(&messagev1.QueryRequest{
		PagingInfo: &messagev1.PagingInfo{
			Cursor:    buildCursor(msgs[1]),
			Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Envelopes, 1)
	require.Equal(t, result.Envelopes[0].TimestampNs, uint64(3))
}

func TestCursorTimestampBackwards(t *testing.T) {
	store, cleanup, msgs := createAndFillDb(t)
	defer cleanup()

	result, err := store.Query(&messagev1.QueryRequest{
		PagingInfo: &messagev1.PagingInfo{
			Cursor:    buildCursor(msgs[1]),
			Direction: messagev1.SortDirection_SORT_DIRECTION_DESCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Envelopes, 1)
	require.Equal(t, result.Envelopes[0].TimestampNs, uint64(1))
}

func TestCursorTimestampTie(t *testing.T) {
	store, cleanup, msgs := createAndFillDb(t)
	defer cleanup()
	result, err := store.Query(&messagev1.QueryRequest{
		PagingInfo: &messagev1.PagingInfo{
			Cursor:    buildCursor(msgs[0]),
			Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Envelopes, 2)
	require.Equal(t, result.Envelopes[0].TimestampNs, uint64(2))
}

func TestPagingInfoGeneration(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()

	results := []*messagev1.Envelope{}
	var cursor *messagev1.Cursor
	for {
		response, err := store.Query(&messagev1.QueryRequest{
			PagingInfo: &messagev1.PagingInfo{
				Limit:     1,
				Cursor:    cursor,
				Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
			},
		})
		require.NoError(t, err)
		require.Len(t, response.Envelopes, 1)
		results = append(results, response.Envelopes...)
		if len(results) == 3 {
			break
		}
		// Use this cursor to continue pagination
		cursor = response.PagingInfo.Cursor
	}
	require.Len(t, results, 3)
	require.Equal(t, results[0].TimestampNs, uint64(1))
	require.Equal(t, results[2].TimestampNs, uint64(3))
}

func TestPagingInfoWithFilter(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()
	additionalEnv := test.NewEnvelope("test1", 2, "")
	err := store.insertMessage(additionalEnv, time.Now().UTC().UnixNano())
	require.NoError(t, err)

	results := []*messagev1.Envelope{}
	var cursor *messagev1.Cursor
	for {
		response, err := store.Query(&messagev1.QueryRequest{
			ContentTopics: []string{"test1"},
			PagingInfo: &messagev1.PagingInfo{
				Limit:     1,
				Cursor:    cursor,
				Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
			},
		})
		require.NoError(t, err)
		require.Len(t, response.Envelopes, 1)
		require.Equal(t, response.Envelopes[0].ContentTopic, "test1")
		results = append(results, response.Envelopes...)
		if len(results) == 2 {
			break
		}
		// Use this cursor to continue pagination
		cursor = response.PagingInfo.Cursor
	}
	require.Len(t, results, 2)
	require.Equal(t, results[0].TimestampNs, uint64(1))
	require.Equal(t, results[1].TimestampNs, uint64(2))
}

func TestLastPage(t *testing.T) {
	store, cleanup, msgs := createAndFillDb(t)
	defer cleanup()

	response, err := store.Query(&messagev1.QueryRequest{
		PagingInfo: &messagev1.PagingInfo{
			Limit:     10,
			Cursor:    buildCursor(msgs[1]),
			Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, response.Envelopes, 1)
	require.Equal(t, response.PagingInfo.Limit, uint32(0))
	require.Nil(t, response.PagingInfo.Cursor)
}

func TestPageSizeOne(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()
	var cursor *messagev1.Cursor
	loops := 0
	for {
		response, err := store.Query(&messagev1.QueryRequest{
			ContentTopics: []string{"test1"},
			PagingInfo: &messagev1.PagingInfo{
				Limit:     1,
				Cursor:    cursor,
				Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
			},
		})
		cursor = response.PagingInfo.Cursor
		require.NoError(t, err)
		if loops == 1 {
			require.Len(t, response.Envelopes, 0)
			require.Equal(t, response.PagingInfo.Limit, uint32(0))
			break
		}
		loops++
	}
}

func TestMlsMessagePublish(t *testing.T) {
	store, cleanup, _ := createAndFillDb(t)
	defer cleanup()

	message := []byte{1, 2, 3}
	contentTopic := "foo"
	ctx := context.Background()

	env, err := store.InsertMlsMessage(ctx, contentTopic, message)
	require.NoError(t, err)

	require.Equal(t, env.ContentTopic, contentTopic)
	require.Equal(t, env.Message, message)

	response, err := store.Query(&messagev1.QueryRequest{
		ContentTopics: []string{contentTopic},
	})
	require.Len(t, response.Envelopes, 1)
	require.Equal(t, response.Envelopes[0].Message, message)
	require.Equal(t, response.Envelopes[0].ContentTopic, contentTopic)
	require.NotNil(t, response.Envelopes[0].TimestampNs)

	parsedTime := time.Unix(0, int64(response.Envelopes[0].TimestampNs))
	// Sanity check to ensure that the timestamps are reasonable
	require.True(t, time.Since(parsedTime) < 10*time.Second || time.Since(parsedTime) > -10*time.Second)

	require.Equal(t, env.TimestampNs, response.Envelopes[0].TimestampNs)
}
