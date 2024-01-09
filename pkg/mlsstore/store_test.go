package mlsstore

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
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
