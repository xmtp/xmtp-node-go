package api

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	associations "github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func newTestService(t *testing.T, ctx context.Context) (*Service, *bun.DB, func()) {
	log := test.NewLog(t)
	db, _, mlsDbCleanup := test.NewMLSDB(t)
	store, err := mlsstore.New(ctx, mlsstore.Config{
		Log: log,
		DB:  db,
	})
	require.NoError(t, err)
	natsServer, err := server.NewServer(&server.Options{
		Port: server.RANDOM_PORT,
	})
	require.NoError(t, err)
	go natsServer.Start()
	if !natsServer.ReadyForConnections(4 * time.Second) {
		t.Fail()
	}

	svc, err := NewService(log, store)
	require.NoError(t, err)

	return svc, db, func() {
		svc.Close()
		mlsDbCleanup()
	}
}

func makeCreateInbox(address string) *associations.IdentityAction {
	return &associations.IdentityAction{
		Kind: &associations.IdentityAction_CreateInbox{
			CreateInbox: &associations.CreateInbox{
				InitialAddress:          address,
				Nonce:                   0,
				InitialAddressSignature: &associations.Signature{},
			},
		},
	}
}
func makeAddAssociation() *associations.IdentityAction {
	return &associations.IdentityAction{
		Kind: &associations.IdentityAction_Add{
			Add: &associations.AddAssociation{
				NewMemberIdentifier:     &associations.MemberIdentifier{},
				ExistingMemberSignature: &associations.Signature{},
				NewMemberSignature:      &associations.Signature{},
			},
		},
	}
}
func makeRevokeAssociation() *associations.IdentityAction {
	return &associations.IdentityAction{
		Kind: &associations.IdentityAction_Revoke{
			Revoke: &associations.RevokeAssociation{
				MemberToRevoke:           &associations.MemberIdentifier{},
				RecoveryAddressSignature: &associations.Signature{},
			},
		},
	}
}
func makeChangeRecoveryAddress() *associations.IdentityAction {
	return &associations.IdentityAction{
		Kind: &associations.IdentityAction_ChangeRecoveryAddress{
			ChangeRecoveryAddress: &associations.ChangeRecoveryAddress{
				NewRecoveryAddress:               "",
				ExistingRecoveryAddressSignature: &associations.Signature{},
			},
		},
	}
}
func makeIdentityUpdate(inbox_id string, actions ...*associations.IdentityAction) *associations.IdentityUpdate {
	return &associations.IdentityUpdate{
		InboxId:           inbox_id,
		ClientTimestampNs: 0,
		Actions:           actions,
	}
}

func publishIdentityUpdateRequest(inbox_id string, actions ...*associations.IdentityAction) *identity.PublishIdentityUpdateRequest {
	return &identity.PublishIdentityUpdateRequest{
		IdentityUpdate: makeIdentityUpdate(inbox_id, actions...),
	}
}

func makeUpdateRequest(inbox_id string, sequence_id uint64) *identity.GetIdentityUpdatesRequest_Request {
	return &identity.GetIdentityUpdatesRequest_Request{
		InboxId:    inbox_id,
		SequenceId: sequence_id,
	}
}

func getIdentityUpdatesRequest(requests ...*identity.GetIdentityUpdatesRequest_Request) *identity.GetIdentityUpdatesRequest {
	return &identity.GetIdentityUpdatesRequest{
		Requests: requests,
	}
}

/*
*
Tests:
- Can't publish more than 256 updates
- Publishes through one are read in the other(s)
- Ordering is preserved
- Concurrency?
- Invalid updates are rejected
*/

func TestPublishedUpdatesCanBeRead(t *testing.T) {
	ctx := context.Background()
	svc, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	inbox_id := "test_inbox"
	address := "test_address"

	_, err := svc.PublishIdentityUpdate(ctx, publishIdentityUpdateRequest(inbox_id, makeCreateInbox(address)))
	require.NoError(t, err)

	res, err := svc.GetIdentityUpdates(ctx, getIdentityUpdatesRequest(makeUpdateRequest(inbox_id, 0)))
	require.NoError(t, err)

	require.Len(t, res.Responses, 1)
	require.Equal(t, res.Responses[0].InboxId, inbox_id)
	require.Len(t, res.Responses[0].Updates, 1)
	require.Len(t, res.Responses[0].Updates[0].Update.Actions, 1)
	require.Equal(t, res.Responses[0].Updates[0].Update.Actions[0].GetCreateInbox().InitialAddress, address)
}

func TestInboxSizeLimit(t *testing.T) {
	ctx := context.Background()
	svc, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	inbox_id := "test_inbox"
	address := "test_address"

	_, err := svc.PublishIdentityUpdate(ctx, publishIdentityUpdateRequest(inbox_id, makeCreateInbox(address)))
	require.NoError(t, err)

	for i := 0; i < 255; i++ {
		_, err = svc.PublishIdentityUpdate(ctx, publishIdentityUpdateRequest(inbox_id, makeAddAssociation()))
		require.NoError(t, err)
	}

	_, err = svc.PublishIdentityUpdate(ctx, publishIdentityUpdateRequest(inbox_id, makeAddAssociation()))
	require.Error(t, err)

	res, err := svc.GetIdentityUpdates(ctx, getIdentityUpdatesRequest(makeUpdateRequest(inbox_id, 0)))
	require.NoError(t, err)

	require.Len(t, res.Responses, 1)
	require.Equal(t, res.Responses[0].InboxId, inbox_id)
	require.Len(t, res.Responses[0].Updates, 256)
}
