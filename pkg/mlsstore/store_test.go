package mlsstore

import (
	"context"
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
	installationId := test.RandomString(32)
	walletAddress := test.RandomString(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, test.RandomBytes(32))
	require.NoError(t, err)

	installationFromDb := &Installation{}
	require.NoError(t, store.db.NewSelect().Model(installationFromDb).Where("id = ?", installationId).Scan(ctx))
	require.Equal(t, walletAddress, installationFromDb.WalletAddress)

	keyPackageFromDB := &KeyPackage{}
	require.NoError(t, store.db.NewSelect().Model(keyPackageFromDB).Where("installation_id = ?", installationId).Scan(ctx))
	require.Equal(t, installationId, keyPackageFromDB.InstallationId)
}

func TestCreateInstallationIdempotent(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := test.RandomString(32)
	walletAddress := test.RandomString(32)
	keyPackage := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, keyPackage)
	require.NoError(t, err)
	err = store.CreateInstallation(ctx, installationId, walletAddress, test.RandomBytes(32))
	require.NoError(t, err)

	keyPackageFromDb := &KeyPackage{}
	require.NoError(t, store.db.NewSelect().Model(keyPackageFromDb).Where("installation_id = ?", installationId).Scan(ctx))
	require.Equal(t, keyPackage, keyPackageFromDb.Data)
}

func TestInsertKeyPackages(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := test.RandomString(32)
	walletAddress := test.RandomString(32)
	keyPackage := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, keyPackage)
	require.NoError(t, err)

	keyPackage2 := test.RandomBytes(32)
	err = store.InsertKeyPackages(ctx, []*KeyPackage{{
		ID:             buildKeyPackageId(keyPackage2),
		InstallationId: installationId,
		CreatedAt:      nowNs(),
		IsLastResort:   false,
		Data:           keyPackage2,
	}})
	require.NoError(t, err)

	keyPackagesFromDb := []*KeyPackage{}
	require.NoError(t, store.db.NewSelect().Model(&keyPackagesFromDb).Where("installation_id = ?", installationId).Scan(ctx))
	require.Len(t, keyPackagesFromDb, 2)

	hasLastResort := false
	hasRegular := false
	for _, keyPackageFromDb := range keyPackagesFromDb {
		require.Equal(t, installationId, keyPackageFromDb.InstallationId)
		if keyPackageFromDb.IsLastResort {
			hasLastResort = true
		}
		if !keyPackageFromDb.IsLastResort {
			hasRegular = true
			require.Equal(t, keyPackage2, keyPackageFromDb.Data)
		}
	}

	require.True(t, hasLastResort)
	require.True(t, hasRegular)
}

func TestConsumeLastResortKeyPackage(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := test.RandomString(32)
	walletAddress := test.RandomString(32)
	keyPackage := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, keyPackage)
	require.NoError(t, err)

	consumeResult, err := store.ConsumeKeyPackages(ctx, []string{installationId})
	require.NoError(t, err)
	require.Len(t, consumeResult, 1)
	require.Equal(t, keyPackage, consumeResult[0].Data)
	require.Equal(t, installationId, consumeResult[0].InstallationId)
}

func TestConsumeMultipleKeyPackages(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := test.RandomString(32)
	walletAddress := test.RandomString(32)
	keyPackage := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, keyPackage)
	require.NoError(t, err)

	keyPackage2 := test.RandomBytes(32)
	require.NoError(t, store.InsertKeyPackages(ctx, []*KeyPackage{{
		ID:             buildKeyPackageId(keyPackage2),
		InstallationId: installationId,
		CreatedAt:      nowNs(),
		IsLastResort:   false,
		Data:           keyPackage2,
	}}))

	consumeResult, err := store.ConsumeKeyPackages(ctx, []string{installationId})
	require.NoError(t, err)
	require.Len(t, consumeResult, 1)
	require.Equal(t, keyPackage2, consumeResult[0].Data)
	require.Equal(t, installationId, consumeResult[0].InstallationId)

	consumeResult, err = store.ConsumeKeyPackages(ctx, []string{installationId})
	require.NoError(t, err)
	require.Len(t, consumeResult, 1)
	// Now we are out of regular key packages. Expect to consume the last resort
	require.Equal(t, keyPackage, consumeResult[0].Data)
}

func TestGetIdentityUpdates(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	walletAddress := test.RandomString(32)

	installationId1 := test.RandomString(32)
	keyPackage1 := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId1, walletAddress, keyPackage1)
	require.NoError(t, err)

	installationId2 := test.RandomString(32)
	keyPackage2 := test.RandomBytes(32)

	err = store.CreateInstallation(ctx, installationId2, walletAddress, keyPackage2)
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
	installationId1 := test.RandomString(32)
	keyPackage1 := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId1, walletAddress1, keyPackage1)
	require.NoError(t, err)

	walletAddress2 := test.RandomString(32)
	installationId2 := test.RandomString(32)
	keyPackage2 := test.RandomBytes(32)

	err = store.CreateInstallation(ctx, installationId2, walletAddress2, keyPackage2)
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
