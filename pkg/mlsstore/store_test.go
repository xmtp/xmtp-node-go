package mlsstore

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

type testStoreOption func(*Config)

func withFixedNow(now time.Time) testStoreOption {
	return func(c *Config) {
		c.now = func() time.Time {
			return now
		}
	}
}

func NewTestStore(t *testing.T, opts ...testStoreOption) (*Store, func()) {
	log := test.NewLog(t)
	db, _, dbCleanup := test.NewMLSDB(t)
	ctx := context.Background()
	c := Config{
		Log: log,
		DB:  db,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	for _, opt := range opts {
		opt(&c)
	}

	store, err := New(ctx, c)
	require.NoError(t, err)

	return store, dbCleanup
}

func TestCreateInstallation(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := test.RandomBytes(32)
	walletAddress := test.RandomString(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, test.RandomBytes(32), test.RandomBytes(32), 0)
	require.NoError(t, err)

	installationFromDb := &Installation{}
	require.NoError(t, store.db.NewSelect().Model(installationFromDb).Where("id = ?", installationId).Scan(ctx))
	require.Equal(t, walletAddress, installationFromDb.WalletAddress)
}

func TestUpdateKeyPackage(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := test.RandomBytes(32)
	walletAddress := test.RandomString(32)
	keyPackage := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, keyPackage, keyPackage, 0)
	require.NoError(t, err)

	keyPackage2 := test.RandomBytes(32)
	err = store.UpdateKeyPackage(ctx, installationId, keyPackage2, 1)
	require.NoError(t, err)

	installationFromDb := &Installation{}
	require.NoError(t, store.db.NewSelect().Model(installationFromDb).Where("id = ?", installationId).Scan(ctx))

	require.Equal(t, keyPackage2, installationFromDb.KeyPackage)
	require.Equal(t, uint64(1), installationFromDb.Expiration)
}

func TestConsumeLastResortKeyPackage(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := test.RandomBytes(32)
	walletAddress := test.RandomString(32)
	keyPackage := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, keyPackage, keyPackage, 0)
	require.NoError(t, err)

	fetchResult, err := store.FetchKeyPackages(ctx, [][]byte{installationId})
	require.NoError(t, err)
	require.Len(t, fetchResult, 1)
	require.Equal(t, keyPackage, fetchResult[0].KeyPackage)
	require.Equal(t, installationId, fetchResult[0].ID)
}

func TestGetIdentityUpdates(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	walletAddress := test.RandomString(32)

	installationId1 := test.RandomBytes(32)
	keyPackage1 := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId1, walletAddress, keyPackage1, keyPackage1, 0)
	require.NoError(t, err)

	installationId2 := test.RandomBytes(32)
	keyPackage2 := test.RandomBytes(32)

	err = store.CreateInstallation(ctx, installationId2, walletAddress, keyPackage2, keyPackage2, 0)
	require.NoError(t, err)

	identityUpdates, err := store.GetIdentityUpdates(ctx, []string{walletAddress}, 0)
	require.NoError(t, err)
	require.Len(t, identityUpdates[walletAddress], 2)
	require.Equal(t, identityUpdates[walletAddress][0].InstallationId, installationId1)
	require.Equal(t, identityUpdates[walletAddress][0].Kind, Create)
	require.Equal(t, identityUpdates[walletAddress][1].InstallationId, installationId2)

	// Make sure that date filtering works
	identityUpdates, err = store.GetIdentityUpdates(ctx, []string{walletAddress}, nowNs()+1000000)
	require.NoError(t, err)
	require.Len(t, identityUpdates[walletAddress], 0)
}

func TestGetIdentityUpdatesMultipleWallets(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	walletAddress1 := test.RandomString(32)
	installationId1 := test.RandomBytes(32)
	keyPackage1 := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId1, walletAddress1, keyPackage1, keyPackage1, 0)
	require.NoError(t, err)

	walletAddress2 := test.RandomString(32)
	installationId2 := test.RandomBytes(32)
	keyPackage2 := test.RandomBytes(32)

	err = store.CreateInstallation(ctx, installationId2, walletAddress2, keyPackage2, keyPackage2, 0)
	require.NoError(t, err)

	identityUpdates, err := store.GetIdentityUpdates(ctx, []string{walletAddress1, walletAddress2}, 0)
	require.NoError(t, err)
	require.Len(t, identityUpdates[walletAddress1], 1)
	require.Len(t, identityUpdates[walletAddress2], 1)
}

func TestGetIdentityUpdatesNoResult(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	walletAddress := test.RandomString(32)

	identityUpdates, err := store.GetIdentityUpdates(ctx, []string{walletAddress}, 0)
	require.NoError(t, err)
	require.Len(t, identityUpdates[walletAddress], 0)
}

func TestIdentityUpdateSort(t *testing.T) {
	updates := IdentityUpdateList([]IdentityUpdate{
		{
			Kind:        Create,
			TimestampNs: 2,
		},
		{
			Kind:        Create,
			TimestampNs: 3,
		},
		{
			Kind:        Create,
			TimestampNs: 1,
		},
	})
	sort.Sort(updates)
	require.Equal(t, updates[0].TimestampNs, uint64(1))
	require.Equal(t, updates[1].TimestampNs, uint64(2))
	require.Equal(t, updates[2].TimestampNs, uint64(3))
}

func TestInsertMessage_Single(t *testing.T) {
	now := time.Now().UTC()
	store, cleanup := NewTestStore(t, withFixedNow(now))
	defer cleanup()

	ctx := context.Background()
	env, err := store.InsertMessage(ctx, "topic", []byte("content"))
	require.NoError(t, err)
	require.NotNil(t, env)
	require.Equal(t, &messagev1.Envelope{
		ContentTopic: "topic",
		TimestampNs:  uint64(now.UnixNano()),
		Message:      []byte("content"),
	}, env)

	msgs := make([]*Message, 0)
	err = store.db.NewSelect().Model(&msgs).Scan(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, &Message{
		Topic:     "topic",
		CreatedAt: now.UnixNano(),
		Content:   []byte("content"),
		TID:       buildMessageTID(now, []byte("content")),
	}, msgs[0])
}

func TestInsertMessage_Duplicate(t *testing.T) {
	now := time.Now().UTC()
	store, cleanup := NewTestStore(t, withFixedNow(now))
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertMessage(ctx, "topic", []byte("content"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic", []byte("content"))
	require.NoError(t, err)

	msgs := make([]*Message, 0)
	err = store.db.NewSelect().Model(&msgs).Scan(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, &Message{
		Topic:     "topic",
		CreatedAt: now.UnixNano(),
		Content:   []byte("content"),
		TID:       buildMessageTID(now, []byte("content")),
	}, msgs[0])
}

func TestInsertMessage_ManyAreOrderedByTime(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertMessage(ctx, "topic", []byte("content1"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic", []byte("content2"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic", []byte("content3"))
	require.NoError(t, err)

	msgs := make([]*Message, 0)
	err = store.db.NewSelect().Model(&msgs).Order("tid DESC").Scan(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 3)
	require.Equal(t, []byte("content3"), msgs[0].Content)
	require.Equal(t, []byte("content2"), msgs[1].Content)
	require.Equal(t, []byte("content1"), msgs[2].Content)
}

func TestQueryMessages_MissingTopic(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()

	resp, err := store.QueryMessages(ctx, &messagev1.QueryRequest{})
	require.EqualError(t, err, "topic is required")
	require.Nil(t, resp)
}

func TestQueryMessages_Filter(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertMessage(ctx, "topic1", []byte("content1"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic2", []byte("content2"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content3"))
	require.NoError(t, err)

	resp, err := store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"unknown"},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 0)

	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 2)
	require.Equal(t, []byte("content3"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content1"), resp.Envelopes[1].Message)

	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic2"},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 1)
	require.Equal(t, []byte("content2"), resp.Envelopes[0].Message)

	// Sort ascending
	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		PagingInfo: &messagev1.PagingInfo{
			Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 2)
	require.Equal(t, []byte("content1"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content3"), resp.Envelopes[1].Message)
}

func TestQueryMessages_Paginate_Cursor(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertMessage(ctx, "topic1", []byte("content1"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic2", []byte("content2"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content3"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic2", []byte("content4"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content5"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content6"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content7"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content8"))
	require.NoError(t, err)

	resp, err := store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 6)
	require.Equal(t, []byte("content8"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content7"), resp.Envelopes[1].Message)
	require.Equal(t, []byte("content6"), resp.Envelopes[2].Message)
	require.Equal(t, []byte("content5"), resp.Envelopes[3].Message)
	require.Equal(t, []byte("content3"), resp.Envelopes[4].Message)
	require.Equal(t, []byte("content1"), resp.Envelopes[5].Message)

	thirdEnv := resp.Envelopes[2]
	fifthEnv := resp.Envelopes[4]

	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		PagingInfo: &messagev1.PagingInfo{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 2)
	require.Equal(t, []byte("content8"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content7"), resp.Envelopes[1].Message)

	// Order descending by default
	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		PagingInfo: &messagev1.PagingInfo{
			Limit: 2,
			Cursor: &messagev1.Cursor{
				Cursor: &messagev1.Cursor_Index{
					Index: &messagev1.IndexCursor{
						SenderTimeNs: thirdEnv.TimestampNs,
						Digest:       buildMessageContentDigest(thirdEnv.Message),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 2)
	require.Equal(t, []byte("content5"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content3"), resp.Envelopes[1].Message)

	// Next page from previous response
	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		PagingInfo:    resp.PagingInfo,
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 1)
	require.Equal(t, []byte("content1"), resp.Envelopes[0].Message)

	// Order ascending
	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		PagingInfo: &messagev1.PagingInfo{
			Limit:     2,
			Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
			Cursor: &messagev1.Cursor{
				Cursor: &messagev1.Cursor_Index{
					Index: &messagev1.IndexCursor{
						SenderTimeNs: fifthEnv.TimestampNs,
						Digest:       buildMessageContentDigest(fifthEnv.Message),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 2)
	require.Equal(t, []byte("content5"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content6"), resp.Envelopes[1].Message)

	// Next page from previous response
	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		PagingInfo:    resp.PagingInfo,
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 2)
	require.Equal(t, []byte("content7"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content8"), resp.Envelopes[1].Message)
}

func TestQueryMessages_Paginate_StartEndTime(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertMessage(ctx, "topic1", []byte("content1"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic2", []byte("content2"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content3"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic2", []byte("content4"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content5"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content6"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content7"))
	require.NoError(t, err)
	_, err = store.InsertMessage(ctx, "topic1", []byte("content8"))
	require.NoError(t, err)

	resp, err := store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 6)
	require.Equal(t, []byte("content8"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content7"), resp.Envelopes[1].Message)
	require.Equal(t, []byte("content6"), resp.Envelopes[2].Message)
	require.Equal(t, []byte("content5"), resp.Envelopes[3].Message)
	require.Equal(t, []byte("content3"), resp.Envelopes[4].Message)
	require.Equal(t, []byte("content1"), resp.Envelopes[5].Message)

	thirdEnv := resp.Envelopes[2]
	fifthEnv := resp.Envelopes[4]

	// Order descending by default
	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		StartTimeNs:   thirdEnv.TimestampNs,
		PagingInfo: &messagev1.PagingInfo{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 2)
	require.Equal(t, []byte("content8"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content7"), resp.Envelopes[1].Message)

	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		EndTimeNs:     thirdEnv.TimestampNs,
		PagingInfo: &messagev1.PagingInfo{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 2)
	require.Equal(t, []byte("content6"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content5"), resp.Envelopes[1].Message)

	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		StartTimeNs:   fifthEnv.TimestampNs,
		EndTimeNs:     thirdEnv.TimestampNs,
		PagingInfo: &messagev1.PagingInfo{
			Limit: 4,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 3)
	require.Equal(t, []byte("content6"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content5"), resp.Envelopes[1].Message)
	require.Equal(t, []byte("content3"), resp.Envelopes[2].Message)

	// Order ascending
	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		StartTimeNs:   thirdEnv.TimestampNs,
		PagingInfo: &messagev1.PagingInfo{
			Limit:     2,
			Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 2)
	require.Equal(t, []byte("content6"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content7"), resp.Envelopes[1].Message)

	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		EndTimeNs:     thirdEnv.TimestampNs,
		PagingInfo: &messagev1.PagingInfo{
			Limit:     2,
			Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 2)
	require.Equal(t, []byte("content1"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content3"), resp.Envelopes[1].Message)

	resp, err = store.QueryMessages(ctx, &messagev1.QueryRequest{
		ContentTopics: []string{"topic1"},
		StartTimeNs:   fifthEnv.TimestampNs,
		EndTimeNs:     thirdEnv.TimestampNs,
		PagingInfo: &messagev1.PagingInfo{
			Limit:     4,
			Direction: messagev1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Envelopes, 3)
	require.Equal(t, []byte("content3"), resp.Envelopes[0].Message)
	require.Equal(t, []byte("content5"), resp.Envelopes[1].Message)
	require.Equal(t, []byte("content6"), resp.Envelopes[2].Message)
}

func nowNs() int64 {
	return time.Now().UTC().UnixNano()
}
