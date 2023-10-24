package mlsvalidate

import (
	"context"
	"fmt"

	svc "github.com/xmtp/proto/v3/go/mls_validation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type IdentityValidationResult struct {
	WalletAddress  string
	InstallationId string
}

type GroupMessageValidationResult struct {
	GroupId string
}

type IdentityInput struct {
	SigningPublicKey []byte
	Identity         []byte
}

type MLSValidationService interface {
	ValidateKeyPackages(ctx context.Context, keyPackages [][]byte) ([]IdentityValidationResult, error)
	ValidateGroupMessages(ctx context.Context, groupMessages [][]byte) ([]GroupMessageValidationResult, error)
}

type MLSValidationServiceImpl struct {
	grpcClient svc.ValidationApiClient
}

func NewMlsValidationService(ctx context.Context, options MlsValidationOptions) (*MLSValidationServiceImpl, error) {
	conn, err := grpc.DialContext(ctx, options.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &MLSValidationServiceImpl{
		grpcClient: svc.NewValidationApiClient(conn),
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
			WalletAddress:  response.WalletAddress,
			InstallationId: response.InstallationId,
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

func (s *MLSValidationServiceImpl) ValidateGroupMessages(ctx context.Context, groupMessages [][]byte) ([]GroupMessageValidationResult, error) {
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

func makeValidateGroupMessagesRequest(groupMessages [][]byte) *svc.ValidateGroupMessagesRequest {
	groupMessageRequests := make([]*svc.ValidateGroupMessagesRequest_GroupMessage, len(groupMessages))
	for i, groupMessage := range groupMessages {
		groupMessageRequests[i] = &svc.ValidateGroupMessagesRequest_GroupMessage{
			GroupMessageBytesTlsSerialized: groupMessage,
		}
	}
	return &svc.ValidateGroupMessagesRequest{
		GroupMessages: groupMessageRequests,
	}
}
