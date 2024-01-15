package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
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

	return svc, db, mlsValidationService, func() {
		svc.Close()
		mlsDbCleanup()
	}
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

	msgs, err := svc.store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: groupId,
	})
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, msgs[0].Data, []byte("test"))
	require.NotEmpty(t, msgs[0].CreatedAt)
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

	msgs, err := svc.store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: installationId,
	})
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, msgs[0].Data, []byte("test"))
	require.NotEmpty(t, msgs[0].CreatedAt)
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

func TestSubscribeGroupMessages_WithoutCursor(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))

	// Initial message that does not get included in the stream.
	mlsValidationService.mockValidateGroupMessages(groupId)
	_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
		Messages: []*mlsv1.GroupMessageInput{
			{
				Version: &mlsv1.GroupMessageInput_V1_{
					V1: &mlsv1.GroupMessageInput_V1{
						Data: []byte("data0"),
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Set of 10 messages that are included in the stream.
	msgs := make([]*mlsv1.GroupMessage, 10)
	for i := 0; i < 10; i++ {
		msgs[i] = &mlsv1.GroupMessage{
			Version: &mlsv1.GroupMessage_V1_{
				V1: &mlsv1.GroupMessage_V1{
					GroupId: groupId,
					Data:    []byte(fmt.Sprintf("data%d", i+1)),
				},
			},
		}
	}

	// Set up expectations of streaming the 10 messages.
	ctrl := gomock.NewController(t)
	stream := NewMockMlsApi_SubscribeGroupMessagesServer(ctrl)
	stream.EXPECT().SendHeader(map[string][]string{"subscribed": {"true"}})
	for _, msg := range msgs {
		stream.EXPECT().Send(newGroupMessageDataEqualsMatcher(msg)).Return(nil).Times(1)
	}
	stream.EXPECT().Context().Return(ctx).AnyTimes()

	go func() {
		err := svc.SubscribeGroupMessages(&mlsv1.SubscribeGroupMessagesRequest{
			Filters: []*mlsv1.SubscribeGroupMessagesRequest_Filter{
				{
					GroupId: groupId,
				},
			},
		}, stream)
		require.NoError(t, err)
	}()
	time.Sleep(50 * time.Millisecond)

	// Send the messages (store and relay).
	for i, msg := range msgs {
		mlsValidationService.mockValidateGroupMessages(groupId)
		_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
			Messages: []*mlsv1.GroupMessageInput{
				{
					Version: &mlsv1.GroupMessageInput_V1_{
						V1: &mlsv1.GroupMessageInput_V1{
							Data: msg.GetV1().Data,
						},
					},
				},
			},
		})
		require.NoError(t, err)

		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)
		err = svc.HandleIncomingWakuRelayMessage(&wakupb.WakuMessage{
			ContentTopic: topic.BuildMLSV1GroupTopic(msg.GetV1().GroupId),
			Timestamp:    int64(msg.GetV1().CreatedNs),
			Payload:      msgB,
		})
		require.NoError(t, err)

		if i == 4 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Expectations should eventually be satisfied.
	require.Eventually(t, ctrl.Satisfied, 5*time.Second, 100*time.Millisecond)
}

func TestSubscribeGroupMessages_WithCursor(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))

	// Initial message before stream starts.
	mlsValidationService.mockValidateGroupMessages(groupId)
	initialMsgs := []*mlsv1.GroupMessageInput{
		{
			Version: &mlsv1.GroupMessageInput_V1_{
				V1: &mlsv1.GroupMessageInput_V1{
					Data: []byte("data1"),
				},
			},
		},
		{
			Version: &mlsv1.GroupMessageInput_V1_{
				V1: &mlsv1.GroupMessageInput_V1{
					Data: []byte("data2"),
				},
			},
		},
		{
			Version: &mlsv1.GroupMessageInput_V1_{
				V1: &mlsv1.GroupMessageInput_V1{
					Data: []byte("data3"),
				},
			},
		},
	}
	for _, msg := range initialMsgs {
		_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
			Messages: []*mlsv1.GroupMessageInput{msg},
		})
		require.NoError(t, err)
	}

	// Set of 10 messages that are included in the stream.
	msgs := make([]*mlsv1.GroupMessage, 10)
	for i := 0; i < 10; i++ {
		msgs[i] = &mlsv1.GroupMessage{
			Version: &mlsv1.GroupMessage_V1_{
				V1: &mlsv1.GroupMessage_V1{
					GroupId: groupId,
					Data:    []byte(fmt.Sprintf("data%d", i+4)),
				},
			},
		}
	}

	// Set up expectations of streaming the 10 messages.
	ctrl := gomock.NewController(t)
	stream := NewMockMlsApi_SubscribeGroupMessagesServer(ctrl)
	stream.EXPECT().SendHeader(map[string][]string{"subscribed": {"true"}})
	stream.EXPECT().Send(newGroupMessageDataEqualsMatcher(&mlsv1.GroupMessage{
		Version: &mlsv1.GroupMessage_V1_{
			V1: &mlsv1.GroupMessage_V1{
				Data: []byte("data3"),
			},
		},
	})).Return(nil).Times(1)
	for _, msg := range msgs {
		stream.EXPECT().Send(newGroupMessageDataEqualsMatcher(msg)).Return(nil).Times(1)
	}
	stream.EXPECT().Context().Return(ctx).AnyTimes()

	go func() {
		err := svc.SubscribeGroupMessages(&mlsv1.SubscribeGroupMessagesRequest{
			Filters: []*mlsv1.SubscribeGroupMessagesRequest_Filter{
				{
					GroupId:  groupId,
					IdCursor: 2,
				},
			},
		}, stream)
		require.NoError(t, err)
	}()
	time.Sleep(50 * time.Millisecond)

	// Send the messages (store and relay).
	for i, msg := range msgs {
		mlsValidationService.mockValidateGroupMessages(groupId)
		_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
			Messages: []*mlsv1.GroupMessageInput{
				{
					Version: &mlsv1.GroupMessageInput_V1_{
						V1: &mlsv1.GroupMessageInput_V1{
							Data: msg.GetV1().Data,
						},
					},
				},
			},
		})
		require.NoError(t, err)

		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)
		err = svc.HandleIncomingWakuRelayMessage(&wakupb.WakuMessage{
			ContentTopic: topic.BuildMLSV1GroupTopic(msg.GetV1().GroupId),
			Timestamp:    int64(msg.GetV1().CreatedNs),
			Payload:      msgB,
		})
		require.NoError(t, err)

		if i == 4 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Expectations should eventually be satisfied.
	require.Eventually(t, ctrl.Satisfied, 5*time.Second, 100*time.Millisecond)
}

func TestSubscribeWelcomeMessages_WithoutCursor(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationKey := []byte(test.RandomString(32))

	// Initial message that does not get included in the stream.
	_, err := svc.SendWelcomeMessages(ctx, &mlsv1.SendWelcomeMessagesRequest{
		Messages: []*mlsv1.WelcomeMessageInput{
			{
				Version: &mlsv1.WelcomeMessageInput_V1_{
					V1: &mlsv1.WelcomeMessageInput_V1{
						InstallationKey: installationKey,
						Data:            []byte("data0"),
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Set of 10 messages that are included in the stream.
	msgs := make([]*mlsv1.WelcomeMessage, 10)
	for i := 0; i < 10; i++ {
		msgs[i] = &mlsv1.WelcomeMessage{
			Version: &mlsv1.WelcomeMessage_V1_{
				V1: &mlsv1.WelcomeMessage_V1{
					InstallationKey: installationKey,
					Data:            []byte(fmt.Sprintf("data%d", i+1)),
				},
			},
		}
	}

	// Set up expectations of streaming the 10 messages.
	ctrl := gomock.NewController(t)
	stream := NewMockMlsApi_SubscribeWelcomeMessagesServer(ctrl)
	stream.EXPECT().SendHeader(map[string][]string{"subscribed": {"true"}})
	for _, msg := range msgs {
		stream.EXPECT().Send(newWelcomeMessageDataEqualsMatcher(msg)).Return(nil).Times(1)
	}
	stream.EXPECT().Context().Return(ctx).AnyTimes()

	go func() {
		err := svc.SubscribeWelcomeMessages(&mlsv1.SubscribeWelcomeMessagesRequest{
			Filters: []*mlsv1.SubscribeWelcomeMessagesRequest_Filter{
				{
					InstallationKey: installationKey,
				},
			},
		}, stream)
		require.NoError(t, err)
	}()
	time.Sleep(50 * time.Millisecond)

	// Send the messages (store and relay).
	for i, msg := range msgs {
		_, err := svc.SendWelcomeMessages(ctx, &mlsv1.SendWelcomeMessagesRequest{
			Messages: []*mlsv1.WelcomeMessageInput{
				{
					Version: &mlsv1.WelcomeMessageInput_V1_{
						V1: &mlsv1.WelcomeMessageInput_V1{
							InstallationKey: installationKey,
							Data:            msg.GetV1().Data,
						},
					},
				},
			},
		})
		require.NoError(t, err)

		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)
		err = svc.HandleIncomingWakuRelayMessage(&wakupb.WakuMessage{
			ContentTopic: topic.BuildMLSV1WelcomeTopic(msg.GetV1().InstallationKey),
			Timestamp:    int64(msg.GetV1().CreatedNs),
			Payload:      msgB,
		})
		require.NoError(t, err)

		if i == 4 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Expectations should eventually be satisfied.
	require.Eventually(t, ctrl.Satisfied, 5*time.Second, 100*time.Millisecond)
}

func TestSubscribeWelcomeMessages_WithCursor(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationKey := []byte(test.RandomString(32))

	// Initial message before stream starts.
	initialMsgs := []*mlsv1.WelcomeMessageInput{
		{
			Version: &mlsv1.WelcomeMessageInput_V1_{
				V1: &mlsv1.WelcomeMessageInput_V1{
					InstallationKey: installationKey,
					Data:            []byte("data1"),
				},
			},
		},
		{
			Version: &mlsv1.WelcomeMessageInput_V1_{
				V1: &mlsv1.WelcomeMessageInput_V1{
					InstallationKey: installationKey,
					Data:            []byte("data2"),
				},
			},
		},
		{
			Version: &mlsv1.WelcomeMessageInput_V1_{
				V1: &mlsv1.WelcomeMessageInput_V1{
					InstallationKey: installationKey,
					Data:            []byte("data3"),
				},
			},
		},
	}
	for _, msg := range initialMsgs {
		_, err := svc.SendWelcomeMessages(ctx, &mlsv1.SendWelcomeMessagesRequest{
			Messages: []*mlsv1.WelcomeMessageInput{msg},
		})
		require.NoError(t, err)
	}

	// Set of 10 messages that are included in the stream.
	msgs := make([]*mlsv1.WelcomeMessage, 10)
	for i := 0; i < 10; i++ {
		msgs[i] = &mlsv1.WelcomeMessage{
			Version: &mlsv1.WelcomeMessage_V1_{
				V1: &mlsv1.WelcomeMessage_V1{
					InstallationKey: installationKey,
					Data:            []byte(fmt.Sprintf("data%d", i+4)),
				},
			},
		}
	}

	// Set up expectations of streaming the 10 messages.
	ctrl := gomock.NewController(t)
	stream := NewMockMlsApi_SubscribeWelcomeMessagesServer(ctrl)
	stream.EXPECT().SendHeader(map[string][]string{"subscribed": {"true"}})
	stream.EXPECT().Send(newWelcomeMessageDataEqualsMatcher(&mlsv1.WelcomeMessage{
		Version: &mlsv1.WelcomeMessage_V1_{
			V1: &mlsv1.WelcomeMessage_V1{
				Data: []byte("data3"),
			},
		},
	})).Return(nil).Times(1)
	for _, msg := range msgs {
		stream.EXPECT().Send(newWelcomeMessageDataEqualsMatcher(msg)).Return(nil).Times(1)
	}
	stream.EXPECT().Context().Return(ctx).AnyTimes()

	go func() {
		err := svc.SubscribeWelcomeMessages(&mlsv1.SubscribeWelcomeMessagesRequest{
			Filters: []*mlsv1.SubscribeWelcomeMessagesRequest_Filter{
				{
					InstallationKey: installationKey,
					IdCursor:        2,
				},
			},
		}, stream)
		require.NoError(t, err)
	}()
	time.Sleep(50 * time.Millisecond)

	// Send the messages (store and relay).
	for i, msg := range msgs {
		_, err := svc.SendWelcomeMessages(ctx, &mlsv1.SendWelcomeMessagesRequest{
			Messages: []*mlsv1.WelcomeMessageInput{
				{
					Version: &mlsv1.WelcomeMessageInput_V1_{
						V1: &mlsv1.WelcomeMessageInput_V1{
							InstallationKey: installationKey,
							Data:            msg.GetV1().Data,
						},
					},
				},
			},
		})
		require.NoError(t, err)

		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)
		err = svc.HandleIncomingWakuRelayMessage(&wakupb.WakuMessage{
			ContentTopic: topic.BuildMLSV1WelcomeTopic(msg.GetV1().InstallationKey),
			Timestamp:    int64(msg.GetV1().CreatedNs),
			Payload:      msgB,
		})
		require.NoError(t, err)

		if i == 4 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Expectations should eventually be satisfied.
	require.Eventually(t, ctrl.Satisfied, 5*time.Second, 100*time.Millisecond)
}

type groupMessageDataEqualsMatcher struct {
	obj *mlsv1.GroupMessage
}

func newGroupMessageDataEqualsMatcher(obj *mlsv1.GroupMessage) *groupMessageDataEqualsMatcher {
	return &groupMessageDataEqualsMatcher{obj}
}

func (m *groupMessageDataEqualsMatcher) Matches(obj interface{}) bool {
	return bytes.Equal(m.obj.GetV1().Data, obj.(*mlsv1.GroupMessage).GetV1().Data)
}

func (m *groupMessageDataEqualsMatcher) String() string {
	return m.obj.String()
}

type welcomeMessageDataEqualsMatcher struct {
	obj *mlsv1.WelcomeMessage
}

func newWelcomeMessageDataEqualsMatcher(obj *mlsv1.WelcomeMessage) *welcomeMessageDataEqualsMatcher {
	return &welcomeMessageDataEqualsMatcher{obj}
}

func (m *welcomeMessageDataEqualsMatcher) Matches(obj interface{}) bool {
	return bytes.Equal(m.obj.GetV1().Data, obj.(*mlsv1.WelcomeMessage).GetV1().Data)
}

func (m *welcomeMessageDataEqualsMatcher) String() string {
	return m.obj.String()
}
