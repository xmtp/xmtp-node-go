package store

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	testutils "github.com/xmtp/xmtp-node-go/pkg/testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"github.com/xmtp/xmtp-node-go/pkg/mocks"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/types"
	"go.uber.org/zap"
)

func NewTestStore(t *testing.T) (*Store, func()) {
	log := testutils.NewLog(t)
	db, _, dbCleanup := testutils.NewMLSDB(t)
	ctx := context.Background()

	store, err := New(ctx, log, db)
	require.NoError(t, err)

	return store, dbCleanup
}

func TestPublishIdentityUpdateParallel(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create a mapping of inboxes to addresses
	inboxes := make(map[string]string)
	for i := 0; i < 50; i++ {
		inboxes[testutils.RandomInboxId()] = fmt.Sprintf("address_%d", i)
	}

	mockMlsValidation := mocks.NewMockMLSValidationService(t)

	// For each inbox_id in the map, return an AssociationStateDiff that adds the corresponding address
	mockMlsValidation.EXPECT().
		GetAssociationState(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, _ []*associations.IdentityUpdate, updates []*associations.IdentityUpdate) (*mlsvalidate.AssociationStateResult, error) {
			inboxId := updates[0].InboxId
			address, ok := inboxes[inboxId]

			if !ok {
				return nil, errors.New("inbox id not found")
			}

			return &mlsvalidate.AssociationStateResult{
				AssociationState: &associations.AssociationState{
					InboxId: inboxId,
				},
				StateDiff: &associations.AssociationStateDiff{
					NewMembers: []*associations.MemberIdentifier{{
						Kind: &associations.MemberIdentifier_EthereumAddress{
							EthereumAddress: address,
						},
					}},
				},
			}, nil
		})

	var wg sync.WaitGroup
	for inboxId := range inboxes {
		inboxId := inboxId
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.PublishIdentityUpdate(ctx, &identity.PublishIdentityUpdateRequest{
				IdentityUpdate: &associations.IdentityUpdate{
					InboxId: inboxId,
				},
			}, mockMlsValidation)
			require.NoError(t, err)
		}()
	}
	wg.Wait()
}

func TestPublishIdentityUpdateSameInboxParallel(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()
	ctx := context.Background()

	inboxId := testutils.RandomInboxId()
	address := testutils.RandomString(32)
	numUpdates := 50

	mockMlsValidation := mocks.NewMockMLSValidationService(t)

	// For each inbox_id in the map, return an AssociationStateDiff that adds the corresponding address
	mockMlsValidation.EXPECT().
		GetAssociationState(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, oldUpdates []*associations.IdentityUpdate, updates []*associations.IdentityUpdate) (*mlsvalidate.AssociationStateResult, error) {
			if len(oldUpdates) > 0 {
				return nil, errors.New("old updates should be empty")
			}

			return &mlsvalidate.AssociationStateResult{
				AssociationState: &associations.AssociationState{
					InboxId: inboxId,
				},
				StateDiff: &associations.AssociationStateDiff{
					NewMembers: []*associations.MemberIdentifier{{
						Kind: &associations.MemberIdentifier_EthereumAddress{
							EthereumAddress: address,
						},
					}},
				},
			}, nil
		})

	var wg sync.WaitGroup
	numErrors := int32(0)
	numSuccesses := int32(0)
	for i := 0; i < numUpdates; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.PublishIdentityUpdate(ctx, &identity.PublishIdentityUpdateRequest{
				IdentityUpdate: &associations.IdentityUpdate{
					InboxId: inboxId,
				},
			}, mockMlsValidation)
			if err != nil {
				atomic.AddInt32(&numErrors, 1)
			} else {
				atomic.AddInt32(&numSuccesses, 1)
			}
		}()
	}
	wg.Wait()
	require.Equal(t, int32(numUpdates), numSuccesses+numErrors)
	// We expect all but one to fail if the old updates array isn't empty
	require.Equal(t, numErrors, int32(numUpdates-1))
}

func TestInboxIds(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()
	ctx := context.Background()

	inbox1 := testutils.RandomInboxId()
	inbox2 := testutils.RandomInboxId()
	correctInbox := testutils.RandomInboxId()
	correctInbox2 := testutils.RandomInboxId()

	_, err := store.queries.InsertAddressLog(
		ctx,
		queries.InsertAddressLogParams{
			Address:               "address",
			InboxID:               inbox1,
			AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 1},
			RevocationSequenceID:  sql.NullInt64{Valid: false},
		},
	)
	require.NoError(t, err)
	_, err = store.queries.InsertAddressLog(
		ctx,
		queries.InsertAddressLogParams{
			Address:               "address",
			InboxID:               inbox1,
			AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 2},
			RevocationSequenceID:  sql.NullInt64{Valid: false},
		},
	)
	require.NoError(t, err)
	_, err = store.queries.InsertAddressLog(
		ctx,
		queries.InsertAddressLogParams{
			Address:               "address",
			InboxID:               inbox1,
			AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 3},
			RevocationSequenceID:  sql.NullInt64{Valid: false},
		},
	)
	require.NoError(t, err)
	_, err = store.queries.InsertAddressLog(
		ctx,
		queries.InsertAddressLogParams{
			Address:               "address",
			InboxID:               correctInbox,
			AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 4},
			RevocationSequenceID:  sql.NullInt64{Valid: false},
		},
	)
	require.NoError(t, err)

	reqs := make([]*identity.GetInboxIdsRequest_Request, 0)
	reqs = append(reqs, &identity.GetInboxIdsRequest_Request{
		Identifier:     "address",
		IdentifierKind: associations.IdentifierKind_IDENTIFIER_KIND_ETHEREUM,
	})
	req := &identity.GetInboxIdsRequest{
		Requests: reqs,
	}
	resp, _ := store.GetInboxIds(context.Background(), req)

	require.Equal(t, correctInbox, *resp.Responses[0].InboxId)

	_, err = store.queries.InsertAddressLog(
		ctx,
		queries.InsertAddressLogParams{
			Address:               "address",
			InboxID:               correctInbox2,
			AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 5},
			RevocationSequenceID:  sql.NullInt64{Valid: false},
		},
	)
	require.NoError(t, err)
	resp, _ = store.GetInboxIds(context.Background(), req)
	require.Equal(t, correctInbox2, *resp.Responses[0].InboxId)

	reqs = append(reqs, &identity.GetInboxIdsRequest_Request{
		Identifier:     "address2",
		IdentifierKind: associations.IdentifierKind_IDENTIFIER_KIND_ETHEREUM,
	})
	req = &identity.GetInboxIdsRequest{
		Requests: reqs,
	}
	_, err = store.queries.InsertAddressLog(
		ctx,
		queries.InsertAddressLogParams{
			Address:               "address2",
			InboxID:               inbox2,
			AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 8},
			RevocationSequenceID:  sql.NullInt64{Valid: false},
		},
	)
	require.NoError(t, err)
	resp, _ = store.GetInboxIds(context.Background(), req)
	require.Equal(t, inbox2, *resp.Responses[1].InboxId)
}

func TestMultipleInboxIds(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()
	ctx := context.Background()

	inbox1 := testutils.RandomInboxId()
	inbox2 := testutils.RandomInboxId()

	_, err := store.queries.InsertAddressLog(
		ctx,
		queries.InsertAddressLogParams{
			Address:               "address_1",
			InboxID:               inbox1,
			AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 1},
			RevocationSequenceID:  sql.NullInt64{Valid: false},
		},
	)
	require.NoError(t, err)
	_, err = store.queries.InsertAddressLog(
		ctx,
		queries.InsertAddressLogParams{
			Address:               "address_2",
			InboxID:               inbox2,
			AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 2},
			RevocationSequenceID:  sql.NullInt64{Valid: false},
		},
	)
	require.NoError(t, err)

	reqs := make([]*identity.GetInboxIdsRequest_Request, 0)
	reqs = append(reqs, &identity.GetInboxIdsRequest_Request{
		Identifier:     "address_1",
		IdentifierKind: associations.IdentifierKind_IDENTIFIER_KIND_ETHEREUM,
	})
	reqs = append(reqs, &identity.GetInboxIdsRequest_Request{
		Identifier:     "address_2",
		IdentifierKind: associations.IdentifierKind_IDENTIFIER_KIND_ETHEREUM,
	})
	req := &identity.GetInboxIdsRequest{
		Requests: reqs,
	}
	resp, _ := store.GetInboxIds(context.Background(), req)
	require.Equal(t, inbox1, *resp.Responses[0].InboxId)
	require.Equal(t, inbox2, *resp.Responses[1].InboxId)
}

func TestCreateInstallation(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := testutils.RandomBytes(32)

	err := store.CreateOrUpdateInstallation(ctx, installationId, testutils.RandomBytes(32))
	require.NoError(t, err)

	installationFromDb, err := store.queries.GetInstallation(ctx, installationId)
	require.NoError(t, err)
	require.Equal(t, installationFromDb.ID, installationId)
}

func TestUpdateKeyPackage(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := testutils.RandomBytes(32)
	keyPackage := testutils.RandomBytes(32)

	err := store.CreateOrUpdateInstallation(ctx, installationId, keyPackage)
	require.NoError(t, err)
	afterCreate, err := store.queries.GetInstallation(ctx, installationId)
	require.NoError(t, err)

	keyPackage2 := testutils.RandomBytes(32)
	err = store.CreateOrUpdateInstallation(ctx, installationId, keyPackage2)
	require.NoError(t, err)

	installationFromDb, err := store.queries.GetInstallation(ctx, installationId)
	require.NoError(t, err)

	require.Equal(t, keyPackage2, installationFromDb.KeyPackage)
	require.Greater(t, installationFromDb.UpdatedAt, afterCreate.UpdatedAt)
	require.Equal(t, installationFromDb.CreatedAt, afterCreate.CreatedAt)
}

func TestConsumeLastResortKeyPackage(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := testutils.RandomBytes(32)
	keyPackage := testutils.RandomBytes(32)

	err := store.CreateOrUpdateInstallation(ctx, installationId, keyPackage)
	require.NoError(t, err)

	fetchResult, err := store.FetchKeyPackages(ctx, [][]byte{installationId})
	require.NoError(t, err)
	require.Len(t, fetchResult, 1)
	require.Equal(t, keyPackage, fetchResult[0].KeyPackage)
	require.Equal(t, installationId, fetchResult[0].ID)
}

func TestInsertGroupMessage_Single(t *testing.T) {
	started := time.Now().UTC().Add(-time.Minute)
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertGroupMessage(
		ctx,
		[]byte("group"),
		[]byte("data"),
		[]byte("sender_hmac"),
		true,
		true,
	)
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, int64(1), msg.ID)
	store.log.Info("Created at", zap.Time("created_at", msg.CreatedAt))
	require.True(
		t,
		msg.CreatedAt.Before(time.Now().UTC().Add(1*time.Minute)) && msg.CreatedAt.After(started),
	)
	require.Equal(t, []byte("group"), msg.GroupID)
	require.Equal(t, []byte("data"), msg.Data)
	require.Equal(t, []byte("sender_hmac"), msg.SenderHmac)
	require.True(t, msg.IsCommit.Bool)
	require.True(t, msg.ShouldPush.Bool)

	msgs, err := store.queries.GetAllGroupMessages(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, *msg, msgs[0])
}

func TestInsertGroupMessage_Duplicate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertGroupMessage(
		ctx,
		[]byte("group"),
		[]byte("data"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, msg)

	msg, err = store.InsertGroupMessage(
		ctx,
		[]byte("group"),
		[]byte("data"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.Nil(t, msg)
	require.IsType(t, &AlreadyExistsError{}, err)
	require.True(t, IsAlreadyExistsError(err))
}

func TestInsertGroupMessage_ManyOrderedByTime(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertGroupMessage(
		ctx,
		[]byte("group"),
		[]byte("data1"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group"),
		[]byte("data2"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group"),
		[]byte("data3"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)

	msgs, err := store.queries.GetAllGroupMessages(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 3)
	require.Equal(t, []byte("data1"), msgs[0].Data)
	require.Equal(t, []byte("data2"), msgs[1].Data)
	require.Equal(t, []byte("data3"), msgs[2].Data)
}

func TestInsertWelcomeMessage_Single(t *testing.T) {
	started := time.Now().UTC().Add(-time.Minute)
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertWelcomeMessage(
		ctx,
		[]byte("installation"),
		[]byte("data"),
		[]byte("hpke"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, int64(1), msg.ID)
	require.True(
		t,
		msg.CreatedAt.Before(time.Now().UTC().Add(1*time.Minute)) && msg.CreatedAt.After(started),
	)
	require.Equal(t, []byte("installation"), msg.InstallationKey)
	require.Equal(t, []byte("data"), msg.Data)
	require.Equal(t, []byte("hpke"), msg.HpkePublicKey)
	require.Equal(t, []byte("metadata"), msg.WelcomeMetadata)

	msgs, err := store.queries.GetAllWelcomeMessages(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, *msg, msgs[0])
}

func TestInsertWelcomeMessage_Duplicate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertWelcomeMessage(
		ctx,
		[]byte("installation"),
		[]byte("data"),
		[]byte("hpke"),
		types.AlgorithmCurve25519,
		[]byte("welcome_metadata"),
	)
	require.NoError(t, err)
	require.NotNil(t, msg)

	msg, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation"),
		[]byte("data"),
		[]byte("hpke"),
		types.AlgorithmCurve25519,
		[]byte("welcome_metadata"),
	)
	require.Nil(t, msg)
	require.IsType(t, &AlreadyExistsError{}, err)
	require.True(t, IsAlreadyExistsError(err))
}

func TestInsertWelcomeMessage_ManyOrderedByTime(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertWelcomeMessage(
		ctx,
		[]byte("installation"),
		[]byte("data1"),
		[]byte("hpke"),
		types.AlgorithmCurve25519,
		[]byte("1"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation"),
		[]byte("data2"),
		[]byte("hpke"),
		types.AlgorithmCurve25519,
		[]byte("2"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation"),
		[]byte("data3"),
		[]byte("hpke"),
		types.AlgorithmCurve25519,
		[]byte("3"),
	)
	require.NoError(t, err)

	msgs, err := store.queries.GetAllWelcomeMessages(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 3)
	require.Equal(t, []byte("data1"), msgs[0].Data)
	require.Equal(t, []byte("data2"), msgs[1].Data)
	require.Equal(t, []byte("data3"), msgs[2].Data)
	require.Equal(t, []byte("1"), msgs[0].WelcomeMetadata)
	require.Equal(t, []byte("2"), msgs[1].WelcomeMetadata)
	require.Equal(t, []byte("3"), msgs[2].WelcomeMetadata)
	require.Greater(t, msgs[1].CreatedAt, msgs[0].CreatedAt)
	require.Greater(t, msgs[2].CreatedAt, msgs[1].CreatedAt)
}

func TestInsertWelcomePointerMessage_Single(t *testing.T) {
	started := time.Now().UTC().Add(-time.Minute)
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertWelcomePointerMessage(
		ctx,
		[]byte("installation_key"),
		[]byte("welcome_pointer_data"),
		[]byte("hpke"),
		types.AlgorithmXwingMlkem768Draft6,
	)
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, int64(1), msg.ID)
	require.True(
		t,
		msg.CreatedAt.Before(time.Now().UTC().Add(1*time.Minute)) && msg.CreatedAt.After(started),
	)
	require.Equal(t, []byte("installation_key"), msg.InstallationKey)
	require.Equal(t, []byte("welcome_pointer_data"), msg.Data)
	require.Equal(t, []byte("hpke"), msg.HpkePublicKey)
	require.Equal(t, int16(1), msg.MessageType) // Welcome pointer type

	msgs, err := store.queries.GetAllWelcomeMessages(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, *msg, msgs[0])
}

func TestInsertWelcomePointerMessage_Comprehensive(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()
	ctx := context.Background()

	tests := []struct {
		name                string
		installationKey     []byte
		welcomePointerData  []byte
		hpkePublicKey       []byte
		wrapperAlgorithm    types.WrapperAlgorithm
		expectError         bool
		expectedMessageType int16
	}{
		{
			name:                "valid welcome pointer with XWING_MLKEM_768_DRAFT_6",
			installationKey:     []byte("installation_key-1"),
			welcomePointerData:  []byte("welcome_pointer_data_1"),
			hpkePublicKey:       []byte("hpke_public_key_1"),
			wrapperAlgorithm:    types.AlgorithmXwingMlkem768Draft6,
			expectError:         false,
			expectedMessageType: 1, // Welcome pointer type
		},
		{
			name:                "valid welcome pointer with CURVE25519",
			installationKey:     []byte("installation_key-2"),
			welcomePointerData:  []byte("welcome_pointer_data_2"),
			hpkePublicKey:       []byte("hpke_public_key_2"),
			wrapperAlgorithm:    types.AlgorithmCurve25519,
			expectError:         false,
			expectedMessageType: 1, // Welcome pointer type
		},
		{
			name:                "invalid welcome pointer with empty HPKE key",
			installationKey:     []byte("installation_key-3"),
			welcomePointerData:  []byte("welcome_pointer_data_3"),
			hpkePublicKey:       []byte{}, // Empty HPKE key should not be allowed
			wrapperAlgorithm:    types.AlgorithmCurve25519,
			expectError:         true,
			expectedMessageType: 1, // Welcome pointer type
		},
		{
			name:                "valid welcome pointer with large data",
			installationKey:     []byte("installation_key-4"),
			welcomePointerData:  make([]byte, 1024), // Large data
			hpkePublicKey:       []byte("hpke_public_key_4"),
			wrapperAlgorithm:    types.AlgorithmXwingMlkem768Draft6,
			expectError:         false,
			expectedMessageType: 1, // Welcome pointer type
		},
	}

	// Insert all test messages
	var insertedMessages []queries.WelcomeMessage
	successCount := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := store.InsertWelcomePointerMessage(
				ctx,
				tt.installationKey,
				tt.welcomePointerData,
				tt.hpkePublicKey,
				tt.wrapperAlgorithm,
			)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, msg)
			} else {
				successCount++
				require.NoError(t, err)
				require.NotNil(t, msg)
				require.Equal(t, successCount, int(msg.ID))
				require.Equal(t, tt.installationKey, msg.InstallationKey)
				require.Equal(t, tt.welcomePointerData, msg.Data)
				require.Equal(t, tt.hpkePublicKey, msg.HpkePublicKey)
				require.Equal(t, int16(tt.wrapperAlgorithm), msg.WrapperAlgorithm)
				require.Equal(t, tt.expectedMessageType, msg.MessageType)
				require.NotZero(t, msg.CreatedAt)
				require.NotNil(t, msg.InstallationKeyDataHash)

				insertedMessages = append(insertedMessages, *msg)
			}
		})
	}

	// Verify all messages were inserted correctly
	allMessages, err := store.queries.GetAllWelcomeMessages(ctx)
	require.NoError(t, err)
	require.Len(t, allMessages, successCount)

	// Verify each message can be queried individually
	for _, expectedMsg := range insertedMessages {
		// Query by installation key
		queryResp, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
			InstallationKey: expectedMsg.InstallationKey,
		})
		require.NoError(t, err)
		require.NotNil(t, queryResp)
		require.GreaterOrEqual(t, len(queryResp.Messages), 1)

		// Find the matching message in the response
		var found bool
		for _, respMsg := range queryResp.Messages {
			if respMsg.GetWelcomePointer() != nil {
				welcomePointer := respMsg.GetWelcomePointer()
				if bytes.Equal(welcomePointer.InstallationKey, expectedMsg.InstallationKey) &&
					bytes.Equal(welcomePointer.WelcomePointer, expectedMsg.Data) &&
					bytes.Equal(welcomePointer.HpkePublicKey, expectedMsg.HpkePublicKey) {
					found = true
					break
				}
			}
		}
		require.True(t, found, "Expected to find welcome pointer message in query response")
	}
}

func TestInsertWelcomePointerMessage_Duplicate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()
	ctx := context.Background()

	installationKey := []byte("duplicate-installation_key")
	welcomePointerData := []byte("duplicate-welcome-pointer-data")
	hpkePublicKey := []byte("duplicate-hpke-public-key")
	wrapperAlgorithm := types.AlgorithmXwingMlkem768Draft6

	// Insert first message
	msg1, err := store.InsertWelcomePointerMessage(
		ctx,
		installationKey,
		welcomePointerData,
		hpkePublicKey,
		wrapperAlgorithm,
	)
	require.NoError(t, err)
	require.NotNil(t, msg1)

	// Try to insert duplicate message (same installation key and data hash)
	msg2, err := store.InsertWelcomePointerMessage(
		ctx,
		installationKey,
		welcomePointerData,
		hpkePublicKey,
		wrapperAlgorithm,
	)
	require.Error(t, err)
	require.Nil(t, msg2)
	require.Contains(t, err.Error(), "duplicate key value violates unique constraint")

	// Verify only one message exists
	allMessages, err := store.queries.GetAllWelcomeMessages(ctx)
	require.NoError(t, err)
	require.Len(t, allMessages, 1)
	require.Equal(t, *msg1, allMessages[0])
}

func TestInsertWelcomePointerMessage_MixedWithRegularWelcome(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Insert a regular welcome message
	regularMsg, err := store.InsertWelcomeMessage(
		ctx,
		[]byte("regular-installation-key"),
		[]byte("regular-welcome-data"),
		[]byte("regular-hpke-public-key"),
		types.AlgorithmCurve25519,
		[]byte("welcome-metadata"),
	)
	require.NoError(t, err)
	require.NotNil(t, regularMsg)
	require.Equal(t, int16(0), regularMsg.MessageType) // Regular welcome message type

	// Insert a welcome pointer message
	pointerMsg, err := store.InsertWelcomePointerMessage(
		ctx,
		[]byte("pointer-installation-key"),
		[]byte("pointer-welcome-data"),
		[]byte("pointer-hpke-public-key"),
		types.AlgorithmXwingMlkem768Draft6,
	)
	require.NoError(t, err)
	require.NotNil(t, pointerMsg)
	require.Equal(t, int16(1), pointerMsg.MessageType) // Welcome pointer type

	// Verify both messages exist
	allMessages, err := store.queries.GetAllWelcomeMessages(ctx)
	require.NoError(t, err)
	require.Len(t, allMessages, 2)

	// Verify message types are different
	messageTypes := make(map[int16]bool)
	for _, msg := range allMessages {
		messageTypes[msg.MessageType] = true
	}
	require.True(t, messageTypes[0], "Should have regular welcome message (type 0)")
	require.True(t, messageTypes[1], "Should have welcome pointer message (type 1)")

	// Query regular welcome message
	regularResp, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("regular-installation-key"),
	})
	require.NoError(t, err)
	require.Len(t, regularResp.Messages, 1)
	require.NotNil(t, regularResp.Messages[0].GetV1())

	// Query welcome pointer message
	pointerResp, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("pointer-installation-key"),
	})
	require.NoError(t, err)
	require.Len(t, pointerResp.Messages, 1)
	require.NotNil(t, pointerResp.Messages[0].GetWelcomePointer())
}

func TestQueryGroupMessagesV1_MissingGroup(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()

	resp, err := store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{})
	require.EqualError(t, err, "group is required")
	require.Nil(t, resp)
}

func TestQueryWelcomeMessagesV1_MissingInstallation(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()

	resp, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{})
	require.EqualError(t, err, "installation is required")
	require.Nil(t, resp)
}

func TestQueryGroupMessagesV1_Filter(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertGroupMessage(
		ctx,
		[]byte("group1"),
		[]byte("data1"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group2"),
		[]byte("data2"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group1"),
		[]byte("data3"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)

	resp, err := store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("unknown"),
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 0)

	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("data3"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("data1"), resp.Messages[1].GetV1().Data)

	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group2"),
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, []byte("data2"), resp.Messages[0].GetV1().Data)

	// Sort ascending
	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
		PagingInfo: &mlsv1.PagingInfo{
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("data1"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("data3"), resp.Messages[1].GetV1().Data)
}

func TestQueryWelcomeMessagesV1_Filter(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertWelcomeMessage(
		ctx,
		[]byte("installation1"),
		[]byte("data1"),
		[]byte("hpke1"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation2"),
		[]byte("data2"),
		[]byte("hpke2"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation1"),
		[]byte("data3"),
		[]byte("hpke3"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)

	resp, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("unknown"),
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 0)

	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("data3"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("data1"), resp.Messages[1].GetV1().Data)
	require.Equal(t, []byte("hpke3"), resp.Messages[0].GetV1().HpkePublicKey)

	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation2"),
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, []byte("data2"), resp.Messages[0].GetV1().Data)

	// Sort ascending
	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo: &mlsv1.PagingInfo{
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("data1"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("data3"), resp.Messages[1].GetV1().Data)
}

func TestQueryGroupMessagesV1_Paginate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertGroupMessage(
		ctx,
		[]byte("group1"),
		[]byte("content1"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group2"),
		[]byte("content2"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group1"),
		[]byte("content3"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group2"),
		[]byte("content4"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group1"),
		[]byte("content5"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group1"),
		[]byte("content6"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group1"),
		[]byte("content7"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(
		ctx,
		[]byte("group1"),
		[]byte("content8"),
		[]byte("sender_hmac"),
		true,
		false,
	)
	require.NoError(t, err)

	resp, err := store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 6)
	require.Equal(t, []byte("content8"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content7"), resp.Messages[1].GetV1().Data)
	require.Equal(t, []byte("content6"), resp.Messages[2].GetV1().Data)
	require.Equal(t, []byte("content5"), resp.Messages[3].GetV1().Data)
	require.Equal(t, []byte("content3"), resp.Messages[4].GetV1().Data)
	require.Equal(t, []byte("content1"), resp.Messages[5].GetV1().Data)

	thirdMsg := resp.Messages[2]
	fifthMsg := resp.Messages[4]

	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content8"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content7"), resp.Messages[1].GetV1().Data)

	// Order descending by default
	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:    2,
			IdCursor: thirdMsg.GetV1().Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content5"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content3"), resp.Messages[1].GetV1().Data)

	// Next page from previous response
	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId:    []byte("group1"),
		PagingInfo: resp.PagingInfo,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, []byte("content1"), resp.Messages[0].GetV1().Data)

	// Order ascending
	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: []byte("group1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:     2,
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
			IdCursor:  fifthMsg.GetV1().Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content5"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content6"), resp.Messages[1].GetV1().Data)

	// Next page from previous response
	resp, err = store.QueryGroupMessagesV1(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId:    []byte("group1"),
		PagingInfo: resp.PagingInfo,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content7"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content8"), resp.Messages[1].GetV1().Data)
}

func TestQueryWelcomeMessagesV1_Paginate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertWelcomeMessage(
		ctx,
		[]byte("installation1"),
		[]byte("content1"),
		[]byte("hpke1"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation2"),
		[]byte("content2"),
		[]byte("hpke2"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation1"),
		[]byte("content3"),
		[]byte("hpke3"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation2"),
		[]byte("content4"),
		[]byte("hpke4"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation1"),
		[]byte("content5"),
		[]byte("hpke5"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation1"),
		[]byte("content6"),
		[]byte("hpke6"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation1"),
		[]byte("content7"),
		[]byte("hpke7"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(
		ctx,
		[]byte("installation1"),
		[]byte("content8"),
		[]byte("hpke8"),
		types.AlgorithmCurve25519,
		[]byte("metadata"),
	)
	require.NoError(t, err)

	resp, err := store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 6)
	require.Equal(t, []byte("content8"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content7"), resp.Messages[1].GetV1().Data)
	require.Equal(t, []byte("content6"), resp.Messages[2].GetV1().Data)
	require.Equal(t, []byte("content5"), resp.Messages[3].GetV1().Data)
	require.Equal(t, []byte("content3"), resp.Messages[4].GetV1().Data)
	require.Equal(t, []byte("content1"), resp.Messages[5].GetV1().Data)

	thirdMsg := resp.Messages[2]
	fifthMsg := resp.Messages[4]

	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content8"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content7"), resp.Messages[1].GetV1().Data)

	// Order descending by default
	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:    2,
			IdCursor: thirdMsg.GetV1().Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content5"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content3"), resp.Messages[1].GetV1().Data)

	// Next page from previous response
	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo:      resp.PagingInfo,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 1)
	require.Equal(t, []byte("content1"), resp.Messages[0].GetV1().Data)

	// Order ascending
	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo: &mlsv1.PagingInfo{
			Limit:     2,
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
			IdCursor:  fifthMsg.GetV1().Id,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content5"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content6"), resp.Messages[1].GetV1().Data)

	// Next page from previous response
	resp, err = store.QueryWelcomeMessagesV1(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: []byte("installation1"),
		PagingInfo:      resp.PagingInfo,
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, 2)
	require.Equal(t, []byte("content7"), resp.Messages[0].GetV1().Data)
	require.Equal(t, []byte("content8"), resp.Messages[1].GetV1().Data)
}

func TestGetNewestGroupMessage(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()
	ctx := context.Background()

	tests := []struct {
		name          string
		setup         func() ([][]byte, []*queries.GroupMessage)
		queryGroups   [][]byte
		expectedCount int
		validate      func([]*queries.GetNewestGroupMessageRow, []*queries.GetNewestGroupMessageMetadataRow, []*queries.GroupMessage)
	}{
		{
			name: "single group",
			setup: func() ([][]byte, []*queries.GroupMessage) {
				group := []byte("group1")
				msg, err := store.InsertGroupMessage(
					ctx,
					group,
					[]byte("data1"),
					[]byte("hmac"),
					true,
					true,
				)
				require.NoError(t, err)
				return [][]byte{group}, []*queries.GroupMessage{msg}
			},
			queryGroups:   [][]byte{[]byte("group1")},
			expectedCount: 1,
			validate: func(results []*queries.GetNewestGroupMessageRow, metadata []*queries.GetNewestGroupMessageMetadataRow, msgs []*queries.GroupMessage) {
				require.NotNil(t, results[0])
				require.Equal(t, []byte("group1"), results[0].GroupID)
				require.Equal(t, msgs[0].ID, results[0].ID)
				require.Equal(t, []byte("data1"), results[0].Data)

				require.NotNil(t, metadata[0])
				require.Equal(t, msgs[0].ID, metadata[0].ID)
				require.Equal(t, results[0].CreatedAt, metadata[0].CreatedAt)
				require.True(t, results[0].IsCommit.Bool)
			},
		},
		{
			name: "multiple groups with highest ID selection",
			setup: func() ([][]byte, []*queries.GroupMessage) {
				group1, group2 := []byte("group1"), []byte("group2")
				_, err := store.InsertGroupMessage(
					ctx,
					group1,
					[]byte("data1"),
					[]byte("hmac"),
					true,
					true,
				)
				require.NoError(t, err)
				msg2, err := store.InsertGroupMessage(
					ctx,
					group1,
					[]byte("data2"),
					[]byte("hmac"),
					true,
					false,
				)
				require.NoError(t, err)
				msg3, err := store.InsertGroupMessage(
					ctx,
					group2,
					[]byte("data3"),
					[]byte("hmac"),
					false,
					true,
				)
				require.NoError(t, err)
				return [][]byte{group1, group2}, []*queries.GroupMessage{msg2, msg3}
			},
			queryGroups:   [][]byte{[]byte("group1"), []byte("group2")},
			expectedCount: 2,
			validate: func(results []*queries.GetNewestGroupMessageRow, metadata []*queries.GetNewestGroupMessageMetadataRow, msgs []*queries.GroupMessage) {
				// Verify results are in input order and contain highest ID messages
				require.NotNil(t, results[0])
				require.Equal(t, []byte("group1"), results[0].GroupID)
				require.Equal(t, msgs[0].ID, results[0].ID) // highest ID for group1
				require.Equal(t, []byte("data2"), results[0].Data)

				require.NotNil(t, results[1])
				require.Equal(t, []byte("group2"), results[1].GroupID)
				require.Equal(t, msgs[1].ID, results[1].ID)

				// Verify metadata matches
				require.Equal(t, msgs[0].ID, metadata[0].ID)
				require.Equal(t, msgs[1].ID, metadata[1].ID)
			},
		},
		{
			name: "no messages",
			setup: func() ([][]byte, []*queries.GroupMessage) {
				return [][]byte{[]byte("group1"), []byte("group2")}, nil
			},
			queryGroups:   [][]byte{[]byte("group1"), []byte("group2")},
			expectedCount: 2,
			validate: func(results []*queries.GetNewestGroupMessageRow, metadata []*queries.GetNewestGroupMessageMetadataRow, msgs []*queries.GroupMessage) {
				require.Nil(t, results[0])
				require.Nil(t, results[1])
				require.Nil(t, metadata[0])
				require.Nil(t, metadata[1])
			},
		},
		{
			name: "mixed results",
			setup: func() ([][]byte, []*queries.GroupMessage) {
				group1, group3 := []byte("group1"), []byte("group3")
				msg1, err := store.InsertGroupMessage(
					ctx,
					group1,
					[]byte("data1"),
					[]byte("hmac"),
					true,
					true,
				)
				require.NoError(t, err)
				msg2, err := store.InsertGroupMessage(
					ctx,
					group3,
					[]byte("data3"),
					[]byte("hmac"),
					false,
					false,
				)
				require.NoError(t, err)
				return [][]byte{
						group1,
						[]byte("group2"),
						group3,
					}, []*queries.GroupMessage{
						msg1,
						msg2,
					}
			},
			queryGroups:   [][]byte{[]byte("group1"), []byte("group2"), []byte("group3")},
			expectedCount: 3,
			validate: func(results []*queries.GetNewestGroupMessageRow, metadata []*queries.GetNewestGroupMessageMetadataRow, msgs []*queries.GroupMessage) {
				require.NotNil(t, results[0])
				require.Equal(t, msgs[0].ID, results[0].ID)

				require.Nil(t, results[1]) // no messages for group2

				require.NotNil(t, results[2])
				require.Equal(t, msgs[1].ID, results[2].ID)

				require.Equal(t, msgs[0].ID, metadata[0].ID)
				require.Nil(t, metadata[1])
				require.Equal(t, msgs[1].ID, metadata[2].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test data
			_, expectedMsgs := tt.setup()

			// Test both methods
			results, err := store.GetNewestGroupMessage(ctx, tt.queryGroups)
			require.NoError(t, err)
			require.Len(t, results, tt.expectedCount)

			metadata, err := store.GetNewestGroupMessageMetadata(ctx, tt.queryGroups)
			require.NoError(t, err)
			require.Len(t, metadata, tt.expectedCount)

			// Validate results
			tt.validate(results, metadata, expectedMsgs)

			// Cleanup for next test
			_, err = store.db.Exec("DELETE FROM group_messages")
			require.NoError(t, err)
		})
	}
}
