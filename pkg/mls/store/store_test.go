package store

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	mlsv1 "github.com/xmtp/proto/v3/go/mls/api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func NewTestStore(t *testing.T) (*Store, func()) {
	log := test.NewLog(t)
	db, _, dbCleanup := test.NewMLSDB(t)
	ctx := context.Background()
	c := Config{
		Log: log,
		DB:  db,
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

func TestInsertGroupMessage_Single(t *testing.T) {
	started := time.Now().UTC()
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertGroupMessage(ctx, "installation", []byte("data"))
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, uint64(1), msg.Id)
	require.True(t, msg.CreatedAt.Before(time.Now().UTC()) && msg.CreatedAt.After(started))
	require.Equal(t, "installation", msg.GroupId)
	require.Equal(t, []byte("data"), msg.Data)

	msgs := make([]*GroupMessage, 0)
	err = store.db.NewSelect().Model(&msgs).Scan(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, msg, msgs[0])
}

func TestInsertGroupMessage_ManyOrderedByTime(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertGroupMessage(ctx, "group", []byte("data1"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group", []byte("data2"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group", []byte("data3"))
	require.NoError(t, err)

	msgs := make([]*GroupMessage, 0)
	err = store.db.NewSelect().Model(&msgs).Order("created_at DESC").Scan(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 3)
	require.Equal(t, []byte("data3"), msgs[0].Data)
	require.Equal(t, []byte("data2"), msgs[1].Data)
	require.Equal(t, []byte("data1"), msgs[2].Data)
}

func TestInsertWelcomeMessage_Single(t *testing.T) {
	started := time.Now().UTC()
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertWelcomeMessage(ctx, "group", []byte("data"))
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, uint64(1), msg.Id)
	require.True(t, msg.CreatedAt.Before(time.Now().UTC()) && msg.CreatedAt.After(started))
	require.Equal(t, "group", msg.InstallationId)
	require.Equal(t, []byte("data"), msg.Data)

	msgs := make([]*WelcomeMessage, 0)
	err = store.db.NewSelect().Model(&msgs).Scan(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, msg, msgs[0])
}

func TestInsertWelcomeMessage_ManyOrderedByTime(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertWelcomeMessage(ctx, "installation", []byte("data1"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation", []byte("data2"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation", []byte("data3"))
	require.NoError(t, err)

	msgs := make([]*WelcomeMessage, 0)
	err = store.db.NewSelect().Model(&msgs).Order("created_at DESC").Scan(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 3)
	require.Equal(t, []byte("data3"), msgs[0].Data)
	require.Equal(t, []byte("data2"), msgs[1].Data)
	require.Equal(t, []byte("data1"), msgs[2].Data)
}

func TestQueryGroupMessagesV1_MissingGroup(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()

	resp, err := store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{})
	require.EqualError(t, err, "group is required")
	require.Nil(t, resp)
}

func TestQueryWelcomeMessagesV1_MissingInstallation(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()

	resp, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{})
	require.EqualError(t, err, "installation is required")
	require.Nil(t, resp)
}

func TestQueryGroupMessagesV1_Filter(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertGroupMessage(ctx, "group1", []byte("data1"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group2", []byte("data2"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group1", []byte("data3"))
	require.NoError(t, err)

	resp, err := store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: "unknown",
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 0)

	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: "group1",
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("data3"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("data1"), resp.Messages[1].GetV1().Data)

	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: "group2",
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, []byte("data2"), resp.Messages[0].GetV1().Data)

	// Sort ascending
	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: "group1",
		PagingInfo: &mlsv1.PagingInfo{
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("data1"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("data3"), resp.Messages[1].GetV1().Data)
}

func TestQueryWelcomeMessagesV1_Filter(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertWelcomeMessage(ctx, "installation1", []byte("data1"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation2", []byte("data2"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation1", []byte("data3"))
	require.NoError(t, err)

	resp, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationId: "unknown",
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 0)

	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationId: "installation1",
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("data3"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("data1"), resp.Messages[1].GetV1().Data)

	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationId: "installation2",
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, []byte("data2"), resp.Messages[0].GetV1().Data)

	// Sort ascending
	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationId: "installation1",
		PagingInfo: &mlsv1.PagingInfo{
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("data1"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("data3"), resp.Messages[1].GetV1().Data)
}

func TestQueryGroupMessagesV1_Paginate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertGroupMessage(ctx, "group1", []byte("content1"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group2", []byte("content2"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group1", []byte("content3"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group2", []byte("content4"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group1", []byte("content5"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group1", []byte("content6"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group1", []byte("content7"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, "group1", []byte("content8"))
	require.NoError(t, err)

	resp, err := store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: "group1",
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 6)
	require.Equal(t, []byte("content8"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content7"), resp.Messages[1].GetV1().Data)
	require.Equal(t, []byte("content6"), resp.Messages[2].GetV1().Data)
	require.Equal(t, []byte("content5"), resp.Messages[3].GetV1().Data)
	require.Equal(t, []byte("content3"), resp.Messages[4].GetV1().Data)
	require.Equal(t, []byte("content1"), resp.Messages[5].GetV1().Data)

	thirdMsg := resp.Messages[2]
	fifthMsg := resp.Messages[4]

	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: "group1",
		PagingInfo: &mlsv1.PagingInfo{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content8"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content7"), resp.Messages[1].GetV1().Data)

	// Order descending by default
	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: "group1",
		PagingInfo: &mlsv1.PagingInfo{
			Limit:  2,
			Cursor: thirdMsg.GetV1().Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content5"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content3"), resp.Messages[1].GetV1().Data)

	// Next page from previous response
	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId:    "group1",
		PagingInfo: resp.PagingInfo,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, []byte("content1"), resp.Messages[0].GetV1().Data)

	// Order ascending
	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: "group1",
		PagingInfo: &mlsv1.PagingInfo{
			Limit:     2,
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
			Cursor:    fifthMsg.GetV1().Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content5"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content6"), resp.Messages[1].GetV1().Data)

	// Next page from previous response
	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId:    "group1",
		PagingInfo: resp.PagingInfo,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content7"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content8"), resp.Messages[1].GetV1().Data)
}

func TestQueryWelcomeMessagesV1_Paginate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertWelcomeMessage(ctx, "installation1", []byte("content1"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation2", []byte("content2"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation1", []byte("content3"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation2", []byte("content4"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation1", []byte("content5"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation1", []byte("content6"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation1", []byte("content7"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, "installation1", []byte("content8"))
	require.NoError(t, err)

	resp, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationId: "installation1",
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 6)
	require.Equal(t, []byte("content8"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content7"), resp.Messages[1].GetV1().Data)
	require.Equal(t, []byte("content6"), resp.Messages[2].GetV1().Data)
	require.Equal(t, []byte("content5"), resp.Messages[3].GetV1().Data)
	require.Equal(t, []byte("content3"), resp.Messages[4].GetV1().Data)
	require.Equal(t, []byte("content1"), resp.Messages[5].GetV1().Data)

	thirdMsg := resp.Messages[2]
	fifthMsg := resp.Messages[4]

	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationId: "installation1",
		PagingInfo: &mlsv1.PagingInfo{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content8"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content7"), resp.Messages[1].GetV1().Data)

	// Order descending by default
	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationId: "installation1",
		PagingInfo: &mlsv1.PagingInfo{
			Limit:  2,
			Cursor: thirdMsg.GetV1().Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content5"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content3"), resp.Messages[1].GetV1().Data)

	// Next page from previous response
	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationId: "installation1",
		PagingInfo:     resp.PagingInfo,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, []byte("content1"), resp.Messages[0].GetV1().Data)

	// Order ascending
	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationId: "installation1",
		PagingInfo: &mlsv1.PagingInfo{
			Limit:     2,
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
			Cursor:    fifthMsg.GetV1().Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content5"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content6"), resp.Messages[1].GetV1().Data)

	// Next page from previous response
	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationId: "installation1",
		PagingInfo:     resp.PagingInfo,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content7"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content8"), resp.Messages[1].GetV1().Data)
}
