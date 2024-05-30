package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	queries "github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"github.com/xmtp/xmtp-node-go/pkg/mocks"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	testutils "github.com/xmtp/xmtp-node-go/pkg/testing"
	"go.uber.org/zap"
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
	mockMlsValidation.EXPECT().GetAssociationState(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ []*associations.IdentityUpdate, updates []*associations.IdentityUpdate) (*mlsvalidate.AssociationStateResult, error) {
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
					Kind: &associations.MemberIdentifier_Address{
						Address: address,
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

func TestInboxIds(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()
	ctx := context.Background()

	inbox1 := testutils.RandomInboxId()
	inbox2 := testutils.RandomInboxId()
	correctInbox := testutils.RandomInboxId()
	correctInbox2 := testutils.RandomInboxId()

	_, err := store.queries.InsertAddressLog(ctx, queries.InsertAddressLogParams{Address: "address", InboxID: inbox1, AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 1}, RevocationSequenceID: sql.NullInt64{Valid: false}})
	require.NoError(t, err)
	_, err = store.queries.InsertAddressLog(ctx, queries.InsertAddressLogParams{Address: "address", InboxID: inbox1, AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 2}, RevocationSequenceID: sql.NullInt64{Valid: false}})
	require.NoError(t, err)
	_, err = store.queries.InsertAddressLog(ctx, queries.InsertAddressLogParams{Address: "address", InboxID: inbox1, AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 3}, RevocationSequenceID: sql.NullInt64{Valid: false}})
	require.NoError(t, err)
	_, err = store.queries.InsertAddressLog(ctx, queries.InsertAddressLogParams{Address: "address", InboxID: correctInbox, AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 4}, RevocationSequenceID: sql.NullInt64{Valid: false}})
	require.NoError(t, err)

	reqs := make([]*identity.GetInboxIdsRequest_Request, 0)
	reqs = append(reqs, &identity.GetInboxIdsRequest_Request{
		Address: "address",
	})
	req := &identity.GetInboxIdsRequest{
		Requests: reqs,
	}
	resp, _ := store.GetInboxIds(context.Background(), req)

	require.Equal(t, correctInbox, *resp.Responses[0].InboxId)

	_, err = store.queries.InsertAddressLog(ctx, queries.InsertAddressLogParams{Address: "address", InboxID: correctInbox2, AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 5}, RevocationSequenceID: sql.NullInt64{Valid: false}})
	require.NoError(t, err)
	resp, _ = store.GetInboxIds(context.Background(), req)
	require.Equal(t, correctInbox2, *resp.Responses[0].InboxId)

	reqs = append(reqs, &identity.GetInboxIdsRequest_Request{Address: "address2"})
	req = &identity.GetInboxIdsRequest{
		Requests: reqs,
	}
	_, err = store.queries.InsertAddressLog(ctx, queries.InsertAddressLogParams{Address: "address2", InboxID: inbox2, AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 8}, RevocationSequenceID: sql.NullInt64{Valid: false}})
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

	_, err := store.queries.InsertAddressLog(ctx, queries.InsertAddressLogParams{Address: "address_1", InboxID: inbox1, AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 1}, RevocationSequenceID: sql.NullInt64{Valid: false}})
	require.NoError(t, err)
	_, err = store.queries.InsertAddressLog(ctx, queries.InsertAddressLogParams{Address: "address_2", InboxID: inbox2, AssociationSequenceID: sql.NullInt64{Valid: true, Int64: 2}, RevocationSequenceID: sql.NullInt64{Valid: false}})
	require.NoError(t, err)

	reqs := make([]*identity.GetInboxIdsRequest_Request, 0)
	reqs = append(reqs, &identity.GetInboxIdsRequest_Request{
		Address: "address_1",
	})
	reqs = append(reqs, &identity.GetInboxIdsRequest_Request{
		Address: "address_2",
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
	installationId := test.RandomBytes(32)
	inboxId := test.RandomInboxId()

	err := store.CreateInstallation(ctx, installationId, inboxId, test.RandomBytes(32), 0)
	require.NoError(t, err)

	installationFromDb, err := store.queries.GetInstallation(ctx, installationId)
	require.NoError(t, err)
	require.Equal(t, installationFromDb.ID, installationId)
}

func TestUpdateKeyPackage(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := test.RandomBytes(32)
	inboxId := test.RandomInboxId()
	keyPackage := test.RandomBytes(32)

	err := store.CreateInstallation(ctx, installationId, inboxId, keyPackage, 0)
	require.NoError(t, err)

	keyPackage2 := test.RandomBytes(32)
	err = store.UpdateKeyPackage(ctx, installationId, keyPackage2, 1)
	require.NoError(t, err)

	installationFromDb, err := store.queries.GetInstallation(ctx, installationId)
	require.NoError(t, err)

	require.Equal(t, keyPackage2, installationFromDb.KeyPackage)
	require.Equal(t, int64(1), installationFromDb.Expiration)
}

func TestConsumeLastResortKeyPackage(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	installationId := test.RandomBytes(32)
	keyPackage := test.RandomBytes(32)
	inboxId := test.RandomInboxId()

	err := store.CreateInstallation(ctx, installationId, inboxId, keyPackage, 0)
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
	msg, err := store.InsertGroupMessage(ctx, []byte("group"), []byte("data"))
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, int64(1), msg.ID)
	store.log.Info("Created at", zap.Time("created_at", msg.CreatedAt))
	require.True(t, msg.CreatedAt.Before(time.Now().UTC().Add(1*time.Minute)) && msg.CreatedAt.After(started))
	require.Equal(t, []byte("group"), msg.GroupID)
	require.Equal(t, []byte("data"), msg.Data)

	msgs, err := store.queries.GetAllGroupMessages(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, *msg, msgs[0])
}

func TestInsertGroupMessage_Duplicate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertGroupMessage(ctx, []byte("group"), []byte("data"))
	require.NoError(t, err)
	require.NotNil(t, msg)

	msg, err = store.InsertGroupMessage(ctx, []byte("group"), []byte("data"))
	require.Nil(t, msg)
	require.IsType(t, &AlreadyExistsError{}, err)
	require.True(t, IsAlreadyExistsError(err))
}

func TestInsertGroupMessage_ManyOrderedByTime(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertGroupMessage(ctx, []byte("group"), []byte("data1"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group"), []byte("data2"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group"), []byte("data3"))
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
	msg, err := store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data"), []byte("hpke"))
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, int64(1), msg.ID)
	require.True(t, msg.CreatedAt.Before(time.Now().UTC().Add(1*time.Minute)) && msg.CreatedAt.After(started))
	require.Equal(t, []byte("installation"), msg.InstallationKey)
	require.Equal(t, []byte("data"), msg.Data)
	require.Equal(t, []byte("hpke"), msg.HpkePublicKey)

	msgs, err := store.queries.GetAllWelcomeMessages(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, *msg, msgs[0])
}

func TestInsertWelcomeMessage_Duplicate(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	msg, err := store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data"), []byte("hpke"))
	require.NoError(t, err)
	require.NotNil(t, msg)

	msg, err = store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data"), []byte("hpke"))
	require.Nil(t, msg)
	require.IsType(t, &AlreadyExistsError{}, err)
	require.True(t, IsAlreadyExistsError(err))
}

func TestInsertWelcomeMessage_ManyOrderedByTime(t *testing.T) {
	store, cleanup := NewTestStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data1"), []byte("hpke"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data2"), []byte("hpke"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation"), []byte("data3"), []byte("hpke"))
	require.NoError(t, err)

	msgs, err := store.queries.GetAllWelcomeMessages(ctx)
	require.NoError(t, err)
	require.Len(t, msgs, 3)
	require.Equal(t, []byte("data1"), msgs[0].Data)
	require.Equal(t, []byte("data2"), msgs[1].Data)
	require.Equal(t, []byte("data3"), msgs[2].Data)
	require.Greater(t, msgs[1].CreatedAt, msgs[0].CreatedAt)
	require.Greater(t, msgs[2].CreatedAt, msgs[1].CreatedAt)
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
	_, err := store.InsertGroupMessage(ctx, []byte("group1"), []byte("data1"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group2"), []byte("data2"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("data3"))
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
	_, err := store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("data1"), []byte("hpke1"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation2"), []byte("data2"), []byte("hpke2"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("data3"), []byte("hpke3"))
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
	_, err := store.InsertGroupMessage(ctx, []byte("group1"), []byte("content1"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group2"), []byte("content2"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("content3"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group2"), []byte("content4"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("content5"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("content6"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("content7"))
	require.NoError(t, err)
	_, err = store.InsertGroupMessage(ctx, []byte("group1"), []byte("content8"))
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
	_, err := store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content1"), []byte("hpke1"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation2"), []byte("content2"), []byte("hpke2"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content3"), []byte("hpke3"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation2"), []byte("content4"), []byte("hpke4"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content5"), []byte("hpke5"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content6"), []byte("hpke6"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content7"), []byte("hpke7"))
	require.NoError(t, err)
	_, err = store.InsertWelcomeMessage(ctx, []byte("installation1"), []byte("content8"), []byte("hpke8"))
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
