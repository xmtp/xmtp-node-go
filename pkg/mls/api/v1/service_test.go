package api

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	mlsv1 "github.com/xmtp/proto/v3/go/mls/api/v1"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
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

func (m *mockedMLSValidationService) ValidateGroupMessages(ctx context.Context, groupMessages []*mlsv1.GroupMessageInput) ([]mlsvalidate.GroupMessageValidationResult, error) {
	args := m.Called(ctx, groupMessages)

	return args.Get(0).([]mlsvalidate.GroupMessageValidationResult), args.Error(1)
}

func newMockedValidationService() *mockedMLSValidationService {
	return new(mockedMLSValidationService)
}

func (m *mockedMLSValidationService) mockValidateKeyPackages(installationId []byte, accountAddress string) *mock.Call {
	return m.On("ValidateKeyPackages", mock.Anything, mock.Anything).Return([]mlsvalidate.IdentityValidationResult{
		{
			InstallationKey:    installationId,
			AccountAddress:     accountAddress,
			CredentialIdentity: []byte("test"),
			Expiration:         0,
		},
	}, nil)
}

func (m *mockedMLSValidationService) mockValidateGroupMessages(groupId []byte) *mock.Call {
	return m.On("ValidateGroupMessages", mock.Anything, mock.Anything).Return([]mlsvalidate.GroupMessageValidationResult{
		{
			GroupId: fmt.Sprintf("%x", groupId),
		},
	}, nil)
}

func newTestService(t *testing.T, ctx context.Context) (*Service, *bun.DB, *mockedMLSValidationService, func()) {
	log := test.NewLog(t)
	db, _, mlsDbCleanup := test.NewMLSDB(t)
	store, err := mlsstore.New(ctx, mlsstore.Config{
		Log: log,
		DB:  db,
	})
	require.NoError(t, err)
	mlsValidationService := newMockedValidationService()

	svc, err := NewService(log, store, mlsValidationService, func(ctx context.Context, wm *wakupb.WakuMessage) error {
		return nil
	})
	require.NoError(t, err)

	return svc, db, mlsValidationService, mlsDbCleanup
}

func TestRegisterInstallation(t *testing.T) {
	ctx := context.Background()
	svc, mlsDb, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId := test.RandomBytes(32)
	accountAddress := test.RandomString(32)

	mlsValidationService.mockValidateKeyPackages(installationId, accountAddress)

	res, err := svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
	})

	require.NoError(t, err)
	require.Equal(t, installationId, res.InstallationKey)

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

	res, err := svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
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

	res, err := svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	uploadRes, err := svc.UploadKeyPackage(ctx, &mlsv1.UploadKeyPackageRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
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

	res, err := svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
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

	res, err = svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test2"),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	consumeRes, err := svc.FetchKeyPackages(ctx, &mlsv1.FetchKeyPackagesRequest{
		InstallationKeys: [][]byte{installationId1, installationId2},
	})
	require.NoError(t, err)
	require.NotNil(t, consumeRes)
	require.Len(t, consumeRes.KeyPackages, 2)
	require.Equal(t, []byte("test"), consumeRes.KeyPackages[0].KeyPackageTlsSerialized)
	require.Equal(t, []byte("test2"), consumeRes.KeyPackages[1].KeyPackageTlsSerialized)

	// Now do it with the installationIds reversed
	consumeRes, err = svc.FetchKeyPackages(ctx, &mlsv1.FetchKeyPackagesRequest{
		InstallationKeys: [][]byte{installationId2, installationId1},
	})

	require.NoError(t, err)
	require.NotNil(t, consumeRes)
	require.Len(t, consumeRes.KeyPackages, 2)
	require.Equal(t, []byte("test2"), consumeRes.KeyPackages[0].KeyPackageTlsSerialized)
	require.Equal(t, []byte("test"), consumeRes.KeyPackages[1].KeyPackageTlsSerialized)
}

// Trying to fetch key packages that don't exist should return nil
func TestFetchKeyPackagesFail(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	consumeRes, err := svc.FetchKeyPackages(ctx, &mlsv1.FetchKeyPackagesRequest{
		InstallationKeys: [][]byte{test.RandomBytes(32)},
	})
	require.Nil(t, err)
	require.Equal(t, []*mlsv1.FetchKeyPackagesResponse_KeyPackage{nil}, consumeRes.KeyPackages)
}

func TestSendGroupMessages(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))

	mlsValidationService.mockValidateGroupMessages(groupId)

	_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
		Messages: []*mlsv1.GroupMessageInput{
			{
				Version: &mlsv1.GroupMessageInput_V1_{
					V1: &mlsv1.GroupMessageInput_V1{
						Data: []byte("test"),
					},
				},
			},
		},
	})
	require.NoError(t, err)

	resp, err := svc.store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: groupId,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, resp.Messages[0].GetV1().Data, []byte("test"))
	require.NotEmpty(t, resp.Messages[0].GetV1().CreatedNs)
}

func TestSendWelcomeMessages(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId := []byte(test.RandomString(32))

	_, err := svc.SendWelcomeMessages(ctx, &mlsv1.SendWelcomeMessagesRequest{
		Messages: []*mlsv1.WelcomeMessageInput{
			{
				Version: &mlsv1.WelcomeMessageInput_V1_{
					V1: &mlsv1.WelcomeMessageInput_V1{
						InstallationKey: []byte(installationId),
						Data:            []byte("test"),
					},
				},
			},
		},
	})
	require.NoError(t, err)

	resp, err := svc.store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: installationId,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, resp.Messages[0].GetV1().Data, []byte("test"))
	require.NotEmpty(t, resp.Messages[0].GetV1().CreatedNs)
}

func TestGetIdentityUpdates(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId := test.RandomBytes(32)
	accountAddress := test.RandomString(32)

	mockCall := mlsValidationService.mockValidateKeyPackages(installationId, accountAddress)

	_, err := svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
	})
	require.NoError(t, err)

	identityUpdates, err := svc.GetIdentityUpdates(ctx, &mlsv1.GetIdentityUpdatesRequest{
		AccountAddresses: []string{accountAddress},
	})
	require.NoError(t, err)
	require.NotNil(t, identityUpdates)
	require.Len(t, identityUpdates.Updates, 1)
	require.Equal(t, identityUpdates.Updates[0].Updates[0].GetNewInstallation().InstallationKey, installationId)
	require.Equal(t, identityUpdates.Updates[0].Updates[0].GetNewInstallation().CredentialIdentity, []byte("test"))

	for _, walletUpdate := range identityUpdates.Updates {
		for _, update := range walletUpdate.Updates {
			require.Equal(t, installationId, update.GetNewInstallation().InstallationKey)
		}
	}

	mockCall.Unset()
	mlsValidationService.mockValidateKeyPackages(test.RandomBytes(32), accountAddress)
	_, err = svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
	})
	require.NoError(t, err)

	identityUpdates, err = svc.GetIdentityUpdates(ctx, &mlsv1.GetIdentityUpdatesRequest{
		AccountAddresses: []string{accountAddress},
	})
	require.NoError(t, err)
	require.Len(t, identityUpdates.Updates, 1)
	require.Len(t, identityUpdates.Updates[0].Updates, 2)
}

func TestSubscribeGroupMessages(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))

	mlsValidationService.mockValidateGroupMessages(groupId)

	_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
		Messages: []*mlsv1.GroupMessageInput{
			{
				Version: &mlsv1.GroupMessageInput_V1_{
					V1: &mlsv1.GroupMessageInput_V1{
						Data: []byte("data"),
					},
				},
			},
		},
	})
	require.NoError(t, err)

	resp, err := svc.store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: groupId,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, resp.Messages[0].GetV1().Data, []byte("data"))
	require.NotEmpty(t, resp.Messages[0].GetV1().CreatedNs)
}
