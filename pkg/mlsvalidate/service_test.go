package mlsvalidate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	svc "github.com/xmtp/xmtp-node-go/pkg/proto/mls_validation/v1"
	"google.golang.org/grpc"
)

type MockedGRPCService struct {
	mock.Mock
}

func (m *MockedGRPCService) GetAssociationState(ctx context.Context, in *svc.GetAssociationStateRequest, opts ...grpc.CallOption) (*svc.GetAssociationStateResponse, error) {
	return nil, nil
}

func (m *MockedGRPCService) ValidateKeyPackages(ctx context.Context, req *svc.ValidateKeyPackagesRequest, opts ...grpc.CallOption) (*svc.ValidateKeyPackagesResponse, error) {
	args := m.Called(ctx, req)

	return args.Get(0).(*svc.ValidateKeyPackagesResponse), args.Error(1)
}

func (m *MockedGRPCService) ValidateGroupMessages(ctx context.Context, req *svc.ValidateGroupMessagesRequest, opts ...grpc.CallOption) (*svc.ValidateGroupMessagesResponse, error) {
	args := m.Called(ctx, req)

	return args.Get(0).(*svc.ValidateGroupMessagesResponse), args.Error(1)
}

func (m *MockedGRPCService) ValidateInboxIdKeyPackages(ctx context.Context, req *svc.ValidateKeyPackagesRequest, opts ...grpc.CallOption) (*svc.ValidateInboxIdKeyPackagesResponse, error) {
	args := m.Called(ctx, req)

	return args.Get(0).(*svc.ValidateInboxIdKeyPackagesResponse), args.Error(1)
}

func (m *MockedGRPCService) VerifySmartContractWalletSignatures(ctx context.Context, req *identity.VerifySmartContractWalletSignaturesRequest, opts ...grpc.CallOption) (*identity.VerifySmartContractWalletSignaturesResponse, error) {
	args := m.Called(ctx, req)

	return args.Get(0).(*identity.VerifySmartContractWalletSignaturesResponse), args.Error(1)
}

func getMockedService() (*MockedGRPCService, MLSValidationService) {
	mockService := new(MockedGRPCService)
	service := &MLSValidationServiceImpl{
		grpcClient: mockService,
	}

	return mockService, service
}

func TestValidateInboxIdKeyPackages(t *testing.T) {
	mockGrpc, service := getMockedService()

	ctx := context.Background()
	installationKey := []byte("key")
	firstResponse := svc.ValidateInboxIdKeyPackagesResponse_Response{
		IsOk:                  true,
		Credential:            nil,
		InstallationPublicKey: installationKey,
		ErrorMessage:          "",
	}
	mockGrpc.On("ValidateInboxIdKeyPackages", ctx, mock.Anything).Return(&svc.ValidateInboxIdKeyPackagesResponse{Responses: []*svc.ValidateInboxIdKeyPackagesResponse_Response{&firstResponse}}, nil)

	res, err := service.ValidateInboxIdKeyPackages(ctx, [][]byte{[]byte("123")})
	assert.NoError(t, err)
	assert.Equal(t, res[0].InstallationKey, installationKey)
}

func TestValidateInboxIdKeyPackagesError(t *testing.T) {
	mockGrpc, service := getMockedService()

	ctx := context.Background()
	firstResponse := svc.ValidateInboxIdKeyPackagesResponse_Response{
		IsOk:                  false,
		Credential:            nil,
		InstallationPublicKey: []byte("foo"),
		ErrorMessage:          "DERP",
	}
	mockGrpc.On("ValidateInboxIdKeyPackages", ctx, mock.Anything).Return(&svc.ValidateInboxIdKeyPackagesResponse{Responses: []*svc.ValidateInboxIdKeyPackagesResponse_Response{&firstResponse}}, nil)

	res, err := service.ValidateInboxIdKeyPackages(ctx, [][]byte{[]byte("123")})
	assert.Error(t, err)
	assert.Nil(t, res)
}
