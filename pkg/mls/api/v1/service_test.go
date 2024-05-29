package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"github.com/xmtp/xmtp-node-go/pkg/mocks"
	"github.com/xmtp/xmtp-node-go/pkg/proto/identity"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"github.com/xmtp/xmtp-node-go/pkg/utils"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func mockValidateInboxIdKeyPackages(m *mocks.MockMLSValidationService, installationId []byte, inboxId string) *mocks.MockMLSValidationService_ValidateInboxIdKeyPackages_Call {
	return m.EXPECT().ValidateInboxIdKeyPackages(mock.Anything, mock.Anything).Return([]mlsvalidate.InboxIdValidationResult{
		{
			InstallationKey: installationId,
			Credential: &identity.MlsCredential{
				InboxId: inboxId,
			},
			Expiration: 0,
		},
	}, nil)
}

func mockValidateGroupMessages(m *mocks.MockMLSValidationService, groupId []byte) *mocks.MockMLSValidationService_ValidateGroupMessages_Call {
	return m.EXPECT().ValidateGroupMessages(mock.Anything, mock.Anything).Return([]mlsvalidate.GroupMessageValidationResult{
		{
			GroupId: fmt.Sprintf("%x", groupId),
		},
	}, nil)
}

func newTestService(t *testing.T, ctx context.Context) (*Service, *bun.DB, *mocks.MockMLSValidationService, func()) {
	log := test.NewLog(t)
	db, _, mlsDbCleanup := test.NewMLSDB(t)
	store, err := mlsstore.New(ctx, mlsstore.Config{
		Log: log,
		DB:  db,
	})
	require.NoError(t, err)
	mockMlsValidation := mocks.NewMockMLSValidationService(t)
	natsServer, err := server.NewServer(&server.Options{
		Port: server.RANDOM_PORT,
	})
	require.NoError(t, err)
	go natsServer.Start()
	if !natsServer.ReadyForConnections(4 * time.Second) {
		t.Fail()
	}

	svc, err := NewService(log, store, mockMlsValidation, natsServer, func(ctx context.Context, wm *wakupb.WakuMessage) error {
		return nil
	})
	require.NoError(t, err)

	return svc, db, mockMlsValidation, func() {
		svc.Close()
		natsServer.Shutdown()
		mlsDbCleanup()
	}
}

func TestRegisterInstallation(t *testing.T) {
	ctx := context.Background()
	svc, mlsDb, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId := test.RandomBytes(32)
	inboxId := test.RandomInboxId()

	mockValidateInboxIdKeyPackages(mlsValidationService, installationId, inboxId)

	res, err := svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
		IsInboxIdCredential: false,
	})

	require.NoError(t, err)
	require.Equal(t, installationId, res.InstallationKey)

	installation, err := queries.New(mlsDb.DB).GetInstallation(ctx, installationId)
	require.NoError(t, err)

	require.Equal(t, inboxId, utils.HexEncode(installation.InboxID))
}

func TestRegisterInstallationError(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	mlsValidationService.EXPECT().ValidateInboxIdKeyPackages(mock.Anything, mock.Anything).Return(nil, errors.New("error validating"))

	res, err := svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
		IsInboxIdCredential: false,
	})
	require.Error(t, err)
	require.Nil(t, res)
}

func TestUploadKeyPackage(t *testing.T) {
	ctx := context.Background()
	svc, mlsDb, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId := test.RandomBytes(32)
	inboxId := test.RandomInboxId()

	mockValidateInboxIdKeyPackages(mlsValidationService, installationId, inboxId)

	res, err := svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
		IsInboxIdCredential: true,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	uploadRes, err := svc.UploadKeyPackage(ctx, &mlsv1.UploadKeyPackageRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test2"),
		},
		IsInboxIdCredential: true,
	})
	require.NoError(t, err)
	require.NotNil(t, uploadRes)

	installation, err := queries.New(mlsDb.DB).GetInstallation(ctx, installationId)
	require.NoError(t, err)
	require.Equal(t, []byte("test2"), installation.KeyPackage)
}

func TestFetchKeyPackages(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId1 := test.RandomBytes(32)
	inboxId := test.RandomInboxId()

	mockCall := mockValidateInboxIdKeyPackages(mlsValidationService, installationId1, inboxId)

	res, err := svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test"),
		},
		IsInboxIdCredential: false,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	// Add a second key package
	installationId2 := test.RandomBytes(32)
	// Unset the original mock so we can set a new one
	mockCall.Unset()

	mockValidateInboxIdKeyPackages(mlsValidationService, installationId2, inboxId)

	res, err = svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: []byte("test2"),
		},
		IsInboxIdCredential: false,
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

	mockValidateGroupMessages(mlsValidationService, groupId)

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
						HpkePublicKey:   []byte("test"),
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

func TestSubscribeGroupMessages_WithoutCursor(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))

	msgs := make([]*mlsv1.GroupMessage, 10)
	for i := 0; i < 10; i++ {
		msgs[i] = &mlsv1.GroupMessage{
			Version: &mlsv1.GroupMessage_V1_{
				V1: &mlsv1.GroupMessage_V1{
					Id:        uint64(i + 1),
					CreatedNs: uint64(i + 1),
					GroupId:   groupId,
					Data:      []byte(fmt.Sprintf("data%d", i+1)),
				},
			},
		}
	}

	stream := mocks.NewMockMlsApi_SubscribeGroupMessagesServer(t)
	stream.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	for _, msg := range msgs {
		stream.EXPECT().Send(mock.MatchedBy(newGroupMessageEqualsMatcher(msg).Matches)).Return(nil).Times(1)
	}
	stream.EXPECT().Context().Return(ctx)

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

	for _, msg := range msgs {
		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)

		err = svc.HandleIncomingWakuRelayMessage(&wakupb.WakuMessage{
			ContentTopic: topic.BuildMLSV1GroupTopic(msg.GetV1().GroupId),
			Timestamp:    int64(msg.GetV1().CreatedNs),
			Payload:      msgB,
		})
		require.NoError(t, err)
	}

	assertExpectationsWithTimeout(t, &stream.Mock, 5*time.Second, 100*time.Millisecond)
}

func TestSubscribeGroupMessages_WithCursor(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))

	// Initial message before stream starts.
	mockValidateGroupMessages(mlsValidationService, groupId)
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
					Id:        uint64(i + 4),
					CreatedNs: uint64(i + 4),
					GroupId:   groupId,
					Data:      []byte(fmt.Sprintf("data%d", i+4)),
				},
			},
		}
	}

	stream := mocks.NewMockMlsApi_SubscribeGroupMessagesServer(t)
	stream.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	stream.EXPECT().Send(mock.MatchedBy(newGroupMessageIdAndDataEqualsMatcher(&mlsv1.GroupMessage{
		Version: &mlsv1.GroupMessage_V1_{
			V1: &mlsv1.GroupMessage_V1{
				Id:   3,
				Data: []byte("data3"),
			},
		},
	}).Matches)).Return(nil).Times(1)
	for _, msg := range msgs {
		stream.EXPECT().Send(mock.MatchedBy(newGroupMessageEqualsMatcher(msg).Matches)).Return(nil).Times(1)
	}
	stream.EXPECT().Context().Return(ctx)

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

	// Send the 10 real-time messages.
	for _, msg := range msgs {
		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)

		err = svc.HandleIncomingWakuRelayMessage(&wakupb.WakuMessage{
			ContentTopic: topic.BuildMLSV1GroupTopic(msg.GetV1().GroupId),
			Timestamp:    int64(msg.GetV1().CreatedNs),
			Payload:      msgB,
		})
		require.NoError(t, err)
	}

	// Expectations should eventually be satisfied.
	assertExpectationsWithTimeout(t, &stream.Mock, 5*time.Second, 100*time.Millisecond)
}

func TestSubscribeWelcomeMessages_WithoutCursor(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationKey := []byte(test.RandomString(32))

	msgs := make([]*mlsv1.WelcomeMessage, 10)
	for i := 0; i < 10; i++ {
		msgs[i] = &mlsv1.WelcomeMessage{
			Version: &mlsv1.WelcomeMessage_V1_{
				V1: &mlsv1.WelcomeMessage_V1{
					Id:              uint64(i + 1),
					CreatedNs:       uint64(i + 1),
					InstallationKey: installationKey,
					Data:            []byte(fmt.Sprintf("data%d", i+1)),
					HpkePublicKey:   []byte(fmt.Sprintf("hpke%d", i+1)),
				},
			},
		}
	}

	stream := mocks.NewMockMlsApi_SubscribeWelcomeMessagesServer(t)
	stream.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	for _, msg := range msgs {
		stream.EXPECT().Send(mock.MatchedBy(newWelcomeMessageEqualsMatcher(msg).Matches)).Return(nil).Times(1)
	}
	stream.EXPECT().Context().Return(ctx)

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

	for _, msg := range msgs {
		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)

		err = svc.HandleIncomingWakuRelayMessage(&wakupb.WakuMessage{
			ContentTopic: topic.BuildMLSV1WelcomeTopic(msg.GetV1().InstallationKey),
			Timestamp:    int64(msg.GetV1().CreatedNs),
			Payload:      msgB,
		})
		require.NoError(t, err)
	}

	assertExpectationsWithTimeout(t, &stream.Mock, 5*time.Second, 100*time.Millisecond)
}

func TestSubscribeWelcomeMessages_WithCursor(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationKey := []byte(test.RandomString(32))
	hpkePublicKey := []byte(test.RandomString(32))

	// Initial message before stream starts.
	initialMsgs := []*mlsv1.WelcomeMessageInput{
		{
			Version: &mlsv1.WelcomeMessageInput_V1_{
				V1: &mlsv1.WelcomeMessageInput_V1{
					InstallationKey: installationKey,
					HpkePublicKey:   hpkePublicKey,
					Data:            []byte("data1"),
				},
			},
		},
		{
			Version: &mlsv1.WelcomeMessageInput_V1_{
				V1: &mlsv1.WelcomeMessageInput_V1{
					InstallationKey: installationKey,
					HpkePublicKey:   hpkePublicKey,
					Data:            []byte("data2"),
				},
			},
		},
		{
			Version: &mlsv1.WelcomeMessageInput_V1_{
				V1: &mlsv1.WelcomeMessageInput_V1{
					InstallationKey: installationKey,
					HpkePublicKey:   hpkePublicKey,
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
					Id:              uint64(i + 4),
					CreatedNs:       uint64(i + 4),
					InstallationKey: installationKey,
					HpkePublicKey:   hpkePublicKey,
					Data:            []byte(fmt.Sprintf("data%d", i+4)),
				},
			},
		}
	}

	// Set up expectations of streaming the 11 messages from cursor.
	stream := mocks.NewMockMlsApi_SubscribeWelcomeMessagesServer(t)
	stream.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	stream.EXPECT().Send(mock.MatchedBy(newWelcomeMessageEqualsMatcherWithoutTimestamp(&mlsv1.WelcomeMessage{
		Version: &mlsv1.WelcomeMessage_V1_{
			V1: &mlsv1.WelcomeMessage_V1{
				Id:              3,
				InstallationKey: installationKey,
				HpkePublicKey:   hpkePublicKey,
				Data:            []byte("data3"),
			},
		},
	}).Matches)).Return(nil).Times(1)
	for _, msg := range msgs {
		stream.EXPECT().Send(mock.MatchedBy(newWelcomeMessageEqualsMatcher(msg).Matches)).Return(nil).Times(1)
	}
	stream.EXPECT().Context().Return(ctx)

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

	// Send the 10 real-time messages.
	for _, msg := range msgs {
		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)

		err = svc.HandleIncomingWakuRelayMessage(&wakupb.WakuMessage{
			ContentTopic: topic.BuildMLSV1WelcomeTopic(msg.GetV1().InstallationKey),
			Timestamp:    int64(msg.GetV1().CreatedNs),
			Payload:      msgB,
		})
		require.NoError(t, err)
	}

	// Expectations should eventually be satisfied.
	assertExpectationsWithTimeout(t, &stream.Mock, 5*time.Second, 100*time.Millisecond)
}

type groupMessageEqualsMatcher struct {
	obj *mlsv1.GroupMessage
}

func newGroupMessageEqualsMatcher(obj *mlsv1.GroupMessage) *groupMessageEqualsMatcher {
	return &groupMessageEqualsMatcher{obj}
}

func (m *groupMessageEqualsMatcher) Matches(obj interface{}) bool {
	return proto.Equal(m.obj, obj.(*mlsv1.GroupMessage))
}

func (m *groupMessageEqualsMatcher) String() string {
	return m.obj.String()
}

type groupMessageIdAndDataEqualsMatcher struct {
	obj *mlsv1.GroupMessage
}

func newGroupMessageIdAndDataEqualsMatcher(obj *mlsv1.GroupMessage) *groupMessageIdAndDataEqualsMatcher {
	return &groupMessageIdAndDataEqualsMatcher{obj}
}

func (m *groupMessageIdAndDataEqualsMatcher) Matches(obj interface{}) bool {
	return m.obj.GetV1().Id == obj.(*mlsv1.GroupMessage).GetV1().Id &&
		bytes.Equal(m.obj.GetV1().Data, obj.(*mlsv1.GroupMessage).GetV1().Data)
}

func (m *groupMessageIdAndDataEqualsMatcher) String() string {
	return m.obj.String()
}

type welcomeMessageEqualsMatcher struct {
	obj *mlsv1.WelcomeMessage
}

func newWelcomeMessageEqualsMatcher(obj *mlsv1.WelcomeMessage) *welcomeMessageEqualsMatcher {
	return &welcomeMessageEqualsMatcher{obj}
}

func (m *welcomeMessageEqualsMatcher) Matches(obj interface{}) bool {
	return proto.Equal(m.obj, obj.(*mlsv1.WelcomeMessage))
}

func (m *welcomeMessageEqualsMatcher) String() string {
	return m.obj.String()
}

type welcomeMessageEqualsMatcherWithoutTimestamp struct {
	obj *mlsv1.WelcomeMessage
}

func newWelcomeMessageEqualsMatcherWithoutTimestamp(obj *mlsv1.WelcomeMessage) *welcomeMessageEqualsMatcherWithoutTimestamp {
	return &welcomeMessageEqualsMatcherWithoutTimestamp{obj}
}

func (m *welcomeMessageEqualsMatcherWithoutTimestamp) Matches(obj interface{}) bool {
	return m.obj.GetV1().Id == obj.(*mlsv1.WelcomeMessage).GetV1().Id &&
		bytes.Equal(m.obj.GetV1().InstallationKey, obj.(*mlsv1.WelcomeMessage).GetV1().InstallationKey) &&
		bytes.Equal(m.obj.GetV1().HpkePublicKey, obj.(*mlsv1.WelcomeMessage).GetV1().HpkePublicKey) &&
		bytes.Equal(m.obj.GetV1().Data, obj.(*mlsv1.WelcomeMessage).GetV1().Data)
}

func (m *welcomeMessageEqualsMatcherWithoutTimestamp) String() string {
	return m.obj.String()
}

func assertExpectationsWithTimeout(t *testing.T, mockObj *mock.Mock, timeout, interval time.Duration) {
	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			if !mockObj.AssertExpectations(t) {
				t.Error("Expectations were not met within the timeout period")
			}
			return
		case <-ticker.C:
			if mockObj.AssertExpectations(t) {
				return
			}
		}
	}
}
