package mlsvalidate

import (
	"context"
	"fmt"

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
	ValidateKeyPackages(ctx context.Context, keyPackages [][]byte) ([]IdentityValidationResult, error)
	ValidateGroupMessages(ctx context.Context, groupMessages []*mlsv1.GroupMessageInput) ([]GroupMessageValidationResult, error)
	GetAssociationState(ctx context.Context, oldUpdates []*associations.IdentityUpdate, newUpdates []*associations.IdentityUpdate) (*AssociationStateResult, error)
}

type MLSValidationServiceImpl struct {
	grpcClient svc.ValidationApiClient
}

func NewMlsValidationService(ctx context.Context, options MLSValidationOptions) (*MLSValidationServiceImpl, error) {
	conn, err := grpc.DialContext(ctx, options.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &MLSValidationServiceImpl{
		grpcClient: svc.NewValidationApiClient(conn),
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

func (s *MLSValidationServiceImpl) ValidateKeyPackages(ctx context.Context, keyPackages [][]byte) ([]IdentityValidationResult, error) {
	req := makeValidateKeyPackageRequest(keyPackages)
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

func makeValidateKeyPackageRequest(keyPackageBytes [][]byte) *svc.ValidateKeyPackagesRequest {
	keyPackageRequests := make([]*svc.ValidateKeyPackagesRequest_KeyPackage, len(keyPackageBytes))
	for i, keyPackage := range keyPackageBytes {
		keyPackageRequests[i] = &svc.ValidateKeyPackagesRequest_KeyPackage{
			KeyPackageBytesTlsSerialized: keyPackage,
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
