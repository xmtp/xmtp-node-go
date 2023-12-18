package api

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	v1 "github.com/xmtp/proto/v3/go/message_api/v1"
	proto "github.com/xmtp/proto/v3/go/message_api/v3"
	messageContents "github.com/xmtp/proto/v3/go/mls/message_contents"
	"github.com/xmtp/xmtp-node-go/pkg/mlsstore"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

type mockedMLSValidationService struct {
	mock.Mock
}

func (m *mockedMLSValidationService) ValidateKeyPackages(ctx context.Context, keyPackages [][]byte) ([]mlsvalidate.IdentityValidationResult, error) {
	args := m.Called(ctx, keyPackages)

	response := args.Get(0)
	if response == nil {
		return nil, args.Error(1)
	}

	return response.([]mlsvalidate.IdentityValidationResult), args.Error(1)
}

func (m *mockedMLSValidationService) ValidateGroupMessages(ctx context.Context, groupMessages [][]byte) ([]mlsvalidate.GroupMessageValidationResult, error) {
	args := m.Called(ctx, groupMessages)

	return args.Get(0).([]mlsvalidate.GroupMessageValidationResult), args.Error(1)
}

func newMockedValidationService() *mockedMLSValidationService {
	return new(mockedMLSValidationService)
}

func (m *mockedMLSValidationService) mockValidateKeyPackages(installationId []byte, accountAddress string) *mock.Call {
	return m.On("ValidateKeyPackages", mock.Anything, mock.Anything).Return([]mlsvalidate.IdentityValidationResult{
		{
			InstallationId:     installationId,
			AccountAddress:     accountAddress,
			CredentialIdentity: []byte("test"),
			Expiration:         0,
		},
	}, nil)
}

func (m *mockedMLSValidationService) mockValidateGroupMessages(groupId string) *mock.Call {
	return m.On("ValidateGroupMessages", mock.Anything, mock.Anything).Return([]mlsvalidate.GroupMessageValidationResult{
		{
			GroupId: groupId,
		},
	}, nil)
}

func newTestService(t *testing.T, ctx context.Context) (*Service, *bun.DB, *mockedMLSValidationService, func()) {
	log := test.NewLog(t)
	mlsDb, _, mlsDbCleanup := test.NewMLSDB(t)
	mlsStore, err := mlsstore.New(ctx, mlsstore.Config{
		Log: log,
		DB:  mlsDb,
	})
	require.NoError(t, err)
	messageDb, _, messageDbCleanup := test.NewDB(t)
	messageStore, err := store.New(&store.Config{
		Log:       log,
		DB:        messageDb,
		ReaderDB:  messageDb,
		CleanerDB: messageDb,
	})
	require.NoError(t, err)
	node, nodeCleanup := test.NewNode(t)
	mlsValidationService := newMockedValidationService()

	svc, err := NewService(node, log, messageStore, mlsStore, mlsValidationService)
	require.NoError(t, err)

	return svc, mlsDb, mlsValidationService, func() {
		messageStore.Close()
		mlsDbCleanup()
		messageDbCleanup()
		nodeCleanup()
	}
}

func TestRegisterInstallation(t *testing.T) {
	ctx := context.Background()
	svc, mlsDb, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId := test.RandomBytes(32)
	accountAddress := test.RandomString(32)

	mlsValidationService.mockValidateKeyPackages(installationId, accountAddress)

	res, err := svc.RegisterInstallation(ctx, &proto.RegisterInstallationRequest{
		KeyPackage: &proto.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
	})

	require.NoError(t, err)
	require.Equal(t, installationId, res.InstallationId)

	installations := []mlsstore.Installation{}
	err = mlsDb.NewSelect().Model(&installations).Where("id = ?", installationId).Scan(ctx)
	require.NoError(t, err)

	require.Len(t, installations, 1)
	require.Equal(t, accountAddress, installations[0].WalletAddress)
}

func TestRegisterInstallationError(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	mlsValidationService.On("ValidateKeyPackages", ctx, mock.Anything).Return(nil, errors.New("error validating"))

	res, err := svc.RegisterInstallation(ctx, &proto.RegisterInstallationRequest{
		KeyPackage: &proto.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
	})
	require.Error(t, err)
	require.Nil(t, res)
}

func TestUploadKeyPackage(t *testing.T) {
	ctx := context.Background()
	svc, mlsDb, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId := test.RandomBytes(32)
	accountAddress := test.RandomString(32)

	mlsValidationService.mockValidateKeyPackages(installationId, accountAddress)

	res, err := svc.RegisterInstallation(ctx, &proto.RegisterInstallationRequest{
		KeyPackage: &proto.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	uploadRes, err := svc.UploadKeyPackage(ctx, &proto.UploadKeyPackageRequest{
		KeyPackage: &proto.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test2"),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, uploadRes)

	installation := &mlsstore.Installation{}
	err = mlsDb.NewSelect().Model(installation).Where("id = ?", installationId).Scan(ctx)
	require.NoError(t, err)
}

func TestFetchKeyPackages(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId1 := test.RandomBytes(32)
	accountAddress1 := test.RandomString(32)

	mockCall := mlsValidationService.mockValidateKeyPackages(installationId1, accountAddress1)

	res, err := svc.RegisterInstallation(ctx, &proto.RegisterInstallationRequest{
		KeyPackage: &proto.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	// Add a second key package
	installationId2 := test.RandomBytes(32)
	accountAddress2 := test.RandomString(32)
	// Unset the original mock so we can set a new one
	mockCall.Unset()
	mlsValidationService.mockValidateKeyPackages(installationId2, accountAddress2)

	res, err = svc.RegisterInstallation(ctx, &proto.RegisterInstallationRequest{
		KeyPackage: &proto.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test2"),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	consumeRes, err := svc.FetchKeyPackages(ctx, &proto.FetchKeyPackagesRequest{
		InstallationIds: [][]byte{installationId1, installationId2},
	})
	require.NoError(t, err)
	require.NotNil(t, consumeRes)
	require.Len(t, consumeRes.KeyPackages, 2)
	require.Equal(t, []byte("test"), consumeRes.KeyPackages[0].KeyPackageTlsSerialized)
	require.Equal(t, []byte("test2"), consumeRes.KeyPackages[1].KeyPackageTlsSerialized)

	// Now do it with the installationIds reversed
	consumeRes, err = svc.FetchKeyPackages(ctx, &proto.FetchKeyPackagesRequest{
		InstallationIds: [][]byte{installationId2, installationId1},
	})

	require.NoError(t, err)
	require.NotNil(t, consumeRes)
	require.Len(t, consumeRes.KeyPackages, 2)
	require.Equal(t, []byte("test2"), consumeRes.KeyPackages[0].KeyPackageTlsSerialized)
	require.Equal(t, []byte("test"), consumeRes.KeyPackages[1].KeyPackageTlsSerialized)
}

// Trying to consume key packages that don't exist should fail
func TestFetchKeyPackagesFail(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	consumeRes, err := svc.FetchKeyPackages(ctx, &proto.FetchKeyPackagesRequest{
		InstallationIds: [][]byte{test.RandomBytes(32)},
	})
	require.Error(t, err)
	require.Nil(t, consumeRes)
}

func TestPublishToGroup(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := test.RandomString(32)

	mlsValidationService.mockValidateGroupMessages(groupId)

	_, err := svc.PublishToGroup(ctx, &proto.PublishToGroupRequest{
		Messages: []*messageContents.GroupMessage{{
			Version: &messageContents.GroupMessage_V1_{
				V1: &messageContents.GroupMessage_V1{
					MlsMessageTlsSerialized: []byte("test"),
				},
			},
		}},
	})
	require.NoError(t, err)

	results, err := svc.messageStore.Query(&v1.QueryRequest{
		ContentTopics: []string{fmt.Sprintf("/xmtp/3/g-%s/proto", groupId)},
	})
	require.NoError(t, err)
	require.Len(t, results.Envelopes, 1)
	require.Equal(t, results.Envelopes[0].Message, []byte("test"))
	require.NotNil(t, results.Envelopes[0].TimestampNs)
}

func TestGetIdentityUpdates(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId := test.RandomBytes(32)
	accountAddress := test.RandomString(32)

	mockCall := mlsValidationService.mockValidateKeyPackages(installationId, accountAddress)

	_, err := svc.RegisterInstallation(ctx, &proto.RegisterInstallationRequest{
		KeyPackage: &proto.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
	})
	require.NoError(t, err)

	identityUpdates, err := svc.GetIdentityUpdates(ctx, &proto.GetIdentityUpdatesRequest{
		AccountAddresses: []string{accountAddress},
	})
	require.NoError(t, err)
	require.NotNil(t, identityUpdates)
	require.Len(t, identityUpdates.Updates, 1)
	require.Equal(t, identityUpdates.Updates[0].Updates[0].GetNewInstallation().InstallationId, installationId)
	require.Equal(t, identityUpdates.Updates[0].Updates[0].GetNewInstallation().CredentialIdentity, []byte("test"))

	for _, walletUpdate := range identityUpdates.Updates {
		for _, update := range walletUpdate.Updates {
			require.Equal(t, installationId, update.GetNewInstallation().InstallationId)
		}
	}

	mockCall.Unset()
	mlsValidationService.mockValidateKeyPackages(test.RandomBytes(32), accountAddress)
	_, err = svc.RegisterInstallation(ctx, &proto.RegisterInstallationRequest{
		KeyPackage: &proto.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
	})
	require.NoError(t, err)

	identityUpdates, err = svc.GetIdentityUpdates(ctx, &proto.GetIdentityUpdatesRequest{
		AccountAddresses: []string{accountAddress},
	})
	require.NoError(t, err)
	require.Len(t, identityUpdates.Updates, 1)
	require.Len(t, identityUpdates.Updates[0].Updates, 2)
}
