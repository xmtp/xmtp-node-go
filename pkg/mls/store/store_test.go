package store

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
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
	require.Equal(t, identityUpdates[walletAddress][0].InstallationKey, installationId1)
	require.Equal(t, identityUpdates[walletAddress][0].Kind, Create)
	require.Equal(t, identityUpdates[walletAddress][1].InstallationKey, installationId2)

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
	started := time.Now().UTC().Add(-time.Minute)
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertGroupMessage(ctx, []byte("group"), []byte("data"))
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, uint64(1), msg.Id)
	require.True(t, msg.CreatedAt.Before(time.Now().UTC()) && msg.CreatedAt.After(started))
	require.Equal(t, []byte("group"), msg.GroupId)
	require.Equal(t, []byte("data"), msg.Data)

	msgs := make([]*GroupMessage, 0)
	err = store.db.NewSelect().Model(&msgs).Scan(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, msg, msgs[0])
}

func TestInsertGroupMessage_Duplicate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertGroupMessage(ctx, []byte("group"), []byte("data"))
	require.NoError(t, err)
	require.NotNil(t, msg)

	msg, err = store.InsertGroupMessage(ctx, []byte("group"), []byte("data"))
	require.Nil(t, msg)
	require.IsType(t, &AlreadyExistsError{}, err)
	require.True(t, IsAlreadyExistsError(err))
}

func TestInsertGroupMessage_ManyOrderedByTime(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertGroupMessage(ctx, []byte("group"), []byte("data1"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group"), []byte("data2"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group"), []byte("data3"))
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
	started := time.Now().UTC().Add(-time.Minute)
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data"))
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, uint64(1), msg.Id)
	require.True(t, msg.CreatedAt.Before(time.Now().UTC()) && msg.CreatedAt.After(started))
	require.Equal(t, []byte("installation"), msg.InstallationKey)
	require.Equal(t, []byte("data"), msg.Data)

	msgs := make([]*WelcomeMessage, 0)
	err = store.db.NewSelect().Model(&msgs).Scan(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, msg, msgs[0])
}

func TestInsertWelcomeMessage_Duplicate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data"))
	require.NoError(t, err)
	require.NotNil(t, msg)

	msg, err = store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data"))
	require.Nil(t, msg)
	require.IsType(t, &AlreadyExistsError{}, err)
	require.True(t, IsAlreadyExistsError(err))
}

func TestInsertWelcomeMessage_ManyOrderedByTime(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data1"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data2"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data3"))
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

	msgs, err := store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{})
	require.EqualError(t, err, "group is required")
	require.Nil(t, msgs)
}

func TestQueryWelcomeMessagesV1_MissingInstallation(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()

	msgs, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{})
	require.EqualError(t, err, "installation is required")
	require.Nil(t, msgs)
}

func TestQueryGroupMessagesV1_Filter(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertGroupMessage(ctx, []byte("group1"), []byte("data1"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group2"), []byte("data2"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("data3"))
	require.NoError(t, err)

	msgs, err := store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("unknown"),
	})
	require.NoError(t, err)
	require.Len(t, msgs, 0)

	msgs, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("data3"), msgs[0].Data)
	require.Equal(t, []byte("data1"), msgs[1].Data)

	msgs, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group2"),
	})
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, []byte("data2"), msgs[0].Data)

	// Sort ascending
	msgs, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
		PagingInfo: &mlsv1.PagingInfo{
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("data1"), msgs[0].Data)
	require.Equal(t, []byte("data3"), msgs[1].Data)
}

func TestQueryWelcomeMessagesV1_Filter(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("data1"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation2"), []byte("data2"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("data3"))
	require.NoError(t, err)

	msgs, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("unknown"),
	})
	require.NoError(t, err)
	require.Len(t, msgs, 0)

	msgs, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("data3"), msgs[0].Data)
	require.Equal(t, []byte("data1"), msgs[1].Data)

	msgs, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation2"),
	})
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, []byte("data2"), msgs[0].Data)

	// Sort ascending
	msgs, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo: &mlsv1.PagingInfo{
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("data1"), msgs[0].Data)
	require.Equal(t, []byte("data3"), msgs[1].Data)
}

func TestQueryGroupMessagesV1_Paginate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertGroupMessage(ctx, []byte("group1"), []byte("content1"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group2"), []byte("content2"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("content3"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group2"), []byte("content4"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("content5"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("content6"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("content7"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("content8"))
	require.NoError(t, err)

	msgs, err := store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
	})
	require.NoError(t, err)
	require.Len(t, msgs, 6)
	require.Equal(t, []byte("content8"), msgs[0].Data)
	require.Equal(t, []byte("content7"), msgs[1].Data)
	require.Equal(t, []byte("content6"), msgs[2].Data)
	require.Equal(t, []byte("content5"), msgs[3].Data)
	require.Equal(t, []byte("content3"), msgs[4].Data)
	require.Equal(t, []byte("content1"), msgs[5].Data)

	thirdMsg := msgs[2]
	fifthMsg := msgs[4]

	msgs, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("content8"), msgs[0].Data)
	require.Equal(t, []byte("content7"), msgs[1].Data)

	// Order descending by default
	msgs, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:    2,
			IdCursor: thirdMsg.Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("content5"), msgs[0].Data)
	require.Equal(t, []byte("content3"), msgs[1].Data)

	// Next page from previous response
	lastMsg := msgs[len(msgs)-1]
	msgs, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:    2,
			IdCursor: lastMsg.Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, []byte("content1"), msgs[0].Data)

	// Order ascending
	msgs, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:     2,
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
			IdCursor:  fifthMsg.Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("content5"), msgs[0].Data)
	require.Equal(t, []byte("content6"), msgs[1].Data)

	// Next page from previous response
	lastMsg = msgs[len(msgs)-1]
	msgs, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:     2,
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
			IdCursor:  lastMsg.Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("content7"), msgs[0].Data)
	require.Equal(t, []byte("content8"), msgs[1].Data)
}

func TestQueryWelcomeMessagesV1_Paginate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content1"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation2"), []byte("content2"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content3"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation2"), []byte("content4"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content5"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content6"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content7"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content8"))
	require.NoError(t, err)

	msgs, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
	})
	require.NoError(t, err)
	require.Len(t, msgs, 6)
	require.Equal(t, []byte("content8"), msgs[0].Data)
	require.Equal(t, []byte("content7"), msgs[1].Data)
	require.Equal(t, []byte("content6"), msgs[2].Data)
	require.Equal(t, []byte("content5"), msgs[3].Data)
	require.Equal(t, []byte("content3"), msgs[4].Data)
	require.Equal(t, []byte("content1"), msgs[5].Data)

	thirdMsg := msgs[2]
	fifthMsg := msgs[4]

	msgs, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("content8"), msgs[0].Data)
	require.Equal(t, []byte("content7"), msgs[1].Data)

	// Order descending by default
	msgs, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:    2,
			IdCursor: thirdMsg.Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("content5"), msgs[0].Data)
	require.Equal(t, []byte("content3"), msgs[1].Data)

	// Next page from previous response
	lastMsg := msgs[len(msgs)-1]
	msgs, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:    2,
			IdCursor: lastMsg.Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, []byte("content1"), msgs[0].Data)

	// Order ascending
	msgs, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:     2,
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
			IdCursor:  fifthMsg.Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("content5"), msgs[0].Data)
	require.Equal(t, []byte("content6"), msgs[1].Data)

	// Next page from previous response
	lastMsg = msgs[len(msgs)-1]
	msgs, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:     2,
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
			IdCursor:  lastMsg.Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("content7"), msgs[0].Data)
	require.Equal(t, []byte("content8"), msgs[1].Data)
}
