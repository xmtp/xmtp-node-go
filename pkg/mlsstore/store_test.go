package mlsstore

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func NewTestStore(t *testing.T) (*Store, func()) {
	log := test.NewLog(t)
	db, _, dbCleanup := test.NewMlsDB(t)
	ctx := context.Background()
	c := Config{
		Log: log,
		DB:  db,
	}

	store, err := New(ctx, c)
	require.NoError(t, err)

	return store, dbCleanup
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Reader.Read(b)
	return b
}

func randomString(n int) string {
	return fmt.Sprintf("%x", randomBytes(n))
}

func TestCreateInstallation(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := randomString(32)
	walletAddress := randomString(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, randomBytes(32))
	require.NoError(t, err)

	installationFromDb := &Installation{}
	require.NoError(t, store.db.NewSelect().Model(installationFromDb).Where("id = ?", installationId).Scan(ctx))
	require.Equal(t, walletAddress, installationFromDb.WalletAddress)

	keyPackageFromDb := &KeyPackage{}
	require.NoError(t, store.db.NewSelect().Model(keyPackageFromDb).Where("installation_id = ?", installationId).Scan(ctx))
	require.Equal(t, installationId, keyPackageFromDb.InstallationId)
}

func TestCreateInstallationIdempotent(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := randomString(32)
	walletAddress := randomString(32)
	keyPackage := randomBytes(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, keyPackage)
	require.NoError(t, err)
	err = store.CreateInstallation(ctx, installationId, walletAddress, randomBytes(32))
	require.NoError(t, err)

	keyPackageFromDb := &KeyPackage{}
	require.NoError(t, store.db.NewSelect().Model(keyPackageFromDb).Where("installation_id = ?", installationId).Scan(ctx))
	require.Equal(t, keyPackage, keyPackageFromDb.Data)
}

func TestInsertKeyPackages(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := randomString(32)
	walletAddress := randomString(32)
	keyPackage := randomBytes(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, keyPackage)
	require.NoError(t, err)

	keyPackage2 := randomBytes(32)
	err = store.InsertKeyPackages(ctx, []*KeyPackage{{
		ID:             buildKeyPackageId(keyPackage2),
		InstallationId: installationId,
		CreatedAt:      nowNs(),
		IsLastResort:   false,
		Data:           keyPackage2,
	}})
	require.NoError(t, err)

	keyPackagesFromDb := []*KeyPackage{}
	store.db.NewSelect().Model(&keyPackagesFromDb).Where("installation_id = ?", installationId).Scan(ctx)
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
	installationId := randomString(32)
	walletAddress := randomString(32)
	keyPackage := randomBytes(32)

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
	installationId := randomString(32)
	walletAddress := randomString(32)
	keyPackage := randomBytes(32)

	err := store.CreateInstallation(ctx, installationId, walletAddress, keyPackage)
	require.NoError(t, err)

	keyPackage2 := randomBytes(32)
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
