package authz

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	testutils "github.com/xmtp/xmtp-node-go/pkg/testing"
	"go.uber.org/zap"
)

const (
	DENIED_IP_ADDRESS   = "0.0.0.0"
	ALLOWED_IP_ADDRESS  = "1.1.1.1"
	PRIORITY_IP_ADDRESS = "9.9.9.9"
)

func newAllowLister(t *testing.T, db *bun.DB) *DatabaseAllowList {
	logger, _ := zap.NewDevelopment()
	allowLister, err := NewDatabaseAllowList(context.Background(), db, logger)
	require.NoError(t, err)

	t.Cleanup(allowLister.Stop)

	return allowLister
}

func fillDb(t *testing.T, db *bun.DB) (ipAddresses []IPAddress) {
	ipAddresses = []IPAddress{
		{
			IPAddress:  ALLOWED_IP_ADDRESS,
			Permission: AllowAll.String(),
		},
		{
			IPAddress:  DENIED_IP_ADDRESS,
			Permission: Denied.String(),
		},
		{
			IPAddress:  PRIORITY_IP_ADDRESS,
			Permission: Priority.String(),
		},
	}
	_, err := db.NewInsert().Model(&ipAddresses).Exec(context.Background())

	require.NoError(t, err)

	return
}

func TestPermissionCheck(t *testing.T) {
	db, _, cleanup := testutils.NewAuthzDB(t)
	defer cleanup()
	ipAddresses := fillDb(t, db)
	allowLister := newAllowLister(t, db)

	for _, ipAddr := range ipAddresses {
		expectedValue := permissionFromString(ipAddr.Permission)
		require.Equal(t, allowLister.GetPermission(ipAddr.IPAddress), expectedValue)
	}
}

func TestDelete(t *testing.T) {
	db, _, cleanup := testutils.NewAuthzDB(t)
	defer cleanup()
	ipAddresses := fillDb(t, db)
	allowLister := newAllowLister(t, db)

	allowedIP := ipAddresses[0]
	require.Equal(t, allowLister.GetPermission(allowedIP.IPAddress), AllowAll)

	// Delete the allowed IP record
	require.NotNil(t, allowedIP.ID)
	updateModel := IPAddress{ID: allowedIP.ID}
	now := time.Now().UTC()
	_, err := allowLister.db.NewUpdate().
		Model(&updateModel).Set("deleted_at = ?", now).
		Where("id = ?", allowedIP.ID).
		Exec(context.Background())

	require.NoError(t, err)
	require.NoError(t, allowLister.loadPermissions())

	require.Equal(t, allowLister.GetPermission(allowedIP.IPAddress), Unspecified)
}

func TestUnknownIP(t *testing.T) {
	db, _, cleanup := testutils.NewAuthzDB(t)
	t.Cleanup(cleanup)
	allowLister := newAllowLister(t, db)

	unknownIPAddress := "172.16.0.1"

	require.Equal(t, allowLister.GetPermission(unknownIPAddress), Unspecified)
}
