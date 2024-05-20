package mlsvalidate

import (
	"context"
	"fmt"

	identity_proto "github.com/xmtp/xmtp-node-go/pkg/proto/identity"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	associations "github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	svc "github.com/xmtp/xmtp-node-go/pkg/proto/mls_validation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type IdentityValidationResult struct {
	AccountAddress     string
	InstallationKey    []byte
	CredentialIdentity []byte
	Expiration         uint64
}

type InboxIdValidationResult struct {
	InstallationKey []byte
	Credential      *identity_proto.MlsCredential
	Expiration      uint64
}

type GroupMessageValidationResult struct {
	GroupId string
}

type AssociationStateResult struct {
	AssociationState *associations.AssociationState     `protobuf:"bytes,1,opt,name=association_state,json=associationState,proto3" json:"association_state,omitempty"`
	StateDiff        *associations.AssociationStateDiff `protobuf:"bytes,2,opt,name=state_diff,json=stateDiff,proto3" json:"state_diff,omitempty"`
}

type IdentityInput struct {
	SigningPublicKey []byte
	Identity         []byte
}

type MLSValidationService interface {
	ValidateInboxIdKeyPackages(ctx context.Context, keyPackages [][]byte) ([]InboxIdValidationResult, error)
	ValidateV3KeyPackages(ctx context.Context, keyPackages [][]byte) ([]IdentityValidationResult, error)
	ValidateGroupMessages(ctx context.Context, groupMessages []*mlsv1.GroupMessageInput) ([]GroupMessageValidationResult, error)
	GetAssociationState(ctx context.Context, oldUpdates []*associations.IdentityUpdate, newUpdates []*associations.IdentityUpdate) (*AssociationStateResult, error)
}

type IdentityStore interface {
	GetInboxLogs(ctx context.Context, req *identity.GetIdentityUpdatesRequest) (*identity.GetIdentityUpdatesResponse, error)
}

type MLSValidationServiceImpl struct {
	grpcClient    svc.ValidationApiClient
	identityStore IdentityStore
}

type InboxIdKeyPackage struct {
	InboxId               string
	InstallationPublicKey []byte
}

func NewMlsValidationService(ctx context.Context, options MLSValidationOptions, identityStore IdentityStore) (*MLSValidationServiceImpl, error) {
	conn, err := grpc.DialContext(ctx, options.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &MLSValidationServiceImpl{
		grpcClient:    svc.NewValidationApiClient(conn),
		identityStore: identityStore,
	}, nil
}

func (s *MLSValidationServiceImpl) GetAssociationState(ctx context.Context, oldUpdates []*associations.IdentityUpdate, newUpdates []*associations.IdentityUpdate) (*AssociationStateResult, error) {
	req := &svc.GetAssociationStateRequest{
		OldUpdates: oldUpdates,
		NewUpdates: newUpdates,
	}
	response, err := s.grpcClient.GetAssociationState(ctx, req)
	if err != nil {
		return nil, err
	}
	return &AssociationStateResult{
		AssociationState: response.GetAssociationState(),
		StateDiff:        response.GetStateDiff(),
	}, nil
}

type KeyAndExpiration struct {
	InstallationId []byte
	Expiration     uint64
}

func (s *MLSValidationServiceImpl) ValidateInboxIdKeyPackages(ctx context.Context, keyPackages [][]byte) ([]InboxIdValidationResult, error) {
	req := makeValidateKeyPackageRequest(keyPackages, true)

	response, err := s.grpcClient.ValidateInboxIdKeyPackages(ctx, req)
	if err != nil {
		return nil, err
	}

	keyPackageCredential := make(map[string]KeyAndExpiration, len(response.Responses))
	identityUpdatesRequest := make([]*identity.GetIdentityUpdatesRequest_Request, len(response.Responses))
	for i, response := range response.Responses {
		if !response.IsOk {
			return nil, fmt.Errorf("validation failed with error %s", response.ErrorMessage)
		}
		keyPackageCredential[response.Credential.InboxId] = KeyAndExpiration{
			InstallationId: response.InstallationPublicKey,
			Expiration:     response.Expiration,
		}
		identityUpdatesRequest[i] = &identity.GetIdentityUpdatesRequest_Request{
			InboxId:    response.Credential.InboxId,
			SequenceId: 0,
		}
	}

	// TODO: do we need to take sequence ID Into account?
	request := &identity.GetIdentityUpdatesRequest{Requests: identityUpdatesRequest}
	identity_updates, err := s.identityStore.GetInboxLogs(ctx, request)
	if err != nil {
		return nil, err
	}

	validation_requests := make([]*svc.ValidateInboxIdsRequest_ValidationRequest, len(identity_updates.Responses))
	for i, response := range identity_updates.Responses {
		validation_requests[i] = makeValidationRequest(response, keyPackageCredential)
	}

	validation_request := svc.ValidateInboxIdsRequest{Requests: validation_requests}
	validate_inbox_response, err := s.grpcClient.ValidateInboxIds(ctx, &validation_request)
	if err != nil {
		return nil, err
	}

	out := make([]InboxIdValidationResult, len(response.Responses))
	for i, response := range validate_inbox_response.Responses {
		if !response.IsOk {
			return nil, fmt.Errorf("validation failed with error %s", response.ErrorMessage)
		}
		out[i] = InboxIdValidationResult{
			InstallationKey: keyPackageCredential[response.InboxId].InstallationId,
			Credential:      &identity_proto.MlsCredential{InboxId: response.InboxId},
			Expiration:      keyPackageCredential[response.InboxId].Expiration,
		}
	}
	return out, nil
}

func makeValidationRequest(update *identity.GetIdentityUpdatesResponse_Response, pub_keys map[string]KeyAndExpiration) *svc.ValidateInboxIdsRequest_ValidationRequest {
	identity_updates := make([]*associations.IdentityUpdate, len(update.Updates))
	for i, identity_log := range update.Updates {
		identity_updates[i] = identity_log.Update
	}

	out := svc.ValidateInboxIdsRequest_ValidationRequest{
		Credential:            &identity_proto.MlsCredential{InboxId: update.InboxId},
		InstallationPublicKey: pub_keys[update.InboxId].InstallationId,
		IdentityUpdates:       identity_updates,
	}

	return &out
}

func (s *MLSValidationServiceImpl) ValidateV3KeyPackages(ctx context.Context, keyPackages [][]byte) ([]IdentityValidationResult, error) {
	req := makeValidateKeyPackageRequest(keyPackages, false)

	response, err := s.grpcClient.ValidateKeyPackages(ctx, req)
	if err != nil {
		return nil, err
	}

	out := make([]IdentityValidationResult, len(response.Responses))
	for i, response := range response.Responses {
		if !response.IsOk {
			return nil, fmt.Errorf("validation failed with error %s", response.ErrorMessage)
		}
		out[i] = IdentityValidationResult{
			AccountAddress:     response.AccountAddress,
			InstallationKey:    response.InstallationId,
			CredentialIdentity: response.CredentialIdentityBytes,
			Expiration:         response.Expiration,
		}
	}

	return out, nil
}

func makeValidateKeyPackageRequest(keyPackageBytes [][]byte, isInboxIdCredential bool) *svc.ValidateKeyPackagesRequest {
	keyPackageRequests := make([]*svc.ValidateKeyPackagesRequest_KeyPackage, len(keyPackageBytes))
	for i, keyPackage := range keyPackageBytes {
		keyPackageRequests[i] = &svc.ValidateKeyPackagesRequest_KeyPackage{
			KeyPackageBytesTlsSerialized: keyPackage,
			IsInboxIdCredential:          isInboxIdCredential,
		}
	}
	return &svc.ValidateKeyPackagesRequest{
		KeyPackages: keyPackageRequests,
	}
}

func (s *MLSValidationServiceImpl) ValidateGroupMessages(ctx context.Context, groupMessages []*mlsv1.GroupMessageInput) ([]GroupMessageValidationResult, error) {
	req := makeValidateGroupMessagesRequest(groupMessages)

	response, err := s.grpcClient.ValidateGroupMessages(ctx, req)
	if err != nil {
		return nil, err
	}

	out := make([]GroupMessageValidationResult, len(response.Responses))
	for i, response := range response.Responses {
		if !response.IsOk {
			return nil, fmt.Errorf("validation failed with error %s", response.ErrorMessage)
		}
		out[i] = GroupMessageValidationResult{
			GroupId: response.GroupId,
		}
	}

	return out, nil
}

func makeValidateGroupMessagesRequest(groupMessages []*mlsv1.GroupMessageInput) *svc.ValidateGroupMessagesRequest {
	groupMessageRequests := make([]*svc.ValidateGroupMessagesRequest_GroupMessage, len(groupMessages))
	for i, groupMessage := range groupMessages {
		groupMessageRequests[i] = &svc.ValidateGroupMessagesRequest_GroupMessage{
			GroupMessageBytesTlsSerialized: groupMessage.GetV1().Data,
		}
	}
	return &svc.ValidateGroupMessagesRequest{
		GroupMessages: groupMessageRequests,
	}
}
