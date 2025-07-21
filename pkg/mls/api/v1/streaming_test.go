package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/mocks"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	messageContentsProto "github.com/xmtp/xmtp-node-go/pkg/proto/mls/message_contents"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// testConfig holds common test configuration
type testConfig struct {
	timeout           time.Duration
	subscriptionDelay time.Duration
	bufferSize        int
}

// defaultTestConfig returns default test configuration
func defaultTestConfig() testConfig {
	return testConfig{
		timeout:           5 * time.Second,
		subscriptionDelay: 100 * time.Millisecond,
		bufferSize:        100,
	}
}

// messageCollector handles collecting and tracking received messages
type messageCollector struct {
	messages chan *mlsv1.GroupMessage
	count    int64
}

// newMessageCollector creates a new message collector
func newMessageCollector(bufferSize int) *messageCollector {
	return &messageCollector{
		messages: make(chan *mlsv1.GroupMessage, bufferSize),
		count:    0,
	}
}

// collect adds a message to the collector
func (mc *messageCollector) collect(msg *mlsv1.GroupMessage) {
	mc.messages <- msg
	atomic.AddInt64(&mc.count, 1)
}

// waitForCount waits for the collector to receive the expected number of messages
func (mc *messageCollector) waitForCount(expected int, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return fmt.Errorf(
				"timeout waiting for messages: expected %d, got %d",
				expected,
				atomic.LoadInt64(&mc.count),
			)
		case <-ticker.C:
			if atomic.LoadInt64(&mc.count) >= int64(expected) {
				return nil
			}
		}
	}
}

// getAllMessages returns all collected messages
func (mc *messageCollector) getAllMessages() []*mlsv1.GroupMessage {
	close(mc.messages)
	var messages []*mlsv1.GroupMessage
	for msg := range mc.messages {
		messages = append(messages, msg)
	}
	return messages
}

// welcomeMessageCollector handles collecting welcome messages
type welcomeMessageCollector struct {
	messages chan *mlsv1.WelcomeMessage
	count    int64
}

// newWelcomeMessageCollector creates a new welcome message collector
func newWelcomeMessageCollector(bufferSize int) *welcomeMessageCollector {
	return &welcomeMessageCollector{
		messages: make(chan *mlsv1.WelcomeMessage, bufferSize),
		count:    0,
	}
}

// collect adds a welcome message to the collector
func (wmc *welcomeMessageCollector) collect(msg *mlsv1.WelcomeMessage) {
	wmc.messages <- msg
	atomic.AddInt64(&wmc.count, 1)
}

// waitForCount waits for the collector to receive the expected number of messages
func (wmc *welcomeMessageCollector) waitForCount(expected int, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return fmt.Errorf(
				"timeout waiting for welcome messages: expected %d, got %d",
				expected,
				atomic.LoadInt64(&wmc.count),
			)
		case <-ticker.C:
			if atomic.LoadInt64(&wmc.count) >= int64(expected) {
				return nil
			}
		}
	}
}

// getAllMessages returns all collected welcome messages
func (wmc *welcomeMessageCollector) getAllMessages() []*mlsv1.WelcomeMessage {
	close(wmc.messages)
	var messages []*mlsv1.WelcomeMessage
	for msg := range wmc.messages {
		messages = append(messages, msg)
	}
	return messages
}

// setupGroupMessageStream creates a mock stream for group message subscriptions
func setupGroupMessageStream(
	t *testing.T,
	ctx context.Context,
	collector *messageCollector,
) *mocks.MockMlsApi_SubscribeGroupMessagesServer {
	stream := mocks.NewMockMlsApi_SubscribeGroupMessagesServer(t)
	stream.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	stream.EXPECT().
		Send(mock.AnythingOfType("*apiv1.GroupMessage")).
		Run(func(msg *mlsv1.GroupMessage) {
			collector.collect(msg)
		}).
		Return(nil).
		Maybe()
	stream.EXPECT().Context().Return(ctx)
	return stream
}

// setupWelcomeMessageStream creates a mock stream for welcome message subscriptions
func setupWelcomeMessageStream(
	t *testing.T,
	ctx context.Context,
	collector *welcomeMessageCollector,
) *mocks.MockMlsApi_SubscribeWelcomeMessagesServer {
	stream := mocks.NewMockMlsApi_SubscribeWelcomeMessagesServer(t)
	stream.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	stream.EXPECT().
		Send(mock.AnythingOfType("*apiv1.WelcomeMessage")).
		Run(func(msg *mlsv1.WelcomeMessage) {
			collector.collect(msg)
		}).
		Return(nil).
		Maybe()
	stream.EXPECT().Context().Return(ctx)
	return stream
}

// startGroupMessageSubscription starts a group message subscription
func startGroupMessageSubscription(
	t *testing.T,
	svc *Service,
	filters []*mlsv1.SubscribeGroupMessagesRequest_Filter,
	stream *mocks.MockMlsApi_SubscribeGroupMessagesServer,
) {
	go func() {
		err := svc.SubscribeGroupMessages(&mlsv1.SubscribeGroupMessagesRequest{
			Filters: filters,
		}, stream)
		if err != nil {
			if status.Code(err) == codes.Unavailable || errors.Is(err, context.Canceled) {
				// Ignore error as context cancellation or shutdown is expected
				return
			}
			// Log unexpected errors using test framework
			t.Errorf("unexpected error in group message subscription: %v", err)
		}
	}()
}

// startWelcomeMessageSubscription starts a welcome message subscription
func startWelcomeMessageSubscription(
	t *testing.T,
	svc *Service,
	filters []*mlsv1.SubscribeWelcomeMessagesRequest_Filter,
	stream *mocks.MockMlsApi_SubscribeWelcomeMessagesServer,
) {
	go func() {
		err := svc.SubscribeWelcomeMessages(&mlsv1.SubscribeWelcomeMessagesRequest{
			Filters: filters,
		}, stream)
		if err != nil {
			if status.Code(err) == codes.Unavailable || errors.Is(err, context.Canceled) {
				// Ignore error as context cancellation or shutdown is expected
				return
			}
			// Log unexpected errors using test framework
			t.Errorf("unexpected error in welcome message subscription: %v", err)
		}
	}()
}

// sendGroupMessages sends multiple group messages concurrently
func sendGroupMessages(
	t *testing.T,
	ctx context.Context,
	svc *Service,
	groupId []byte,
	numGoroutines, messagesPerGoroutine int,
	messagePrefix string,
) {
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineId int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msgData := fmt.Sprintf("%s_%d_%d", messagePrefix, goroutineId, j)
				_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
					Messages: []*mlsv1.GroupMessageInput{
						{
							Version: &mlsv1.GroupMessageInput_V1_{
								V1: &mlsv1.GroupMessageInput_V1{
									Data:       []byte(msgData),
									SenderHmac: []byte(fmt.Sprintf("hmac_%d_%d", goroutineId, j)),
									ShouldPush: true,
								},
							},
						},
					},
				})
				require.NoError(t, err)
			}
		}(i)
	}
	wg.Wait()
}

// sendWelcomeMessages sends multiple welcome messages concurrently
func sendWelcomeMessages(
	t *testing.T,
	ctx context.Context,
	svc *Service,
	installationKey, hpkePublicKey []byte,
	numGoroutines, messagesPerGoroutine int,
	messagePrefix string,
) {
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineId int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msgData := fmt.Sprintf("%s_%d_%d", messagePrefix, goroutineId, j)
				_, err := svc.SendWelcomeMessages(ctx, &mlsv1.SendWelcomeMessagesRequest{
					Messages: []*mlsv1.WelcomeMessageInput{
						{
							Version: &mlsv1.WelcomeMessageInput_V1_{
								V1: &mlsv1.WelcomeMessageInput_V1{
									InstallationKey:  installationKey,
									Data:             []byte(msgData),
									HpkePublicKey:    hpkePublicKey,
									WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
									WelcomeMetadata: []byte(
										fmt.Sprintf("metadata_%d_%d", goroutineId, j),
									),
								},
							},
						},
					},
				})
				require.NoError(t, err)
			}
		}(i)
	}
	wg.Wait()
}

// populateGroupMessages sends initial messages to a group
func populateGroupMessages(
	t *testing.T,
	ctx context.Context,
	svc *Service,
	groupId []byte,
	count int,
	prefix string,
) {
	for i := 0; i < count; i++ {
		_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
			Messages: []*mlsv1.GroupMessageInput{
				{
					Version: &mlsv1.GroupMessageInput_V1_{
						V1: &mlsv1.GroupMessageInput_V1{
							Data:       []byte(fmt.Sprintf("%s_%d", prefix, i)),
							SenderHmac: []byte(fmt.Sprintf("hmac_%d", i)),
							ShouldPush: true,
						},
					},
				},
			},
		})
		require.NoError(t, err)
	}
}

// populateWelcomeMessages sends initial welcome messages
func populateWelcomeMessages(
	t *testing.T,
	ctx context.Context,
	svc *Service,
	installationKey, hpkePublicKey []byte,
	count int,
	prefix string,
) {
	for i := 0; i < count; i++ {
		_, err := svc.SendWelcomeMessages(ctx, &mlsv1.SendWelcomeMessagesRequest{
			Messages: []*mlsv1.WelcomeMessageInput{
				{
					Version: &mlsv1.WelcomeMessageInput_V1_{
						V1: &mlsv1.WelcomeMessageInput_V1{
							InstallationKey:  installationKey,
							Data:             []byte(fmt.Sprintf("%s_%d", prefix, i)),
							HpkePublicKey:    hpkePublicKey,
							WrapperAlgorithm: messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
						},
					},
				},
			},
		})
		require.NoError(t, err)
	}
}

// validateMessageOrdering validates that messages are in ascending order by ID
func validateMessageOrdering(t *testing.T, messages []*mlsv1.GroupMessage) {
	for i := 1; i < len(messages); i++ {
		assert.Greater(t, messages[i].GetV1().Id, messages[i-1].GetV1().Id,
			"Messages should be in ascending sequence_id order")
	}
}

// validateWelcomeMessageOrdering validates that welcome messages are in ascending order by ID
func validateWelcomeMessageOrdering(t *testing.T, messages []*mlsv1.WelcomeMessage) {
	for i := 1; i < len(messages); i++ {
		assert.Greater(t, messages[i].GetV1().Id, messages[i-1].GetV1().Id,
			"Welcome messages should be in ascending sequence_id order")
	}
}

// validateNoDuplicates validates that no duplicate message IDs exist
func validateNoDuplicates(t *testing.T, messages []*mlsv1.GroupMessage) {
	seenIds := make(map[uint64]bool)
	for _, msg := range messages {
		id := msg.GetV1().Id
		assert.False(t, seenIds[id], "Should not receive duplicate messages")
		seenIds[id] = true
	}
}

// validateWelcomeMessageNoDuplicates validates that no duplicate welcome message IDs exist
func validateWelcomeMessageNoDuplicates(t *testing.T, messages []*mlsv1.WelcomeMessage) {
	seenIds := make(map[uint64]bool)
	for _, msg := range messages {
		id := msg.GetV1().Id
		assert.False(t, seenIds[id], "Should not receive duplicate welcome messages")
		seenIds[id] = true
	}
}

// validateMessagesAfterCursor validates that all messages have IDs greater than the cursor
func validateMessagesAfterCursor(
	t *testing.T,
	messages []*mlsv1.GroupMessage,
	cursorPosition uint64,
) {
	for _, msg := range messages {
		assert.Greater(t, msg.GetV1().Id, cursorPosition,
			"All messages should have sequence_id greater than cursor")
	}
}

// validateWelcomeMessagesAfterCursor validates that all welcome messages have IDs greater than the cursor
func validateWelcomeMessagesAfterCursor(
	t *testing.T,
	messages []*mlsv1.WelcomeMessage,
	cursorPosition uint64,
) {
	for _, msg := range messages {
		assert.Greater(t, msg.GetV1().Id, cursorPosition,
			"All welcome messages should have sequence_id greater than cursor")
	}
}

// TestGroupMessageStreaming_OrderingWithConcurrentWrites tests that messages are received in correct sequence_id order
// even when multiple goroutines are writing concurrently
func TestGroupMessageStreaming_OrderingWithConcurrentWrites(t *testing.T) {
	ctx := context.Background()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	config := defaultTestConfig()
	groupId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupId)

	collector := newMessageCollector(config.bufferSize)
	stream := setupGroupMessageStream(t, ctx, collector)

	// Start subscription
	startGroupMessageSubscription(t, svc, []*mlsv1.SubscribeGroupMessagesRequest_Filter{
		{GroupId: groupId},
	}, stream)

	time.Sleep(config.subscriptionDelay)

	// Send messages concurrently
	numGoroutines := 10
	messagesPerGoroutine := 10
	totalMessages := numGoroutines * messagesPerGoroutine

	sendGroupMessages(t, ctx, svc, groupId, numGoroutines, messagesPerGoroutine, "msg")

	// Wait for all messages
	err := collector.waitForCount(totalMessages, config.timeout)
	require.NoError(t, err, "Should receive all messages within timeout")

	// Validate results
	messages := collector.getAllMessages()
	assert.GreaterOrEqual(t, len(messages), totalMessages, "Should receive all messages")
	validateMessageOrdering(t, messages)
	validateNoDuplicates(t, messages)
}

// TestGroupMessageStreaming_CursorConsistency tests that subscriptions with cursors receive all messages after the cursor
func TestGroupMessageStreaming_CursorConsistency(t *testing.T) {
	ctx := context.Background()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	config := defaultTestConfig()
	groupId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupId)

	// Pre-populate with messages
	numInitialMessages := 20
	populateGroupMessages(t, ctx, svc, groupId, numInitialMessages, "initial")

	// Get the current messages to establish cursor position
	resp, err := svc.QueryGroupMessages(ctx, &mlsv1.QueryGroupMessagesRequest{
		GroupId: groupId,
		PagingInfo: &mlsv1.PagingInfo{
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, numInitialMessages)

	// Use cursor at position 10 (should receive messages 12-20 plus new messages)
	cursorPosition := resp.Messages[9].GetV1().Id // 10th message
	expectedMessagesAfterCursor := numInitialMessages - 10

	collector := newMessageCollector(config.bufferSize)
	stream := setupGroupMessageStream(t, ctx, collector)

	// Start subscription with cursor
	startGroupMessageSubscription(t, svc, []*mlsv1.SubscribeGroupMessagesRequest_Filter{
		{
			GroupId:  groupId,
			IdCursor: cursorPosition,
		},
	}, stream)

	// Allow subscription to start and process historical messages
	time.Sleep(200 * time.Millisecond)

	// Send additional messages
	numNewMessages := 5
	populateGroupMessages(t, ctx, svc, groupId, numNewMessages, "new")

	expectedTotalMessages := expectedMessagesAfterCursor + numNewMessages

	// Wait for all expected messages
	err = collector.waitForCount(expectedTotalMessages, config.timeout)
	require.NoError(t, err, "Should receive all messages after cursor within timeout")

	// Validate results
	messages := collector.getAllMessages()
	assert.GreaterOrEqual(
		t,
		len(messages),
		expectedTotalMessages,
		"Should receive all messages after cursor",
	)
	validateMessagesAfterCursor(t, messages, cursorPosition)
	validateMessageOrdering(t, messages)
}

// TestGroupMessageStreaming_RaceConditionHandling tests various race conditions
func TestGroupMessageStreaming_RaceConditionHandling(t *testing.T) {
	ctx := context.Background()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupId)

	// Test scenario: Start subscription, send messages, then start another subscription
	// Both should receive the same messages in order

	var receivedMessages1 []*mlsv1.GroupMessage
	var receivedMessages2 []*mlsv1.GroupMessage
	var mu1, mu2 sync.Mutex

	stream1 := mocks.NewMockMlsApi_SubscribeGroupMessagesServer(t)
	stream1.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	stream1.EXPECT().
		Send(mock.AnythingOfType("*apiv1.GroupMessage")).
		Run(func(msg *mlsv1.GroupMessage) {
			mu1.Lock()
			receivedMessages1 = append(receivedMessages1, msg)
			mu1.Unlock()
		}).
		Return(nil).
		Maybe()
	stream1.EXPECT().Context().Return(ctx)

	stream2 := mocks.NewMockMlsApi_SubscribeGroupMessagesServer(t)
	stream2.EXPECT().SendHeader(metadata.New(map[string]string{"subscribed": "true"})).Return(nil)
	stream2.EXPECT().
		Send(mock.AnythingOfType("*apiv1.GroupMessage")).
		Run(func(msg *mlsv1.GroupMessage) {
			mu2.Lock()
			receivedMessages2 = append(receivedMessages2, msg)
			mu2.Unlock()
		}).
		Return(nil).
		Maybe()
	stream2.EXPECT().Context().Return(ctx)

	// Start first subscription
	startGroupMessageSubscription(t, svc, []*mlsv1.SubscribeGroupMessagesRequest_Filter{
		{GroupId: groupId},
	}, stream1)

	time.Sleep(50 * time.Millisecond)

	// Send some messages
	numMessages := 10
	populateGroupMessages(t, ctx, svc, groupId, numMessages, "race_test")

	time.Sleep(50 * time.Millisecond)

	// Start second subscription (should receive all messages from the beginning)
	startGroupMessageSubscription(t, svc, []*mlsv1.SubscribeGroupMessagesRequest_Filter{
		{GroupId: groupId},
	}, stream2)

	// Wait for messages to be processed
	time.Sleep(2 * time.Second)

	// Verify both subscriptions received messages in order
	mu1.Lock()
	msgs1 := make([]*mlsv1.GroupMessage, len(receivedMessages1))
	copy(msgs1, receivedMessages1)
	mu1.Unlock()

	mu2.Lock()
	msgs2 := make([]*mlsv1.GroupMessage, len(receivedMessages2))
	copy(msgs2, receivedMessages2)
	mu2.Unlock()

	assert.GreaterOrEqual(
		t,
		len(msgs1),
		numMessages,
		"First subscription should receive all messages",
	)
	assert.GreaterOrEqual(
		t,
		len(msgs2),
		numMessages,
		"Second subscription should receive all messages",
	)

	// Verify ordering in both subscriptions
	validateMessageOrdering(t, msgs1)
	validateMessageOrdering(t, msgs2)
}

// TestWelcomeMessageStreaming_OrderingWithConcurrentWrites tests Welcome message ordering with concurrent writes
func TestWelcomeMessageStreaming_OrderingWithConcurrentWrites(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	config := defaultTestConfig()
	installationKey := []byte(test.RandomString(32))
	hpkePublicKey := []byte(test.RandomString(32))

	collector := newWelcomeMessageCollector(config.bufferSize)
	stream := setupWelcomeMessageStream(t, ctx, collector)

	// Start subscription
	startWelcomeMessageSubscription(t, svc, []*mlsv1.SubscribeWelcomeMessagesRequest_Filter{
		{InstallationKey: installationKey},
	}, stream)

	// Send welcome messages concurrently
	numGoroutines := 5
	messagesPerGoroutine := 8
	totalMessages := numGoroutines * messagesPerGoroutine

	sendWelcomeMessages(
		t,
		ctx,
		svc,
		installationKey,
		hpkePublicKey,
		numGoroutines,
		messagesPerGoroutine,
		"welcome",
	)

	// Wait for all messages
	err := collector.waitForCount(totalMessages, config.timeout)
	require.NoError(t, err, "Should receive all welcome messages within timeout")

	// Validate results
	messages := collector.getAllMessages()
	assert.GreaterOrEqual(t, len(messages), totalMessages, "Should receive all welcome messages")
	validateWelcomeMessageOrdering(t, messages)
	validateWelcomeMessageNoDuplicates(t, messages)

	for _, welcomeMessage := range messages {
		require.NotZero(t, welcomeMessage.GetV1().WelcomeMetadata)
		require.Equal(
			t,
			welcomeMessage.GetV1().WrapperAlgorithm,
			messageContentsProto.WelcomeWrapperAlgorithm_WELCOME_WRAPPER_ALGORITHM_XWING_MLKEM_768_DRAFT_6,
		)
	}
}

// TestWelcomeMessageStreaming_CursorConsistency tests cursor behavior for welcome messages
func TestWelcomeMessageStreaming_CursorConsistency(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	config := defaultTestConfig()
	installationKey := []byte(test.RandomString(32))
	hpkePublicKey := []byte(test.RandomString(32))

	// Pre-populate with welcome messages
	numInitialMessages := 15
	populateWelcomeMessages(
		t,
		ctx,
		svc,
		installationKey,
		hpkePublicKey,
		numInitialMessages,
		"initial_welcome",
	)

	// Get current messages to establish cursor
	resp, err := svc.QueryWelcomeMessages(ctx, &mlsv1.QueryWelcomeMessagesRequest{
		InstallationKey: installationKey,
		PagingInfo: &mlsv1.PagingInfo{
			Direction: mlsv1.SortDirection_SORT_DIRECTION_ASCENDING,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Messages, numInitialMessages)

	// Use cursor at position 5 (should receive messages 6-15 plus new messages)
	cursorPosition := resp.Messages[4].GetV1().Id
	expectedMessagesAfterCursor := numInitialMessages - 5

	collector := newWelcomeMessageCollector(config.bufferSize)
	stream := setupWelcomeMessageStream(t, ctx, collector)

	// Start subscription with cursor
	startWelcomeMessageSubscription(t, svc, []*mlsv1.SubscribeWelcomeMessagesRequest_Filter{
		{
			InstallationKey: installationKey,
			IdCursor:        cursorPosition,
		},
	}, stream)

	time.Sleep(200 * time.Millisecond)

	// Send additional welcome messages
	numNewMessages := 7
	populateWelcomeMessages(
		t,
		ctx,
		svc,
		installationKey,
		hpkePublicKey,
		numNewMessages,
		"new_welcome",
	)

	expectedTotalMessages := expectedMessagesAfterCursor + numNewMessages

	// Wait for all expected messages
	err = collector.waitForCount(expectedTotalMessages, config.timeout)
	require.NoError(t, err, "Should receive all welcome messages after cursor within timeout")

	// Validate results
	messages := collector.getAllMessages()
	assert.GreaterOrEqual(
		t,
		len(messages),
		expectedTotalMessages,
		"Should receive all welcome messages after cursor",
	)
	validateWelcomeMessagesAfterCursor(t, messages, cursorPosition)
	validateWelcomeMessageOrdering(t, messages)
}

// TestStreaming_ErrorHandling tests various error scenarios
func TestStreaming_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	t.Run("GroupMessages_InvalidGroupId", func(t *testing.T) {
		stream := mocks.NewMockMlsApi_SubscribeGroupMessagesServer(t)
		stream.EXPECT().
			SendHeader(metadata.New(map[string]string{"subscribed": "true"})).
			Return(nil)
		stream.EXPECT().Context().Return(ctx)

		err := svc.SubscribeGroupMessages(&mlsv1.SubscribeGroupMessagesRequest{
			Filters: []*mlsv1.SubscribeGroupMessagesRequest_Filter{
				{
					GroupId: nil, // Invalid group ID
				},
			},
		}, stream)

		// Should return an error for this invalid request
		assert.Error(t, err)
	})

	t.Run("WelcomeMessages_InvalidInstallationKey", func(t *testing.T) {
		stream := mocks.NewMockMlsApi_SubscribeWelcomeMessagesServer(t)
		stream.EXPECT().
			SendHeader(metadata.New(map[string]string{"subscribed": "true"})).
			Return(nil)
		stream.EXPECT().Context().Return(ctx)

		err := svc.SubscribeWelcomeMessages(&mlsv1.SubscribeWelcomeMessagesRequest{
			Filters: []*mlsv1.SubscribeWelcomeMessagesRequest_Filter{
				{
					InstallationKey: nil, // Invalid installation key
				},
			},
		}, stream)

		// Should return an error here
		assert.Error(t, err)
	})

	t.Run("GroupMessages_ContextCancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)

		stream := mocks.NewMockMlsApi_SubscribeGroupMessagesServer(t)
		stream.EXPECT().
			SendHeader(metadata.New(map[string]string{"subscribed": "true"})).
			Return(nil)
		stream.EXPECT().Context().Return(cancelCtx)

		testGroupId := []byte(test.RandomString(32))

		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel() // Cancel context after subscription starts
		}()

		err := svc.SubscribeGroupMessages(&mlsv1.SubscribeGroupMessagesRequest{
			Filters: []*mlsv1.SubscribeGroupMessagesRequest_Filter{
				{
					GroupId: testGroupId,
				},
			},
		}, stream)

		// Should return without error when context is cancelled
		assert.NoError(t, err)
	})

	t.Run("WelcomeMessages_ContextCancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)

		stream := mocks.NewMockMlsApi_SubscribeWelcomeMessagesServer(t)
		stream.EXPECT().
			SendHeader(metadata.New(map[string]string{"subscribed": "true"})).
			Return(nil)
		stream.EXPECT().Context().Return(cancelCtx)

		installationKey := []byte(test.RandomString(32))

		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel() // Cancel context after subscription starts
		}()

		err := svc.SubscribeWelcomeMessages(&mlsv1.SubscribeWelcomeMessagesRequest{
			Filters: []*mlsv1.SubscribeWelcomeMessagesRequest_Filter{
				{
					InstallationKey: installationKey,
				},
			},
		}, stream)

		// Should return without error when context is cancelled
		assert.NoError(t, err)
	})
}

// TestStreaming_MultipleFilters tests subscriptions with multiple filters
func TestStreaming_MultipleFilters(t *testing.T) {
	ctx := context.Background()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	config := defaultTestConfig()

	// Create multiple group IDs
	groupId1 := []byte(test.RandomString(32))
	groupId2 := []byte(test.RandomString(32))
	groupId3 := []byte(test.RandomString(32))

	collector := newMessageCollector(config.bufferSize)
	stream := setupGroupMessageStream(t, ctx, collector)

	// Start subscription with multiple filters
	startGroupMessageSubscription(t, svc, []*mlsv1.SubscribeGroupMessagesRequest_Filter{
		{GroupId: groupId1},
		{GroupId: groupId2},
		{GroupId: groupId3},
	}, stream)

	time.Sleep(config.subscriptionDelay)

	// Send messages to different groups
	groupIds := [][]byte{groupId1, groupId2, groupId3}
	messagesPerGroup := 5
	totalMessages := len(groupIds) * messagesPerGroup

	for i, groupId := range groupIds {
		validationSvc.ExpectedCalls = nil // Clear all previous mocks
		mockValidateGroupMessages(validationSvc, groupId)
		populateGroupMessages(
			t,
			ctx,
			svc,
			groupId,
			messagesPerGroup,
			fmt.Sprintf("group_%d_msg", i),
		)
	}

	// Wait for all messages
	err := collector.waitForCount(totalMessages, config.timeout)
	require.NoError(t, err, "Should receive messages from all groups within timeout")

	// Validate results
	messages := collector.getAllMessages()
	assert.GreaterOrEqual(
		t,
		len(messages),
		totalMessages,
		"Should receive messages from all groups",
	)

	// Verify messages from each group are represented
	groupCounts := make(map[string]int)
	for _, msg := range messages {
		groupId := string(msg.GetV1().GroupId)
		groupCounts[groupId]++
	}

	assert.Equal(t, len(groupIds), len(groupCounts), "Should receive messages from all groups")
	for groupId, count := range groupCounts {
		assert.GreaterOrEqual(t, count, messagesPerGroup,
			"Should receive all messages from group %s", groupId)
	}

	// Verify overall ordering (messages should be in sequence_id order)
	validateMessageOrdering(t, messages)
}

// TestStreaming_HighThroughput tests the streaming under high message volume
func TestStreaming_HighThroughput(t *testing.T) {
	ctx := context.Background()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	config := testConfig{
		timeout:           15 * time.Second,
		subscriptionDelay: 100 * time.Millisecond,
		bufferSize:        1000,
	}

	groupId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupId)

	collector := newMessageCollector(config.bufferSize)
	stream := setupGroupMessageStream(t, ctx, collector)

	// Start subscription
	startGroupMessageSubscription(t, svc, []*mlsv1.SubscribeGroupMessagesRequest_Filter{
		{GroupId: groupId},
	}, stream)

	time.Sleep(config.subscriptionDelay)

	// Send a large number of messages with high concurrency
	numGoroutines := 20
	messagesPerGoroutine := 25
	totalMessages := numGoroutines * messagesPerGoroutine

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineId int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msgData := fmt.Sprintf("high_throughput_%d_%d", goroutineId, j)
				_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
					Messages: []*mlsv1.GroupMessageInput{
						{
							Version: &mlsv1.GroupMessageInput_V1_{
								V1: &mlsv1.GroupMessageInput_V1{
									Data:       []byte(msgData),
									SenderHmac: []byte(fmt.Sprintf("hmac_%d_%d", goroutineId, j)),
									ShouldPush: true,
								},
							},
						},
					},
				})
				require.NoError(t, err)

				// Add some randomness to timing
				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Wait for all messages
	err := collector.waitForCount(totalMessages, config.timeout)
	require.NoError(t, err, "Should receive all high throughput messages within timeout")

	// Validate results
	messages := collector.getAllMessages()
	assert.GreaterOrEqual(
		t,
		len(messages),
		totalMessages,
		"Should receive all high throughput messages",
	)

	// Verify messages are in order and no duplicates
	seenIds := make(map[uint64]bool)
	for i, msg := range messages {
		id := msg.GetV1().Id
		assert.False(t, seenIds[id], "Should not have duplicate messages in high throughput test")
		seenIds[id] = true

		if i > 0 {
			assert.Greater(t, id, messages[i-1].GetV1().Id,
				"Messages should be in ascending order even under high throughput")
		}
	}
}

// TestStreaming_MessageIntegrity tests that message content is preserved correctly
func TestStreaming_MessageIntegrity(t *testing.T) {
	ctx := context.Background()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	config := defaultTestConfig()
	groupId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupId)

	// Create messages with various data patterns
	testMessages := []struct {
		data       string
		senderHmac string
		shouldPush bool
	}{
		{"simple_message", "hmac1", true},
		{"message_with_unicode_🚀", "hmac2", false},
		{"very_long_message_" + string(make([]byte, 1000)), "hmac4", true}, // Large message
		{"binary_data_\x00\x01\x02\x03", "hmac5", false},                   // Binary data
	}

	collector := newMessageCollector(len(testMessages))
	stream := setupGroupMessageStream(t, ctx, collector)

	// Start subscription
	startGroupMessageSubscription(t, svc, []*mlsv1.SubscribeGroupMessagesRequest_Filter{
		{GroupId: groupId},
	}, stream)

	time.Sleep(config.subscriptionDelay)

	// Send test messages
	for _, testMsg := range testMessages {
		_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
			Messages: []*mlsv1.GroupMessageInput{
				{
					Version: &mlsv1.GroupMessageInput_V1_{
						V1: &mlsv1.GroupMessageInput_V1{
							Data:       []byte(testMsg.data),
							SenderHmac: []byte(testMsg.senderHmac),
							ShouldPush: testMsg.shouldPush,
						},
					},
				},
			},
		})
		require.NoError(t, err)
	}

	// Wait for all messages
	err := collector.waitForCount(len(testMessages), config.timeout)
	require.NoError(t, err, "Should receive all integrity test messages within timeout")

	// Validate results
	messages := collector.getAllMessages()
	assert.Len(t, messages, len(testMessages), "Should receive all integrity test messages")

	// Verify message content integrity
	for i, msg := range messages {
		expectedMsg := testMessages[i]
		assert.Equal(t, []byte(expectedMsg.data), msg.GetV1().Data,
			"Message data should match exactly")
		assert.Equal(t, []byte(expectedMsg.senderHmac), msg.GetV1().SenderHmac,
			"Sender HMAC should match exactly")
		assert.Equal(t, expectedMsg.shouldPush, msg.GetV1().ShouldPush,
			"Should push flag should match exactly")
		assert.True(t, bytes.Equal(groupId, msg.GetV1().GroupId),
			"Group ID should match exactly")
	}
}
