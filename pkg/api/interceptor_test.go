package api

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	identityv1 "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	messagev1 "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/ratelimiter"
)

// Helper function to create context with client IP for testing
func contextWithClientIP(ip string) context.Context {
	md := metadata.New(map[string]string{
		"x-forwarded-for": ip,
	})
	return metadata.NewIncomingContext(context.Background(), md)
}

// MockRateLimiter implements the RateLimiter interface for testing
type MockRateLimiter struct {
	mu         sync.RWMutex
	spendCalls []SpendCall
	shouldFail bool
	failError  error
}

type SpendCall struct {
	LimitType  ratelimiter.LimitType
	Bucket     string
	Cost       uint16
	IsPriority bool
}

func NewMockRateLimiter() *MockRateLimiter {
	return &MockRateLimiter{
		spendCalls: make([]SpendCall, 0),
	}
}

func (m *MockRateLimiter) Spend(
	limitType ratelimiter.LimitType,
	bucket string,
	cost uint16,
	isPriority bool,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	call := SpendCall{
		LimitType:  limitType,
		Bucket:     bucket,
		Cost:       cost,
		IsPriority: isPriority,
	}
	m.spendCalls = append(m.spendCalls, call)

	if m.shouldFail {
		return m.failError
	}
	return nil
}

func (m *MockRateLimiter) GetSpendCalls() []SpendCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([]SpendCall, len(m.spendCalls))
	copy(calls, m.spendCalls)
	return calls
}

func (m *MockRateLimiter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.spendCalls = make([]SpendCall, 0)
	m.shouldFail = false
	m.failError = nil
}

func (m *MockRateLimiter) SetShouldFail(shouldFail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
	m.failError = err
}

func TestRateLimitInterceptor_RequestTypesAndCosts(t *testing.T) {
	logger := zap.NewNop()
	mockLimiter := NewMockRateLimiter()
	interceptor := NewRateLimitInterceptor(mockLimiter, logger)

	testCases := []struct {
		name          string
		request       interface{}
		expectedCost  uint16
		expectedLimit ratelimiter.LimitType
	}{
		{
			name:          "RegisterInstallationRequest",
			request:       &mlsv1.RegisterInstallationRequest{},
			expectedCost:  1,
			expectedLimit: ratelimiter.PUBLISH,
		},
		{
			name:          "UploadKeyPackageRequest",
			request:       &mlsv1.UploadKeyPackageRequest{},
			expectedCost:  1,
			expectedLimit: ratelimiter.PUBLISH,
		},
		{
			name:          "PublishIdentityUpdateRequest",
			request:       &identityv1.PublishIdentityUpdateRequest{},
			expectedCost:  1,
			expectedLimit: ratelimiter.PUBLISH,
		},
		{
			name: "SendWelcomeMessagesRequest with single message",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{{}},
			},
			expectedCost:  1,
			expectedLimit: ratelimiter.PUBLISH,
		},
		{
			name: "SendWelcomeMessagesRequest with multiple messages",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: []*mlsv1.WelcomeMessageInput{{}, {}, {}},
			},
			expectedCost:  3,
			expectedLimit: ratelimiter.PUBLISH,
		},
		{
			name: "SendGroupMessagesRequest with single message",
			request: &mlsv1.SendGroupMessagesRequest{
				Messages: []*mlsv1.GroupMessageInput{{}},
			},
			expectedCost:  1,
			expectedLimit: ratelimiter.PUBLISH,
		},
		{
			name: "SendGroupMessagesRequest with multiple messages",
			request: &mlsv1.SendGroupMessagesRequest{
				Messages: []*mlsv1.GroupMessageInput{{}, {}, {}, {}},
			},
			expectedCost:  4,
			expectedLimit: ratelimiter.PUBLISH,
		},
		{
			name: "PublishRequest with single envelope",
			request: &messagev1.PublishRequest{
				Envelopes: []*messagev1.Envelope{{}},
			},
			expectedCost:  1,
			expectedLimit: ratelimiter.PUBLISH,
		},
		{
			name: "PublishRequest with multiple envelopes",
			request: &messagev1.PublishRequest{
				Envelopes: []*messagev1.Envelope{{}, {}},
			},
			expectedCost:  2,
			expectedLimit: ratelimiter.PUBLISH,
		},
		{
			name: "BatchQueryRequest with single request",
			request: &messagev1.BatchQueryRequest{
				Requests: []*messagev1.QueryRequest{{}},
			},
			expectedCost:  1,
			expectedLimit: ratelimiter.DEFAULT,
		},
		{
			name: "BatchQueryRequest with multiple requests",
			request: &messagev1.BatchQueryRequest{
				Requests: []*messagev1.QueryRequest{{}, {}, {}},
			},
			expectedCost:  3,
			expectedLimit: ratelimiter.DEFAULT,
		},
		{
			name:          "QueryRequest (default case)",
			request:       &messagev1.QueryRequest{},
			expectedCost:  1,
			expectedLimit: ratelimiter.DEFAULT,
		},
		{
			name:          "SubscribeRequest (default case)",
			request:       &messagev1.SubscribeRequest{},
			expectedCost:  1,
			expectedLimit: ratelimiter.DEFAULT,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLimiter.Reset()

			// Create a context with a client IP
			ctx := contextWithClientIP("192.168.1.1")

			// Test unary interceptor
			err := interceptor.applyLimits(ctx, "/TestService/TestMethod", tc.request)
			require.NoError(t, err)

			calls := mockLimiter.GetSpendCalls()
			require.Len(t, calls, 1)

			call := calls[0]
			require.Equal(t, tc.expectedLimit, call.LimitType)
			require.Equal(t, tc.expectedCost, call.Cost)
			require.Equal(t, "R192.168.1.1"+string(tc.expectedLimit), call.Bucket)
			require.False(t, call.IsPriority)
		})
	}
}

func TestRateLimitInterceptor_AllowlistedIPs(t *testing.T) {
	logger := zap.NewNop()
	mockLimiter := NewMockRateLimiter()
	interceptor := NewRateLimitInterceptor(mockLimiter, logger)

	allowlistedIPs := []string{"12.76.45.218", "104.190.130.147"}

	for _, ip := range allowlistedIPs {
		t.Run(fmt.Sprintf("allowlisted IP %s", ip), func(t *testing.T) {
			mockLimiter.Reset()

			// Create a context with an allowlisted IP
			ctx := contextWithClientIP(ip)

			// Test with a default request
			err := interceptor.applyLimits(
				ctx,
				"/TestService/TestMethod",
				&messagev1.QueryRequest{},
			)
			require.NoError(t, err)

			// Should not have any spend calls since allowlisted IPs are not rate limited
			calls := mockLimiter.GetSpendCalls()
			require.Len(t, calls, 0)
		})
	}
}

func TestRateLimitInterceptor_PriorityBuckets(t *testing.T) {
	logger := zap.NewNop()
	mockLimiter := NewMockRateLimiter()
	interceptor := NewRateLimitInterceptor(mockLimiter, logger)

	testCases := []struct {
		name             string
		ip               string
		request          interface{}
		expectedBucket   string
		expectedPriority bool
	}{
		{
			name:             "regular IP with DEFAULT limit",
			ip:               "192.168.1.1",
			request:          &messagev1.QueryRequest{},
			expectedBucket:   "R192.168.1.1" + string(ratelimiter.DEFAULT),
			expectedPriority: false,
		},
		{
			name:             "regular IP with PUBLISH limit",
			ip:               "192.168.1.1",
			request:          &mlsv1.RegisterInstallationRequest{},
			expectedBucket:   "R192.168.1.1" + string(ratelimiter.PUBLISH),
			expectedPriority: false,
		},
		{
			name:             "allowlisted IP should not be rate limited",
			ip:               "12.76.45.218",
			request:          &messagev1.QueryRequest{},
			expectedBucket:   "",
			expectedPriority: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLimiter.Reset()

			ctx := contextWithClientIP(tc.ip)

			err := interceptor.applyLimits(ctx, "/TestService/TestMethod", tc.request)
			require.NoError(t, err)

			calls := mockLimiter.GetSpendCalls()

			if tc.expectedBucket == "" {
				// Allowlisted IP should not generate any calls
				require.Len(t, calls, 0)
			} else {
				require.Len(t, calls, 1)
				call := calls[0]
				require.Equal(t, tc.expectedBucket, call.Bucket)
				require.Equal(t, tc.expectedPriority, call.IsPriority)
			}
		})
	}
}

func TestRateLimitInterceptor_NoIPAddress(t *testing.T) {
	logger := zap.NewNop()
	mockLimiter := NewMockRateLimiter()
	interceptor := NewRateLimitInterceptor(mockLimiter, logger)

	ctx := context.Background()
	// Don't set any IP address

	err := interceptor.applyLimits(ctx, "/TestService/TestMethod", &messagev1.QueryRequest{})
	require.NoError(t, err)

	calls := mockLimiter.GetSpendCalls()
	require.Len(t, calls, 1)

	call := calls[0]
	require.Equal(t, "Rip_unknown"+string(ratelimiter.DEFAULT), call.Bucket)
	require.False(t, call.IsPriority)
}

func TestRateLimitInterceptor_RateLimitExceeded(t *testing.T) {
	logger := zap.NewNop()
	mockLimiter := NewMockRateLimiter()
	interceptor := NewRateLimitInterceptor(mockLimiter, logger)

	expectedError := fmt.Errorf("rate limit exceeded")
	mockLimiter.SetShouldFail(true, expectedError)

	ctx := contextWithClientIP("192.168.1.1")

	err := interceptor.applyLimits(ctx, "/TestService/TestMethod", &messagev1.QueryRequest{})
	require.Error(t, err)

	// Should return ResourceExhausted status code
	grpcErr, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.ResourceExhausted, grpcErr.Code())
}

func TestRateLimitInterceptor_UnaryInterceptor(t *testing.T) {
	logger := zap.NewNop()
	mockLimiter := NewMockRateLimiter()
	interceptor := NewRateLimitInterceptor(mockLimiter, logger)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	ctx := contextWithClientIP("192.168.1.1")

	unaryInterceptor := interceptor.Unary()

	resp, err := unaryInterceptor(ctx, &messagev1.QueryRequest{}, &grpc.UnaryServerInfo{
		FullMethod: "/message_api.v1.MessageApi/Query",
	}, handler)

	require.NoError(t, err)
	require.Equal(t, "success", resp)

	calls := mockLimiter.GetSpendCalls()
	require.Len(t, calls, 1)
	require.Equal(t, ratelimiter.DEFAULT, calls[0].LimitType)
}

func TestRateLimitInterceptor_StreamInterceptor(t *testing.T) {
	logger := zap.NewNop()
	mockLimiter := NewMockRateLimiter()
	interceptor := NewRateLimitInterceptor(mockLimiter, logger)

	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	ctx := contextWithClientIP("192.168.1.1")

	mockStream := &mockServerStream{ctx: ctx}

	streamInterceptor := interceptor.Stream()

	err := streamInterceptor(&messagev1.SubscribeRequest{}, mockStream, &grpc.StreamServerInfo{
		FullMethod: "/message_api.v1.MessageApi/Subscribe",
	}, handler)

	require.NoError(t, err)

	calls := mockLimiter.GetSpendCalls()
	require.Len(t, calls, 1)
	require.Equal(t, ratelimiter.DEFAULT, calls[0].LimitType)
}

func TestRateLimitInterceptor_StreamInterceptorWithRateLimit(t *testing.T) {
	logger := zap.NewNop()
	mockLimiter := NewMockRateLimiter()
	interceptor := NewRateLimitInterceptor(mockLimiter, logger)

	expectedError := fmt.Errorf("rate limit exceeded")
	mockLimiter.SetShouldFail(true, expectedError)

	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	ctx := contextWithClientIP("192.168.1.1")

	mockStream := &mockServerStream{ctx: ctx}

	streamInterceptor := interceptor.Stream()

	err := streamInterceptor(&messagev1.SubscribeRequest{}, mockStream, &grpc.StreamServerInfo{
		FullMethod: "/message_api.v1.MessageApi/Subscribe",
	}, handler)

	require.Error(t, err)

	grpcErr, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.ResourceExhausted, grpcErr.Code())
}

func TestRateLimitInterceptor_CostCalculation(t *testing.T) {
	logger := zap.NewNop()
	mockLimiter := NewMockRateLimiter()
	interceptor := NewRateLimitInterceptor(mockLimiter, logger)

	testCases := []struct {
		name         string
		request      interface{}
		expectedCost uint16
	}{
		{
			name: "SendWelcomeMessagesRequest with 10 messages",
			request: &mlsv1.SendWelcomeMessagesRequest{
				Messages: make([]*mlsv1.WelcomeMessageInput, 10),
			},
			expectedCost: 10,
		},
		{
			name: "SendGroupMessagesRequest with 5 messages",
			request: &mlsv1.SendGroupMessagesRequest{
				Messages: make([]*mlsv1.GroupMessageInput, 5),
			},
			expectedCost: 5,
		},
		{
			name: "PublishRequest with 7 envelopes",
			request: &messagev1.PublishRequest{
				Envelopes: make([]*messagev1.Envelope, 7),
			},
			expectedCost: 7,
		},
		{
			name: "BatchQueryRequest with 3 requests",
			request: &messagev1.BatchQueryRequest{
				Requests: make([]*messagev1.QueryRequest, 3),
			},
			expectedCost: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLimiter.Reset()

			ctx := contextWithClientIP("192.168.1.1")

			err := interceptor.applyLimits(ctx, "/TestService/TestMethod", tc.request)
			require.NoError(t, err)

			calls := mockLimiter.GetSpendCalls()
			require.Len(t, calls, 1)
			require.Equal(t, tc.expectedCost, calls[0].Cost)
		})
	}
}

func TestRateLimitInterceptor_IsAllowListedIp(t *testing.T) {
	testCases := []struct {
		ip       string
		expected bool
	}{
		{"12.76.45.218", true},
		{"104.190.130.147", true},
		{"192.168.1.1", false},
		{"127.0.0.1", false},
		{"10.0.0.1", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.ip, func(t *testing.T) {
			result := isAllowListedIp(tc.ip)
			require.Equal(t, tc.expected, result)
		})
	}
}

// mockServerStream implements grpc.ServerStream for testing
type mockServerStream struct {
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func (m *mockServerStream) SendMsg(msg interface{}) error {
	return nil
}

func (m *mockServerStream) RecvMsg(msg interface{}) error {
	return nil
}

func (m *mockServerStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerStream) SetTrailer(md metadata.MD) {
}
