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
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"github.com/xmtp/xmtp-node-go/pkg/mocks"
	"github.com/xmtp/xmtp-node-go/pkg/proto/identity"
	"github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	messageApi "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	messageContentsProto "github.com/xmtp/xmtp-node-go/pkg/proto/mls/message_contents"
	"github.com/xmtp/xmtp-node-go/pkg/subscriptions"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func mockValidateInboxIdKeyPackages(
	m *mocks.MockMLSValidationService,
	installationId []byte,
	inboxId string,
) *mocks.MockMLSValidationService_ValidateInboxIdKeyPackages_Call {
	return m.EXPECT().
		ValidateInboxIdKeyPackages(mock.Anything, mock.Anything).
		Return([]mlsvalidate.InboxIdValidationResult{
			{
				InstallationKey: installationId,
				Credential: &identity.MlsCredential{
					InboxId: inboxId,
				},
				Expiration: 0,
			},
		}, nil)
}

func mockValidateGroupMessages(
	m *mocks.MockMLSValidationService,
	groupId []byte,
) *mocks.MockMLSValidationService_ValidateGroupMessages_Call {
	return m.EXPECT().
		ValidateGroupMessages(mock.Anything, mock.Anything).
		Return([]mlsvalidate.GroupMessageValidationResult{
			{
				GroupId: fmt.Sprintf("%x", groupId),
			},
		}, nil)
}

func newTestService(
	t *testing.T,
	ctx context.Context,
) (*Service, *bun.DB, *mocks.MockMLSValidationService, func()) {
	log := test.NewLog(t)
	db, _, mlsDbCleanup := test.NewMLSDB(t)
	store, err := mlsstore.New(ctx, log, db)
	require.NoError(t, err)
	mockMlsValidation := mocks.NewMockMLSValidationService(t)

	subDispatcher := subscriptions.NewSubscriptionDispatcher(log)

	svc, err := NewService(log, store, store, subDispatcher, mockMlsValidation, false, nil)
	require.NoError(t, err)

	return svc, db, mockMlsValidation, func() {
		svc.Close()
		mlsDbCleanup()
	}
}

func TestRegisterInstallation(t *testing.T) {
	ctx := context.Background()
	svc, mlsDb, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationId := test.RandomBytes(32)
	keyPackage := []byte("test")

	mockValidateInboxIdKeyPackages(mlsValidationService, installationId, test.RandomInboxId())

	res, err := svc.RegisterInstallation(ctx, &mlsv1.RegisterInstallationRequest{
		KeyPackage: &mlsv1.KeyPackageUpload{
			KeyPackageTlsSerialized: keyPackage,
		},
		IsInboxIdCredential: false,
	})

	require.NoError(t, err)
	require.Equal(t, installationId, res.InstallationKey)

	installation, err := queries.New(mlsDb.DB).GetInstallation(ctx, installationId)
	require.NoError(t, err)

	require.Equal(t, installationId, installation.ID)
	require.Equal(t, []byte("test"), installation.KeyPackage)
}

func TestRegisterInstallationError(t *testing.T) {
	ctx := context.Background()
	svc, _, mlsValidationService, cleanup := newTestService(t, ctx)
	defer cleanup()

	mlsValidationService.EXPECT().
		ValidateInboxIdKeyPackages(mock.Anything, mock.Anything).
		Return(nil, errors.New("error validating"))

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

	resp, err := svc.writerStore.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
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
						InstallationKey:  []byte(installationId),
						Data:             []byte("test"),
						HpkePublicKey:    []byte("test"),
						WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	resp, err := svc.writerStore.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: installationId,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, resp.Messages[0].GetV1().Data, []byte("test"))
	require.Equal(
		t,
		resp.Messages[0].GetV1().WrapperAlgorithm,
		messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
	)
	require.NotEmpty(t, resp.Messages[0].GetV1().CreatedNs)
}

func TestSendWelcomePointerMessages(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	installationKey := []byte(test.RandomString(32))
	welcomePointer := []byte(test.RandomString(32))
	hpkePublicKey := []byte(test.RandomString(32))

	_, err := svc.SendWelcomeMessages(ctx, &mlsv1.SendWelcomeMessagesRequest{
		Messages: []*mlsv1.WelcomeMessageInput{
			{
				Version: &mlsv1.WelcomeMessageInput_WelcomePointer_{
					WelcomePointer: &mlsv1.WelcomeMessageInput_WelcomePointer{
						InstallationKey:  installationKey,
						WelcomePointer:   welcomePointer,
						HpkePublicKey:    hpkePublicKey,
						WrapperAlgorithm: messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	resp, err := svc.writerStore.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: installationKey,
	})
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(resp.Messages), 1)

	// Find the welcome pointer message
	var welcomePointerMsg *mlsv1.WelcomeMessage
	for _, msg := range resp.Messages {
		if msg.GetWelcomePointer() != nil {
			welcomePointerMsg = msg
			break
		}
	}
	require.NotNil(t, welcomePointerMsg)
	require.Equal(t, welcomePointerMsg.GetWelcomePointer().InstallationKey, installationKey)
	require.Equal(t, welcomePointerMsg.GetWelcomePointer().WelcomePointer, welcomePointer)
	require.Equal(t, welcomePointerMsg.GetWelcomePointer().HpkePublicKey, hpkePublicKey)
	require.Equal(
		t,
		welcomePointerMsg.GetWelcomePointer().WrapperAlgorithm,
		messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
	)
	require.NotEmpty(t, welcomePointerMsg.GetWelcomePointer().CreatedNs)
}

func TestSendWelcomeMessagesConcurrent(t *testing.T) {
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
						Data:            []byte("test1"),
						HpkePublicKey:   []byte("test1"),
					},
				},
			},
			{
				Version: &mlsv1.WelcomeMessageInput_V1_{
					V1: &mlsv1.WelcomeMessageInput_V1{
						InstallationKey: []byte(installationId),
						Data:            []byte("test2"),
						HpkePublicKey:   []byte("test2"),
					},
				},
			},
		},
	})
	require.NoError(t, err)

	resp, err := svc.writerStore.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: installationId,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)

	hasMessage1 := false
	hasMessage2 := false
	for _, msg := range resp.Messages {
		if bytes.Equal(msg.GetV1().Data, []byte("test1")) {
			hasMessage1 = true
		}
		if bytes.Equal(msg.GetV1().Data, []byte("test2")) {
			hasMessage2 = true
		}
	}
	require.True(t, hasMessage1, "should have message with data 'test1'")
	require.True(t, hasMessage2, "should have message with data 'test2'")
}

func TestSubscribeGroupMessages_WithoutCursor(t *testing.T) {
	ctx := context.Background()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))

	mockValidateGroupMessages(validationSvc, groupId)

	inputs := make([]*mlsv1.GroupMessageInput, 10)
	for i := 0; i < 10; i++ {
		inputs[i] = &mlsv1.GroupMessageInput{
			Version: &mlsv1.GroupMessageInput_V1_{
				V1: &mlsv1.GroupMessageInput_V1{
					Data:       []byte(fmt.Sprintf("data%d", i+1)),
					SenderHmac: []byte(fmt.Sprintf("hmac%d", i+1)),
					ShouldPush: true,
				},
			},
		}
	}

	// Publish the first 5 messages before the stream starts
	for _, input := range inputs[:5] {
		_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
			Messages: []*mlsv1.GroupMessageInput{input},
		})
		require.NoError(t, err)
	}

	stream := mocks.NewMockMlsApi_SubscribeGroupMessagesServer(t)
	stream.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	allIDs := make(chan uint64, 10)
	for _, input := range inputs {
		input := input
		runner := func(msg *mlsv1.GroupMessage) {
			allIDs <- msg.GetV1().Id
		}
		stream.EXPECT().
			Send(mock.MatchedBy(newGroupMessageDataAndHmacEqualsMatcher(input).Matches)).
			Run(runner).
			Return(nil).
			Times(1)
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
		if err != nil {
			require.Error(t, err)
			require.Equal(t, codes.Unavailable, status.Code(err))
		}
	}()

	// Publish the last 5 messages after the stream is already running
	for _, input := range inputs[5:] {
		_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
			Messages: []*mlsv1.GroupMessageInput{input},
		})
		require.NoError(t, err)
	}

	assertExpectationsWithTimeout(t, &stream.Mock, 3*time.Second, 100*time.Millisecond)

	close(allIDs)

	highestId := uint64(0)
	numSent := 0
	for id := range allIDs {
		require.Greater(t, id, highestId)
		highestId = id
		numSent++
	}

	require.Equal(t, numSent, len(inputs))
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
					Data:       []byte("data1"),
					SenderHmac: []byte("hmac1"),
					ShouldPush: true,
				},
			},
		},
		{
			Version: &mlsv1.GroupMessageInput_V1_{
				V1: &mlsv1.GroupMessageInput_V1{
					Data:       []byte("data2"),
					SenderHmac: []byte("hmac2"),
					ShouldPush: true,
				},
			},
		},
		{
			Version: &mlsv1.GroupMessageInput_V1_{
				V1: &mlsv1.GroupMessageInput_V1{
					Data:       []byte("data3"),
					SenderHmac: []byte("hmac3"),
					ShouldPush: false,
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
					Id:         uint64(i + 4),
					CreatedNs:  uint64(i + 4),
					GroupId:    groupId,
					Data:       []byte(fmt.Sprintf("data%d", i+4)),
					SenderHmac: []byte(fmt.Sprintf("hmac%d", i+4)),
					ShouldPush: true,
				},
			},
		}
	}

	stream := mocks.NewMockMlsApi_SubscribeGroupMessagesServer(t)
	stream.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	stream.EXPECT().Send(mock.MatchedBy(newGroupMessageIdAndDataEqualsMatcher(&mlsv1.GroupMessage{
		Version: &mlsv1.GroupMessage_V1_{
			V1: &mlsv1.GroupMessage_V1{
				Id:         3,
				Data:       []byte("data3"),
				SenderHmac: []byte("hmac3"),
				ShouldPush: false,
			},
		},
	}).Matches)).Return(nil).Times(1)

	for _, msg := range msgs {
		stream.EXPECT().
			Send(mock.MatchedBy(newGroupMessageEqualsMatcher(msg).Matches)).
			Return(nil).
			Times(1)
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
		if err != nil {
			require.Error(t, err)
			require.Equal(t, codes.Unavailable, status.Code(err))
		}
	}()

	// Make sure that the stream has started before we send more messages
	time.Sleep(100 * time.Millisecond)

	// Send the 10 real-time messages.
	for _, msg := range msgs {
		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)

		svc.subDispatcher.HandleEnvelope(&messageApi.Envelope{
			ContentTopic: topic.BuildMLSV1GroupTopic(msg.GetV1().GroupId),
			TimestampNs:  uint64(msg.GetV1().CreatedNs),
			Message:      msgB,
		})
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
					Id:               uint64(i + 1),
					CreatedNs:        uint64(i + 1),
					InstallationKey:  installationKey,
					Data:             []byte(fmt.Sprintf("data%d", i+1)),
					HpkePublicKey:    []byte(fmt.Sprintf("hpke%d", i+1)),
					WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
				},
			},
		}
	}

	stream := mocks.NewMockMlsApi_SubscribeWelcomeMessagesServer(t)
	stream.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	for _, msg := range msgs {
		stream.EXPECT().
			Send(mock.MatchedBy(newWelcomeMessageEqualsMatcher(msg).Matches)).
			Return(nil).
			Times(1)
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
		if err != nil {
			require.Error(t, err)
			require.Equal(t, codes.Unavailable, status.Code(err))
		}
	}()
	// Make sure that the stream has started before we send more messages
	time.Sleep(100 * time.Millisecond)

	for _, msg := range msgs {
		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)

		svc.subDispatcher.HandleEnvelope(&messageApi.Envelope{
			ContentTopic: topic.BuildMLSV1WelcomeTopic(msg.GetV1().InstallationKey),
			TimestampNs:  uint64(msg.GetV1().CreatedNs),
			Message:      msgB,
		})
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
					InstallationKey:  installationKey,
					HpkePublicKey:    hpkePublicKey,
					Data:             []byte("data1"),
					WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
				},
			},
		},
		{
			Version: &mlsv1.WelcomeMessageInput_V1_{
				V1: &mlsv1.WelcomeMessageInput_V1{
					InstallationKey:  installationKey,
					HpkePublicKey:    hpkePublicKey,
					Data:             []byte("data2"),
					WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519,
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
					Id:               uint64(i + 4),
					CreatedNs:        uint64(i + 4),
					InstallationKey:  installationKey,
					HpkePublicKey:    hpkePublicKey,
					Data:             []byte(fmt.Sprintf("data%d", i+4)),
					WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
				},
			},
		}
	}

	// Set up expectations of streaming the 11 messages from cursor.
	stream := mocks.NewMockMlsApi_SubscribeWelcomeMessagesServer(t)
	stream.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	stream.EXPECT().
		Send(mock.MatchedBy(newWelcomeMessageEqualsMatcherWithoutTimestamp(&mlsv1.WelcomeMessage{
			Version: &mlsv1.WelcomeMessage_V1_{
				V1: &mlsv1.WelcomeMessage_V1{
					Id:               3,
					InstallationKey:  installationKey,
					HpkePublicKey:    hpkePublicKey,
					Data:             []byte("data3"),
					WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519,
				},
			},
		}).Matches)).
		Return(nil).
		Times(1)
	for _, msg := range msgs {
		stream.EXPECT().
			Send(mock.MatchedBy(newWelcomeMessageEqualsMatcher(msg).Matches)).
			Return(nil).
			Times(1)
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
		if err != nil {
			require.Error(t, err)
			require.Equal(t, codes.Unavailable, status.Code(err))
		}
	}()
	// Make sure that the stream has started before we send more messages
	time.Sleep(100 * time.Millisecond)

	// Send the 10 real-time messages.
	for _, msg := range msgs {
		msgB, err := proto.Marshal(msg)
		require.NoError(t, err)

		svc.subDispatcher.HandleEnvelope(&messageApi.Envelope{
			ContentTopic: topic.BuildMLSV1WelcomeTopic(msg.GetV1().InstallationKey),
			TimestampNs:  uint64(msg.GetV1().CreatedNs),
			Message:      msgB,
		})
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
	return proto.Equal(m.obj, obj.(*mlsv1.GroupMessage)) &&
		bytes.Equal(m.obj.GetV1().SenderHmac, obj.(*mlsv1.GroupMessage).GetV1().SenderHmac) &&
		m.obj.GetV1().ShouldPush == obj.(*mlsv1.GroupMessage).GetV1().ShouldPush
}

func (m *groupMessageEqualsMatcher) String() string {
	return m.obj.String()
}

type groupMessageDataAndHmacEqualsMatcher struct {
	obj *mlsv1.GroupMessageInput
}

func newGroupMessageDataAndHmacEqualsMatcher(
	obj *mlsv1.GroupMessageInput,
) *groupMessageDataAndHmacEqualsMatcher {
	return &groupMessageDataAndHmacEqualsMatcher{obj}
}

func (m *groupMessageDataAndHmacEqualsMatcher) Matches(obj interface{}) bool {
	return bytes.Equal(m.obj.GetV1().Data, obj.(*mlsv1.GroupMessage).GetV1().Data) &&
		bytes.Equal(m.obj.GetV1().SenderHmac, obj.(*mlsv1.GroupMessage).GetV1().SenderHmac)
}

func (m *groupMessageDataAndHmacEqualsMatcher) String() string {
	return m.obj.String()
}

type groupMessageIdAndDataEqualsMatcher struct {
	obj *mlsv1.GroupMessage
}

func newGroupMessageIdAndDataEqualsMatcher(
	obj *mlsv1.GroupMessage,
) *groupMessageIdAndDataEqualsMatcher {
	return &groupMessageIdAndDataEqualsMatcher{obj}
}

func (m *groupMessageIdAndDataEqualsMatcher) Matches(obj interface{}) bool {
	return m.obj.GetV1().Id == obj.(*mlsv1.GroupMessage).GetV1().Id &&
		bytes.Equal(m.obj.GetV1().Data, obj.(*mlsv1.GroupMessage).GetV1().Data) &&
		bytes.Equal(m.obj.GetV1().SenderHmac, obj.(*mlsv1.GroupMessage).GetV1().SenderHmac) &&
		m.obj.GetV1().ShouldPush == obj.(*mlsv1.GroupMessage).GetV1().ShouldPush
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

func newWelcomeMessageEqualsMatcherWithoutTimestamp(
	obj *mlsv1.WelcomeMessage,
) *welcomeMessageEqualsMatcherWithoutTimestamp {
	return &welcomeMessageEqualsMatcherWithoutTimestamp{obj}
}

func (m *welcomeMessageEqualsMatcherWithoutTimestamp) Matches(obj interface{}) bool {
	return m.obj.GetV1().Id == obj.(*mlsv1.WelcomeMessage).GetV1().Id &&
		bytes.Equal(
			m.obj.GetV1().InstallationKey,
			obj.(*mlsv1.WelcomeMessage).GetV1().InstallationKey,
		) &&
		bytes.Equal(
			m.obj.GetV1().HpkePublicKey,
			obj.(*mlsv1.WelcomeMessage).GetV1().HpkePublicKey,
		) &&
		bytes.Equal(m.obj.GetV1().Data, obj.(*mlsv1.WelcomeMessage).GetV1().Data)
}

func (m *welcomeMessageEqualsMatcherWithoutTimestamp) String() string {
	return m.obj.String()
}

func assertExpectationsWithTimeout(
	t *testing.T,
	mockObj *mock.Mock,
	timeout, interval time.Duration,
) {
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

func randomSignature() *associations.RecoverableEd25519Signature {
	return &associations.RecoverableEd25519Signature{
		Bytes:     []byte(test.RandomString(32)),
		PublicKey: []byte(test.RandomString(32)),
	}
}

func TestBatchPublishCommitLog(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))
	serializedEntry := []byte(test.RandomString(64))
	signature := randomSignature()

	_, err := svc.BatchPublishCommitLog(ctx, &mlsv1.BatchPublishCommitLogRequest{
		Requests: []*mlsv1.PublishCommitLogRequest{
			{
				GroupId:                  groupId,
				SerializedCommitLogEntry: serializedEntry,
				Signature:                signature,
			},
		},
	})
	require.NoError(t, err)

	// Verify the commit log entry was stored
	resp, err := svc.writerStore.QueryCommitLog(ctx, &mlsv1.QueryCommitLogRequest{
		GroupId: groupId,
	})
	require.NoError(t, err)
	require.Len(t, resp.CommitLogEntries, 1)
	require.Equal(t, serializedEntry, resp.CommitLogEntries[0].SerializedCommitLogEntry)
	require.Equal(t, signature.Bytes, resp.CommitLogEntries[0].Signature.Bytes)
	require.Equal(t, signature.PublicKey, resp.CommitLogEntries[0].Signature.PublicKey)
	require.Equal(t, uint64(1), resp.CommitLogEntries[0].SequenceId)
}

func TestBatchPublishCommitLog_MultipleEntries(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId1 := []byte(test.RandomString(32))
	groupId2 := []byte(test.RandomString(32))
	serializedEntry1 := []byte(test.RandomString(64))
	serializedEntry2 := []byte(test.RandomString(64))
	signature1 := randomSignature()
	signature2 := randomSignature()

	_, err := svc.BatchPublishCommitLog(ctx, &mlsv1.BatchPublishCommitLogRequest{
		Requests: []*mlsv1.PublishCommitLogRequest{
			{
				GroupId:                  groupId1,
				SerializedCommitLogEntry: serializedEntry1,
				Signature:                signature1,
			},
			{
				GroupId:                  groupId2,
				SerializedCommitLogEntry: serializedEntry2,
				Signature:                signature2,
			},
		},
	})
	require.NoError(t, err)

	// Verify both commit log entries were stored
	resp1, err := svc.writerStore.QueryCommitLog(ctx, &mlsv1.QueryCommitLogRequest{
		GroupId: groupId1,
	})
	require.NoError(t, err)
	require.Len(t, resp1.CommitLogEntries, 1)
	require.Equal(t, serializedEntry1, resp1.CommitLogEntries[0].SerializedCommitLogEntry)
	require.Equal(t, signature1.Bytes, resp1.CommitLogEntries[0].Signature.Bytes)
	require.Equal(t, signature1.PublicKey, resp1.CommitLogEntries[0].Signature.PublicKey)

	resp2, err := svc.writerStore.QueryCommitLog(ctx, &mlsv1.QueryCommitLogRequest{
		GroupId: groupId2,
	})
	require.NoError(t, err)
	require.Len(t, resp2.CommitLogEntries, 1)
	require.Equal(t, serializedEntry2, resp2.CommitLogEntries[0].SerializedCommitLogEntry)
	require.Equal(t, signature2.Bytes, resp2.CommitLogEntries[0].Signature.Bytes)
	require.Equal(t, signature2.PublicKey, resp2.CommitLogEntries[0].Signature.PublicKey)
}

func TestBatchPublishCommitLog_InvalidRequest(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	// Test with nil entry
	_, err := svc.BatchPublishCommitLog(ctx, &mlsv1.BatchPublishCommitLogRequest{
		Requests: []*mlsv1.PublishCommitLogRequest{nil},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid commit log entry")

	// Test with empty group ID
	_, err = svc.BatchPublishCommitLog(ctx, &mlsv1.BatchPublishCommitLogRequest{
		Requests: []*mlsv1.PublishCommitLogRequest{
			{
				GroupId:                  []byte{},
				SerializedCommitLogEntry: []byte("test"),
				Signature:                randomSignature(),
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid commit log entry")

	// Test with empty encrypted entry
	_, err = svc.BatchPublishCommitLog(ctx, &mlsv1.BatchPublishCommitLogRequest{
		Requests: []*mlsv1.PublishCommitLogRequest{
			{
				GroupId:                  []byte("test"),
				SerializedCommitLogEntry: []byte{},
				Signature:                randomSignature(),
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid commit log entry")
}

func TestBatchQueryCommitLog(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId1 := []byte(test.RandomString(32))
	groupId2 := []byte(test.RandomString(32))
	serializedEntry1 := []byte(test.RandomString(64))
	serializedEntry2 := []byte(test.RandomString(64))
	signature1 := randomSignature()
	signature2 := randomSignature()

	// Insert commit log entries
	_, err := svc.BatchPublishCommitLog(ctx, &mlsv1.BatchPublishCommitLogRequest{
		Requests: []*mlsv1.PublishCommitLogRequest{
			{
				GroupId:                  groupId1,
				SerializedCommitLogEntry: serializedEntry1,
				Signature:                signature1,
			},
			{
				GroupId:                  groupId2,
				SerializedCommitLogEntry: serializedEntry2,
				Signature:                signature2,
			},
		},
	})
	require.NoError(t, err)

	// Query both groups
	resp, err := svc.BatchQueryCommitLog(ctx, &mlsv1.BatchQueryCommitLogRequest{
		Requests: []*mlsv1.QueryCommitLogRequest{
			{
				GroupId: groupId1,
			},
			{
				GroupId: groupId2,
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Responses, 2)

	// Verify first group response
	require.Equal(t, groupId1, resp.Responses[0].GroupId)
	require.Len(t, resp.Responses[0].CommitLogEntries, 1)
	require.Equal(
		t,
		serializedEntry1,
		resp.Responses[0].CommitLogEntries[0].SerializedCommitLogEntry,
	)
	require.Equal(t, signature1.Bytes, resp.Responses[0].CommitLogEntries[0].Signature.Bytes)
	require.Equal(
		t,
		signature1.PublicKey,
		resp.Responses[0].CommitLogEntries[0].Signature.PublicKey,
	)
	require.Equal(t, uint64(1), resp.Responses[0].CommitLogEntries[0].SequenceId)

	// Verify second group response
	require.Equal(t, groupId2, resp.Responses[1].GroupId)
	require.Len(t, resp.Responses[1].CommitLogEntries, 1)
	require.Equal(
		t,
		serializedEntry2,
		resp.Responses[1].CommitLogEntries[0].SerializedCommitLogEntry,
	)
	require.Equal(t, signature2.Bytes, resp.Responses[1].CommitLogEntries[0].Signature.Bytes)
	require.Equal(
		t,
		signature2.PublicKey,
		resp.Responses[1].CommitLogEntries[0].Signature.PublicKey,
	)
	require.Equal(t, uint64(2), resp.Responses[1].CommitLogEntries[0].SequenceId)
}

func TestBatchQueryCommitLog_WithPaging(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))
	serializedEntry := []byte("entry")

	// Insert 5 commit log entries
	_, err := svc.BatchPublishCommitLog(ctx, &mlsv1.BatchPublishCommitLogRequest{
		Requests: []*mlsv1.PublishCommitLogRequest{
			{
				GroupId:                  groupId,
				SerializedCommitLogEntry: serializedEntry,
				Signature:                randomSignature(),
			},
			{
				GroupId:                  groupId,
				SerializedCommitLogEntry: serializedEntry,
				Signature:                randomSignature(),
			},
			{
				GroupId:                  groupId,
				SerializedCommitLogEntry: serializedEntry,
				Signature:                randomSignature(),
			},
			{
				GroupId:                  groupId,
				SerializedCommitLogEntry: serializedEntry,
				Signature:                randomSignature(),
			},
			{
				GroupId:                  groupId,
				SerializedCommitLogEntry: serializedEntry,
				Signature:                randomSignature(),
			},
		},
	})
	require.NoError(t, err)

	// Query with limit
	resp, err := svc.BatchQueryCommitLog(ctx, &mlsv1.BatchQueryCommitLogRequest{
		Requests: []*mlsv1.QueryCommitLogRequest{
			{
				GroupId: groupId,
				PagingInfo: &mlsv1.PagingInfo{
					Limit: 3,
				},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Responses, 1)
	require.Len(t, resp.Responses[0].CommitLogEntries, 3)

	// Verify paging info
	require.NotNil(t, resp.Responses[0].PagingInfo)
	require.Equal(t, uint32(3), resp.Responses[0].PagingInfo.Limit)
	require.Equal(t, uint64(3), resp.Responses[0].PagingInfo.IdCursor)
	require.Equal(
		t,
		mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		resp.Responses[0].PagingInfo.Direction,
	)

	// Query next page
	resp2, err := svc.BatchQueryCommitLog(ctx, &mlsv1.BatchQueryCommitLogRequest{
		Requests: []*mlsv1.QueryCommitLogRequest{
			{
				GroupId: groupId,
				PagingInfo: &mlsv1.PagingInfo{
					Limit:    3,
					IdCursor: resp.Responses[0].PagingInfo.IdCursor,
				},
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, resp2.Responses, 1)
	require.Len(t, resp2.Responses[0].CommitLogEntries, 2) // Remaining 2 entries
}

func TestBatchQueryCommitLog_EmptyGroup(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	unknownGroupId := []byte(test.RandomString(32))

	resp, err := svc.BatchQueryCommitLog(ctx, &mlsv1.BatchQueryCommitLogRequest{
		Requests: []*mlsv1.QueryCommitLogRequest{
			{
				GroupId: unknownGroupId,
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Responses, 1)
	require.Equal(t, unknownGroupId, resp.Responses[0].GroupId)
	require.Len(t, resp.Responses[0].CommitLogEntries, 0)
}

func TestBatchQueryCommitLog_Error(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	resp, err := svc.BatchQueryCommitLog(ctx, &mlsv1.BatchQueryCommitLogRequest{
		Requests: []*mlsv1.QueryCommitLogRequest{
			{
				GroupId: []byte{},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid request")
	require.Nil(t, resp)

	groupId := []byte(test.RandomString(32))
	resp, err = svc.BatchQueryCommitLog(ctx, &mlsv1.BatchQueryCommitLogRequest{
		Requests: []*mlsv1.QueryCommitLogRequest{
			{
				GroupId: groupId,
				PagingInfo: &mlsv1.PagingInfo{
					Direction: mlsv1.SortDirection_SORT_DIRECTION_DESCENDING,
				},
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "descending direction is not supported")
	require.Nil(t, resp)
}

func TestValidateSendWelcomeMessagesRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     *mlsv1.SendWelcomeMessagesRequest
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil request",
			request:     nil,
			expectError: true,
			errorMsg:    "no welcome messages to send",
		},
		{
			name: "empty messages",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{},
			},
			expectError: true,
			errorMsg:    "no welcome messages to send",
		},
		{
			name: "nil message input",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{nil},
			},
			expectError: true,
			errorMsg:    "invalid welcome message",
		},
		{
			name: "missing version",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: nil,
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid welcome message: missing version",
		},
		// V1 Welcome Message Tests
		{
			name: "valid V1 welcome message with HPKE key",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_V1_{
							V1: &mlsv1.WelcomeMessageInput_V1{
								InstallationKey:  []byte("installation-key"),
								Data:             []byte("welcome-data"),
								HpkePublicKey:    []byte("hpke-public-key"),
								WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519,
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid V1 welcome message with symmetric key algorithm",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_V1_{
							V1: &mlsv1.WelcomeMessageInput_V1{
								InstallationKey:  []byte("installation-key"),
								Data:             []byte("welcome-data"),
								HpkePublicKey:    []byte{}, // Empty HPKE key is allowed for symmetric key
								WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_SYMMETRIC_KEY,
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid V1 welcome message - empty data",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_V1_{
							V1: &mlsv1.WelcomeMessageInput_V1{
								InstallationKey:  []byte("installation-key"),
								Data:             []byte{}, // Empty data
								HpkePublicKey:    []byte("hpke-public-key"),
								WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519,
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid welcome message",
		},
		{
			name: "invalid V1 welcome message - empty installation key",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_V1_{
							V1: &mlsv1.WelcomeMessageInput_V1{
								InstallationKey:  []byte{}, // Empty installation key
								Data:             []byte("welcome-data"),
								HpkePublicKey:    []byte("hpke-public-key"),
								WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519,
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid welcome message",
		},
		{
			name: "invalid V1 welcome message - empty HPKE key with non-symmetric algorithm",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_V1_{
							V1: &mlsv1.WelcomeMessageInput_V1{
								InstallationKey:  []byte("installation-key"),
								Data:             []byte("welcome-data"),
								HpkePublicKey:    []byte{},                                                                          // Empty HPKE key
								WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519, // Non-symmetric algorithm
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid welcome message",
		},
		// Welcome Pointer Tests
		{
			name: "valid welcome pointer message",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_WelcomePointer_{
							WelcomePointer: &mlsv1.WelcomeMessageInput_WelcomePointer{
								InstallationKey:  []byte("installation-key"),
								WelcomePointer:   []byte("welcome-pointer-data"),
								HpkePublicKey:    []byte("hpke-public-key"),
								WrapperAlgorithm: messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid welcome pointer message - empty installation key",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_WelcomePointer_{
							WelcomePointer: &mlsv1.WelcomeMessageInput_WelcomePointer{
								InstallationKey:  []byte{}, // Empty installation key
								WelcomePointer:   []byte("welcome-pointer-data"),
								HpkePublicKey:    []byte("hpke-public-key"),
								WrapperAlgorithm: messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid welcome pointer message",
		},
		{
			name: "invalid welcome pointer message - empty welcome pointer",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_WelcomePointer_{
							WelcomePointer: &mlsv1.WelcomeMessageInput_WelcomePointer{
								InstallationKey:  []byte("installation-key"),
								WelcomePointer:   []byte{}, // Empty welcome pointer
								HpkePublicKey:    []byte("hpke-public-key"),
								WrapperAlgorithm: messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid welcome pointer message",
		},
		{
			name: "invalid welcome pointer message - empty HPKE public key",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_WelcomePointer_{
							WelcomePointer: &mlsv1.WelcomeMessageInput_WelcomePointer{
								InstallationKey:  []byte("installation-key"),
								WelcomePointer:   []byte("welcome-pointer-data"),
								HpkePublicKey:    []byte{}, // Empty HPKE public key
								WrapperAlgorithm: messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid welcome pointer message",
		},
		// Mixed message types
		{
			name: "valid mixed V1 and welcome pointer messages",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_V1_{
							V1: &mlsv1.WelcomeMessageInput_V1{
								InstallationKey:  []byte("installation-key-1"),
								Data:             []byte("welcome-data-1"),
								HpkePublicKey:    []byte("hpke-public-key-1"),
								WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519,
							},
						},
					},
					{
						Version: &mlsv1.WelcomeMessageInput_WelcomePointer_{
							WelcomePointer: &mlsv1.WelcomeMessageInput_WelcomePointer{
								InstallationKey:  []byte("installation-key-2"),
								WelcomePointer:   []byte("welcome-pointer-data-2"),
								HpkePublicKey:    []byte("hpke-public-key-2"),
								WrapperAlgorithm: messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid mixed messages - one valid, one invalid",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{
					{
						Version: &mlsv1.WelcomeMessageInput_V1_{
							V1: &mlsv1.WelcomeMessageInput_V1{
								InstallationKey:  []byte("installation-key-1"),
								Data:             []byte("welcome-data-1"),
								HpkePublicKey:    []byte("hpke-public-key-1"),
								WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_CURVE25519,
							},
						},
					},
					{
						Version: &mlsv1.WelcomeMessageInput_WelcomePointer_{
							WelcomePointer: &mlsv1.WelcomeMessageInput_WelcomePointer{
								InstallationKey:  []byte{}, // Invalid - empty installation key
								WelcomePointer:   []byte("welcome-pointer-data-2"),
								HpkePublicKey:    []byte("hpke-public-key-2"),
								WrapperAlgorithm: messageContentsProto.WelcomePointerWrapperAlgorithm_WELCOME_POINTER_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid welcome pointer message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSendWelcomeMessagesRequest(tt.request)
			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
