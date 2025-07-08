package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	associations "github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

type mockedMLSValidationService struct {
	mock.Mock
}

func (m *mockedMLSValidationService) GetAssociationState(
	ctx context.Context,
	oldUpdates []*associations.IdentityUpdate,
	newUpdates []*associations.IdentityUpdate,
) (*mlsvalidate.AssociationStateResult, error) {
	member_map := make([]*associations.MemberMap, 0)
	member_map = append(member_map, &associations.MemberMap{
		Key: &associations.MemberIdentifier{
			Kind: &associations.MemberIdentifier_EthereumAddress{EthereumAddress: "key_address"},
		},
		Value: &associations.Member{
			Identifier: &associations.MemberIdentifier{
				Kind: &associations.MemberIdentifier_EthereumAddress{EthereumAddress: "ident"},
			},
			AddedByEntity: &associations.MemberIdentifier{
				Kind: &associations.MemberIdentifier_EthereumAddress{
					EthereumAddress: "added_by_entity",
				},
			},
		},
	})

	new_members := make([]*associations.MemberIdentifier, 0)

	new_members = append(
		new_members,
		&associations.MemberIdentifier{
			Kind: &associations.MemberIdentifier_EthereumAddress{EthereumAddress: "0x01"},
		},
	)
	new_members = append(
		new_members,
		&associations.MemberIdentifier{
			Kind: &associations.MemberIdentifier_EthereumAddress{EthereumAddress: "0x02"},
		},
	)
	new_members = append(
		new_members,
		&associations.MemberIdentifier{
			Kind: &associations.MemberIdentifier_EthereumAddress{EthereumAddress: "0x03"},
		},
	)

	out := mlsvalidate.AssociationStateResult{
		AssociationState: &associations.AssociationState{
			InboxId:            "test_inbox",
			Members:            member_map,
			RecoveryIdentifier: "recovery",
			SeenSignatures:     [][]byte{[]byte("seen"), []byte("sig")},
		},
		StateDiff: &associations.AssociationStateDiff{
			NewMembers:     new_members,
			RemovedMembers: nil,
		},
	}
	return &out, nil
}

func (m *mockedMLSValidationService) ValidateV3KeyPackages(
	ctx context.Context,
	keyPackages [][]byte,
) ([]mlsvalidate.IdentityValidationResult, error) {
	return nil, nil
}

func (m *mockedMLSValidationService) ValidateGroupMessages(
	ctx context.Context,
	groupMessages []*mlsv1.GroupMessageInput,
) ([]mlsvalidate.GroupMessageValidationResult, error) {
	return nil, nil
}

func (m *mockedMLSValidationService) ValidateInboxIdKeyPackages(
	ctx context.Context,
	keyPackages [][]byte,
) ([]mlsvalidate.InboxIdValidationResult, error) {
	return nil, nil
}

func (m *mockedMLSValidationService) VerifySmartContractWalletSignatures(
	ctx context.Context,
	req *identity.VerifySmartContractWalletSignaturesRequest,
) (*identity.VerifySmartContractWalletSignaturesResponse, error) {
	args := m.Called(ctx, req)

	return args.Get(0).(*identity.VerifySmartContractWalletSignaturesResponse), args.Error(1)
}

func newMockedValidationService() *mockedMLSValidationService {
	return new(mockedMLSValidationService)
}

func newTestService(t *testing.T, ctx context.Context) (*Service, *bun.DB, func()) {
	log := test.NewLog(t)
	db, _, mlsDbCleanup := test.NewMLSDB(t)
	store, err := mlsstore.New(ctx, mlsstore.Config{
		Log: log,
		DB:  db,
	})
	require.NoError(t, err)
	mlsValidationService := newMockedValidationService()

	svc, err := NewService(log, store, mlsValidationService)
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
				InitialIdentifier:          address,
				InitialIdentifierKind:      associations.IdentifierKind_IDENTIFIER_KIND_ETHEREUM,
				Nonce:                      0,
				InitialIdentifierSignature: &associations.Signature{},
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
				MemberToRevoke:              &associations.MemberIdentifier{},
				RecoveryIdentifierSignature: &associations.Signature{},
			},
		},
	}
}

func makeChangeRecoveryAddress() *associations.IdentityAction {
	return &associations.IdentityAction{
		Kind: &associations.IdentityAction_ChangeRecoveryAddress{
			ChangeRecoveryAddress: &associations.ChangeRecoveryAddress{
				NewRecoveryIdentifier:               "",
				ExistingRecoveryIdentifierSignature: &associations.Signature{},
			},
		},
	}
}

func makeIdentityUpdate(
	inbox_id string,
	actions ...*associations.IdentityAction,
) *associations.IdentityUpdate {
	return &associations.IdentityUpdate{
		InboxId:           inbox_id,
		ClientTimestampNs: 0,
		Actions:           actions,
	}
}

func publishIdentityUpdateRequest(
	inbox_id string,
	actions ...*associations.IdentityAction,
) *identity.PublishIdentityUpdateRequest {
	return &identity.PublishIdentityUpdateRequest{
		IdentityUpdate: makeIdentityUpdate(inbox_id, actions...),
	}
}

func makeUpdateRequest(
	inbox_id string,
	sequence_id uint64,
) *identity.GetIdentityUpdatesRequest_Request {
	return &identity.GetIdentityUpdatesRequest_Request{
		InboxId:    inbox_id,
		SequenceId: sequence_id,
	}
}

func getIdentityUpdatesRequest(
	requests ...*identity.GetIdentityUpdatesRequest_Request,
) *identity.GetIdentityUpdatesRequest {
	return &identity.GetIdentityUpdatesRequest{
		Requests: requests,
	}
}

func TestPublishedUpdatesCanBeRead(t *testing.T) {
	ctx := context.Background()
	svc, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	inbox_id := test.RandomInboxId()
	address := "test_address"

	_, err := svc.PublishIdentityUpdate(
		ctx,
		publishIdentityUpdateRequest(inbox_id, makeCreateInbox(address)),
	)
	require.NoError(t, err)

	res, err := svc.GetIdentityUpdates(
		ctx,
		getIdentityUpdatesRequest(makeUpdateRequest(inbox_id, 0)),
	)
	require.NoError(t, err)

	require.Len(t, res.Responses, 1)
	require.Equal(t, res.Responses[0].InboxId, inbox_id)
	require.Len(t, res.Responses[0].Updates, 1)
	require.Len(t, res.Responses[0].Updates[0].Update.Actions, 1)
	require.Equal(
		t,
		res.Responses[0].Updates[0].Update.Actions[0].GetCreateInbox().InitialIdentifier,
		address,
	)
}

func TestPublishedUpdatesAreInOrder(t *testing.T) {
	ctx := context.Background()
	svc, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	inbox_id := test.RandomInboxId()
	address := "test_address"

	_, err := svc.PublishIdentityUpdate(
		ctx,
		publishIdentityUpdateRequest(inbox_id, makeCreateInbox(address)),
	)
	require.NoError(t, err)
	_, err = svc.PublishIdentityUpdate(
		ctx,
		publishIdentityUpdateRequest(inbox_id, makeAddAssociation(), makeChangeRecoveryAddress()),
	)
	require.NoError(t, err)
	_, err = svc.PublishIdentityUpdate(
		ctx,
		publishIdentityUpdateRequest(inbox_id, makeRevokeAssociation()),
	)
	require.NoError(t, err)

	res, err := svc.GetIdentityUpdates(
		ctx,
		getIdentityUpdatesRequest(makeUpdateRequest(inbox_id, 0)),
	)
	require.NoError(t, err)

	require.Len(t, res.Responses, 1)
	require.Equal(t, res.Responses[0].InboxId, inbox_id)
	require.Len(t, res.Responses[0].Updates, 3)
	require.NotNil(t, res.Responses[0].Updates[0].Update.Actions[0].GetCreateInbox())
	require.NotNil(t, res.Responses[0].Updates[1].Update.Actions[0].GetAdd())
	require.NotNil(t, res.Responses[0].Updates[1].Update.Actions[1].GetChangeRecoveryAddress())
	require.NotNil(t, res.Responses[0].Updates[2].Update.Actions[0].GetRevoke())

	res, err = svc.GetIdentityUpdates(
		ctx,
		getIdentityUpdatesRequest(makeUpdateRequest(inbox_id, 1)),
	)
	require.NoError(t, err)

	require.Len(t, res.Responses, 1)
	require.Equal(t, res.Responses[0].InboxId, inbox_id)
	require.Len(t, res.Responses[0].Updates, 2)
	require.NotNil(t, res.Responses[0].Updates[0].Update.Actions[0].GetAdd())
	require.NotNil(t, res.Responses[0].Updates[0].Update.Actions[1].GetChangeRecoveryAddress())
	require.NotNil(t, res.Responses[0].Updates[1].Update.Actions[0].GetRevoke())
}

func TestQueryMultipleInboxes(t *testing.T) {
	ctx := context.Background()
	svc, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	first_inbox_id := test.RandomInboxId()
	second_inbox_id := test.RandomInboxId()
	first_address := test.RandomInboxId()
	second_address := test.RandomInboxId()

	_, err := svc.PublishIdentityUpdate(
		ctx,
		publishIdentityUpdateRequest(first_inbox_id, makeCreateInbox(first_address)),
	)
	require.NoError(t, err)
	_, err = svc.PublishIdentityUpdate(
		ctx,
		publishIdentityUpdateRequest(second_inbox_id, makeCreateInbox(second_address)),
	)
	require.NoError(t, err)

	res, err := svc.GetIdentityUpdates(
		ctx,
		getIdentityUpdatesRequest(
			makeUpdateRequest(first_inbox_id, 0),
			makeUpdateRequest(second_inbox_id, 0),
		),
	)
	require.NoError(t, err)

	require.Len(t, res.Responses, 2)
	require.Equal(
		t,
		res.Responses[0].Updates[0].Update.Actions[0].GetCreateInbox().InitialIdentifier,
		first_address,
	)
	require.Equal(
		t,
		res.Responses[1].Updates[0].Update.Actions[0].GetCreateInbox().InitialIdentifier,
		second_address,
	)
}

func TestInboxSizeLimit(t *testing.T) {
	ctx := context.Background()
	svc, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	inbox_id := test.RandomInboxId()
	address := "test_address"

	_, err := svc.PublishIdentityUpdate(
		ctx,
		publishIdentityUpdateRequest(inbox_id, makeCreateInbox(address)),
	)
	require.NoError(t, err)

	for i := 0; i < 255; i++ {
		_, err = svc.PublishIdentityUpdate(
			ctx,
			publishIdentityUpdateRequest(inbox_id, makeAddAssociation()),
		)
		require.NoError(t, err)
	}

	_, err = svc.PublishIdentityUpdate(
		ctx,
		publishIdentityUpdateRequest(inbox_id, makeAddAssociation()),
	)
	require.Error(t, err)

	res, err := svc.GetIdentityUpdates(
		ctx,
		getIdentityUpdatesRequest(makeUpdateRequest(inbox_id, 0)),
	)
	require.NoError(t, err)

	require.Len(t, res.Responses, 1)
	require.Equal(t, res.Responses[0].InboxId, inbox_id)
	require.Len(t, res.Responses[0].Updates, 256)
}
