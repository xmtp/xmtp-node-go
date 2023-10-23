package mlsvalidate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	svc "github.com/xmtp/proto/v3/go/mls_validation/v1"
	"google.golang.org/grpc"
)

type MockedGrpcService struct {
	mock.Mock
}

func (m *MockedGrpcService) ValidateKeyPackages(ctx context.Context, req *svc.ValidateKeyPackagesRequest, opts ...grpc.CallOption) (*svc.ValidateKeyPackagesResponse, error) {
	args := m.Called(ctx, req)

	return args.Get(0).(*svc.ValidateKeyPackagesResponse), args.Error(1)
}

func (m *MockedGrpcService) ValidateGroupMessages(ctx context.Context, req *svc.ValidateGroupMessagesRequest, opts ...grpc.CallOption) (*svc.ValidateGroupMessagesResponse, error) {
	args := m.Called(ctx, req)

	return args.Get(0).(*svc.ValidateGroupMessagesResponse), args.Error(1)
}

func getMockedService() (*MockedGrpcService, MlsValidationService) {
	mockService := new(MockedGrpcService)
	service := &MlsValidationServiceImpl{
		grpcClient: mockService,
	}

	return mockService, service
}

func TestValidateKeyPackages(t *testing.T) {
	mockGrpc, service := getMockedService()

	ctx := context.Background()

	firstResponse := svc.ValidateKeyPackagesResponse_ValidationResponse{
		IsOk:           true,
		WalletAddress:  "0x123",
		InstallationId: "123",
		ErrorMessage:   "",
	}

	mockGrpc.On("ValidateKeyPackages", ctx, mock.Anything).Return(&svc.ValidateKeyPackagesResponse{
		Responses: []*svc.ValidateKeyPackagesResponse_ValidationResponse{&firstResponse},
	}, nil)

	res, err := service.ValidateKeyPackages(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "0x123", res[0].WalletAddress)
	assert.Equal(t, "123", res[0].InstallationId)
}

func TestValidateKeyPackagesError(t *testing.T) {

}
