package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mocks"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// fakeSubscribeStream is an in-test implementation of mlsv1.MlsApi_SubscribeServer that
// drives the bidirectional XIP-83 Subscribe handler: the test pushes client frames with
// send() (the handler reads them via Recv) and reads the frames the handler emits via Send.
type fakeSubscribeStream struct {
	ctx       context.Context
	incoming  chan *mlsv1.SubscribeRequest
	failCh    chan error    // inject a transport Recv error (not a clean half-close)
	blockSend chan struct{} // when set, Send blocks on it (a non-reading client)

	mu   sync.Mutex
	sent []*mlsv1.SubscribeResponse

	closeOnce sync.Once
}

func newFakeSubscribeStream(ctx context.Context) *fakeSubscribeStream {
	return &fakeSubscribeStream{
		ctx:      ctx,
		incoming: make(chan *mlsv1.SubscribeRequest, 64),
		failCh:   make(chan error, 1),
	}
}

// send queues a client -> server frame.
func (f *fakeSubscribeStream) send(req *mlsv1.SubscribeRequest) { f.incoming <- req }

// closeSend simulates the client half-closing its send direction (Recv -> io.EOF).
func (f *fakeSubscribeStream) closeSend() { f.closeOnce.Do(func() { close(f.incoming) }) }

// failRecv makes the next Recv return err, simulating a mid-stream transport failure.
func (f *fakeSubscribeStream) failRecv(err error) { f.failCh <- err }

// responses returns a snapshot copy of everything the handler has sent so far.
func (f *fakeSubscribeStream) responses() []*mlsv1.SubscribeResponse {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*mlsv1.SubscribeResponse, len(f.sent))
	copy(out, f.sent)
	return out
}

// --- mlsv1.MlsApi_SubscribeServer ---

func (f *fakeSubscribeStream) Send(resp *mlsv1.SubscribeResponse) error {
	f.mu.Lock()
	block := f.blockSend
	f.mu.Unlock()
	if block != nil {
		<-block
	}
	f.mu.Lock()
	f.sent = append(f.sent, resp)
	f.mu.Unlock()
	return nil
}

// freezeSends makes every subsequent Send block until the returned release func is called.
// Unlike setting blockSend directly, it is safe to call after Subscribe has started (the
// field is guarded by mu), so a test can quiesce the sender mid-stream to force frames to
// queue.
func (f *fakeSubscribeStream) freezeSends() (release func()) {
	ch := make(chan struct{})
	f.mu.Lock()
	f.blockSend = ch
	f.mu.Unlock()
	return func() { close(ch) }
}

func (f *fakeSubscribeStream) Recv() (*mlsv1.SubscribeRequest, error) {
	select {
	case err := <-f.failCh:
		return nil, err
	case req, ok := <-f.incoming:
		if !ok {
			return nil, io.EOF
		}
		return req, nil
	case <-f.ctx.Done():
		return nil, f.ctx.Err()
	}
}

func (f *fakeSubscribeStream) Context() context.Context     { return f.ctx }
func (f *fakeSubscribeStream) SendHeader(metadata.MD) error { return nil }
func (f *fakeSubscribeStream) SetHeader(metadata.MD) error  { return nil }
func (f *fakeSubscribeStream) SetTrailer(metadata.MD)       {}
func (f *fakeSubscribeStream) SendMsg(m interface{}) error  { return nil }
func (f *fakeSubscribeStream) RecvMsg(m interface{}) error  { return nil }

// --- request builders ---

// groupTopic / welcomeTopic build kind-prefixed wire topics (XIP-49 §3.3.2).
func groupTopic(groupId []byte) []byte {
	return append([]byte{topicKindGroupMessagesV1}, groupId...)
}

func welcomeTopic(installationKey []byte) []byte {
	return append([]byte{topicKindWelcomeMessagesV1}, installationKey...)
}

func addSub(topic []byte, cursor uint64) *mlsv1.SubscribeRequest_V1_Mutate_Subscription {
	return &mlsv1.SubscribeRequest_V1_Mutate_Subscription{Topic: topic, IdCursor: cursor}
}

func subReqMutate(m *mlsv1.SubscribeRequest_V1_Mutate) *mlsv1.SubscribeRequest {
	return &mlsv1.SubscribeRequest{
		Version: &mlsv1.SubscribeRequest_V1_{V1: &mlsv1.SubscribeRequest_V1{
			Request: &mlsv1.SubscribeRequest_V1_Mutate_{Mutate: m},
		}},
	}
}

func subReqPong(nonce uint64) *mlsv1.SubscribeRequest {
	return &mlsv1.SubscribeRequest{
		Version: &mlsv1.SubscribeRequest_V1_{V1: &mlsv1.SubscribeRequest_V1{
			Request: &mlsv1.SubscribeRequest_V1_Pong{Pong: &mlsv1.Pong{Nonce: nonce}},
		}},
	}
}

// --- response extractors (operate on a snapshot) ---

func groupMsgsFrom(resps []*mlsv1.SubscribeResponse) []*mlsv1.GroupMessage {
	var out []*mlsv1.GroupMessage
	for _, r := range resps {
		if msgs := r.GetV1().GetMessages(); msgs != nil {
			out = append(out, msgs.GetGroupMessages()...)
		}
	}
	return out
}

func welcomeMsgsFrom(resps []*mlsv1.SubscribeResponse) []*mlsv1.WelcomeMessage {
	var out []*mlsv1.WelcomeMessage
	for _, r := range resps {
		if msgs := r.GetV1().GetMessages(); msgs != nil {
			out = append(out, msgs.GetWelcomeMessages()...)
		}
	}
	return out
}

func hasStarted(resps []*mlsv1.SubscribeResponse) bool {
	for _, r := range resps {
		if r.GetV1().GetStarted() != nil {
			return true
		}
	}
	return false
}

// catchupCompletesFrom returns the echoed mutate_ids of every CatchupComplete, in order.
func catchupCompletesFrom(resps []*mlsv1.SubscribeResponse) []uint64 {
	var out []uint64
	for _, r := range resps {
		if cc := r.GetV1().GetCatchupComplete(); cc != nil {
			out = append(out, cc.GetMutateId())
		}
	}
	return out
}

func pingsFrom(resps []*mlsv1.SubscribeResponse) []uint64 {
	var out []uint64
	for _, r := range resps {
		if p := r.GetV1().GetPing(); p != nil {
			out = append(out, p.GetNonce())
		}
	}
	return out
}

func containsPong(resps []*mlsv1.SubscribeResponse, nonce uint64) bool {
	for _, r := range resps {
		if p := r.GetV1().GetPong(); p != nil && p.GetNonce() == nonce {
			return true
		}
	}
	return false
}

func groupMsgsWithData(msgs []*mlsv1.GroupMessage, data string) []*mlsv1.GroupMessage {
	var out []*mlsv1.GroupMessage
	for _, m := range msgs {
		if string(m.GetV1().GetData()) == data {
			out = append(out, m)
		}
	}
	return out
}

func welcomesForKey(msgs []*mlsv1.WelcomeMessage, key []byte) []*mlsv1.WelcomeMessage {
	var out []*mlsv1.WelcomeMessage
	for _, m := range msgs {
		if string(m.GetV1().GetInstallationKey()) == string(key) {
			out = append(out, m)
		}
	}
	return out
}

// frameIndex returns the index of the first response satisfying pred, or -1.
func frameIndex(resps []*mlsv1.SubscribeResponse, pred func(*mlsv1.SubscribeResponse) bool) int {
	for i, r := range resps {
		if pred(r) {
			return i
		}
	}
	return -1
}

// topicsLiveHas reports whether the frame is a TopicsLive containing the wire topic.
func topicsLiveHas(r *mlsv1.SubscribeResponse, wireTopic []byte) bool {
	for _, t := range r.GetV1().GetTopicsLive().GetTopics() {
		if string(t) == string(wireTopic) {
			return true
		}
	}
	return false
}

// lastFrameWithGroupMsgs returns the index of the last Messages frame carrying a message
// for the given group, or -1.
func lastFrameWithGroupMsgs(resps []*mlsv1.SubscribeResponse, groupId []byte) int {
	last := -1
	for i, r := range resps {
		for _, m := range r.GetV1().GetMessages().GetGroupMessages() {
			if string(m.GetV1().GetGroupId()) == string(groupId) {
				last = i
			}
		}
	}
	return last
}

// lastFrameWithWelcomes is the welcome-topic analogue of lastFrameWithGroupMsgs.
func lastFrameWithWelcomes(resps []*mlsv1.SubscribeResponse, key []byte) int {
	last := -1
	for i, r := range resps {
		for _, m := range r.GetV1().GetMessages().GetWelcomeMessages() {
			if string(m.GetV1().GetInstallationKey()) == string(key) {
				last = i
			}
		}
	}
	return last
}

// --- test helpers ---

// waitForResponses polls the stream until pred is satisfied, failing the test on timeout.
func waitForResponses(
	t *testing.T,
	stream *fakeSubscribeStream,
	timeout time.Duration,
	desc string,
	pred func([]*mlsv1.SubscribeResponse) bool,
) []*mlsv1.SubscribeResponse {
	t.Helper()
	deadline := time.After(timeout)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		resps := stream.responses()
		if pred(resps) {
			return resps
		}
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for %s", desc)
			return resps
		case <-tick.C:
		}
	}
}

// publishGroup sends one group message with the exact data, (re)mocking validation for
// the target group first (a single mock returns a fixed group id, so it must be reset
// whenever the target group changes).
func publishGroup(
	t *testing.T,
	ctx context.Context,
	svc *Service,
	validationSvc *mocks.MockMLSValidationService,
	groupId []byte,
	data string,
) {
	t.Helper()
	validationSvc.ExpectedCalls = nil
	mockValidateGroupMessages(validationSvc, groupId)
	_, err := svc.SendGroupMessages(ctx, &mlsv1.SendGroupMessagesRequest{
		Messages: []*mlsv1.GroupMessageInput{
			{
				Version: &mlsv1.GroupMessageInput_V1_{
					V1: &mlsv1.GroupMessageInput_V1{
						Data:       []byte(data),
						SenderHmac: []byte("hmac"),
						ShouldPush: true,
					},
				},
			},
		},
	})
	require.NoError(t, err)
}

// TestSubscribe_CatchUpThenLiveNoDuplicates exercises the live-gate: history is sent
// before live, live messages published while catch-up is in flight are buffered and then
// flushed, and nothing is duplicated or reordered.
func TestSubscribe_CatchUpThenLiveNoDuplicates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupId)

	const history = 25
	populateGroupMessages(t, ctx, svc, groupId, history, "hist")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// Subscribe from the beginning, then immediately publish live messages so they race
	// the catch-up.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupId), 0)},
		MutateId: 1,
	}))
	const live = 10
	populateGroupMessages(t, ctx, svc, groupId, live, "live")

	total := history + live
	resps := waitForResponses(
		t,
		stream,
		10*time.Second,
		fmt.Sprintf("%d group messages", total),
		func(rs []*mlsv1.SubscribeResponse) bool { return len(groupMsgsFrom(rs)) >= total },
	)

	require.True(t, hasStarted(resps), "Started must be sent")
	waitForResponses(
		t,
		stream,
		5*time.Second,
		"CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(catchupCompletesFrom(rs)) >= 1
		},
	)

	msgs := groupMsgsFrom(resps)
	require.GreaterOrEqual(t, len(msgs), total)
	validateMessageOrdering(t, msgs)
	validateNoDuplicates(t, msgs)

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_MutateRemoveStopsDelivery verifies that removing a group in place stops
// delivery for that group while a co-subscribed group keeps flowing on the same stream.
func TestSubscribe_MutateRemoveStopsDelivery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))
	sentinelId := []byte(test.RandomString(32))

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// Subscribe both groups at the live edge (they have no history yet).
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupId), 0),
			addSub(groupTopic(sentinelId), 0),
		},
		MutateId: 1,
	}))
	waitForResponses(
		t,
		stream,
		5*time.Second,
		"CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(catchupCompletesFrom(rs)) >= 1
		},
	)

	// The group delivers before removal.
	publishGroup(t, ctx, svc, validationSvc, groupId, "g-live-1")
	waitForResponses(
		t,
		stream,
		5*time.Second,
		"first group message",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsFrom(rs), "g-live-1")) >= 1
		},
	)

	// Remove the group, then Ping. The single main loop processes the remove before the
	// ping, so observing the Pong proves sub.Remove(group) has executed.
	stream.send(
		subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{Removes: [][]byte{groupTopic(groupId)}}),
	)
	stream.send(
		&mlsv1.SubscribeRequest{Version: &mlsv1.SubscribeRequest_V1_{V1: &mlsv1.SubscribeRequest_V1{
			Request: &mlsv1.SubscribeRequest_V1_Ping{Ping: &mlsv1.Ping{Nonce: 42}},
		}}},
	)
	waitForResponses(
		t,
		stream,
		5*time.Second,
		"pong(42)",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return containsPong(rs, 42)
		},
	)

	// Publish to the removed group (must be dropped) then to the still-subscribed
	// sentinel. The dbWorker dispatches strictly in id order, so once the sentinel
	// message arrives the removed group's message has already been processed and dropped.
	publishGroup(t, ctx, svc, validationSvc, groupId, "g-after-remove")
	publishGroup(t, ctx, svc, validationSvc, sentinelId, "sentinel-1")
	resps := waitForResponses(
		t,
		stream,
		5*time.Second,
		"sentinel message",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsFrom(rs), "sentinel-1")) >= 1
		},
	)

	all := groupMsgsFrom(resps)
	require.Empty(t, groupMsgsWithData(all, "g-after-remove"), "removed group must not deliver")
	require.Len(
		t,
		groupMsgsWithData(all, "g-live-1"),
		1,
		"pre-removal message should be delivered exactly once",
	)

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_PingPongKeepsStreamAlive verifies the WebSocket-style heartbeat: the node
// Pings when idle and the stream stays open as long as the client Pongs.
func TestSubscribe_PingPongKeepsStreamAlive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.pingInterval = 200 * time.Millisecond
	svc.pongDeadline = 2 * time.Second

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	for round := 0; round < 3; round++ {
		resps := waitForResponses(
			t,
			stream,
			3*time.Second,
			fmt.Sprintf("server ping #%d", round+1),
			func(rs []*mlsv1.SubscribeResponse) bool {
				return len(pingsFrom(rs)) > round
			},
		)
		nonces := pingsFrom(resps)
		stream.send(subReqPong(nonces[round]))
	}

	// The stream must still be open (the handler has not returned).
	select {
	case err := <-errCh:
		t.Fatalf("stream closed unexpectedly: %v", err)
	default:
	}

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_ReapsOnMissedPong verifies that a peer that never answers a Ping is reaped.
func TestSubscribe_ReapsOnMissedPong(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.pingInterval = 150 * time.Millisecond
	svc.pongDeadline = 150 * time.Millisecond

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// The node Pings when idle; we never Pong.
	waitForResponses(
		t,
		stream,
		2*time.Second,
		"server ping",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(pingsFrom(rs)) >= 1
		},
	)

	select {
	case err := <-errCh:
		require.Equal(
			t,
			codes.DeadlineExceeded,
			status.Code(err),
			"missed pong should reap with DeadlineExceeded",
		)
	case <-time.After(3 * time.Second):
		t.Fatal("stream was not reaped after a missed pong")
	}
}

// TestSubscribe_MultiplexesMultipleIdentities is the herald-multiplexer case: two
// installations' welcomes and a group are all subscribed on one stream, caught up and
// streamed live, each routed to the right topic with no cross-contamination or dupes.
func TestSubscribe_MultiplexesMultipleIdentities(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	instA := []byte(test.RandomString(32))
	instB := []byte(test.RandomString(32))
	hpke := []byte(test.RandomString(32))
	groupId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupId)

	// History for both installations and the group.
	populateWelcomeMessages(t, ctx, svc, instA, hpke, 4, "welA")
	populateWelcomeMessages(t, ctx, svc, instB, hpke, 6, "welB")
	populateGroupMessages(t, ctx, svc, groupId, 5, "grp")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupId), 0),
			addSub(welcomeTopic(instA), 0),
			addSub(welcomeTopic(instB), 0),
		},
		MutateId: 1,
	}))

	// Catch-up delivers each identity's history, correctly attributed.
	waitForResponses(
		t,
		stream,
		10*time.Second,
		"all history",
		func(rs []*mlsv1.SubscribeResponse) bool {
			w := welcomeMsgsFrom(rs)
			return len(welcomesForKey(w, instA)) >= 4 &&
				len(welcomesForKey(w, instB)) >= 6 &&
				len(groupMsgsFrom(rs)) >= 5
		},
	)

	// Live on both welcome topics and the group, all over the one stream.
	populateWelcomeMessages(t, ctx, svc, instA, hpke, 2, "welA-live")
	populateWelcomeMessages(t, ctx, svc, instB, hpke, 3, "welB-live")
	publishGroup(t, ctx, svc, validationSvc, groupId, "grp-live")

	resps := waitForResponses(
		t,
		stream,
		10*time.Second,
		"all live",
		func(rs []*mlsv1.SubscribeResponse) bool {
			w := welcomeMsgsFrom(rs)
			return len(welcomesForKey(w, instA)) >= 6 &&
				len(welcomesForKey(w, instB)) >= 9 &&
				len(groupMsgsWithData(groupMsgsFrom(rs), "grp-live")) >= 1
		},
	)

	w := welcomeMsgsFrom(resps)
	welA := welcomesForKey(w, instA)
	welB := welcomesForKey(w, instB)
	require.Len(t, welA, 6, "installation A should receive its history + live, nothing else")
	require.Len(t, welB, 9, "installation B should receive its history + live, nothing else")
	validateWelcomeMessageOrdering(t, welA)
	validateWelcomeMessageOrdering(t, welB)
	validateWelcomeMessageNoDuplicates(t, welA)
	validateWelcomeMessageNoDuplicates(t, welB)

	g := groupMsgsFrom(resps)
	validateMessageOrdering(t, g)
	validateNoDuplicates(t, g)

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_TopicsLiveMarksLiveBoundary verifies the live-boundary signals: TopicsLive
// is emitted per opened topic AFTER that topic's history (so every later frame for the
// topic is live tail), and each Mutate that adds subscriptions is a catch-up wave that
// ends with its own CatchupComplete — echoing the Mutate's mutate_id — after the wave's
// last marker, including waves started mid-stream.
func TestSubscribe_TopicsLiveMarksLiveBoundary(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	instA := []byte(test.RandomString(32))
	hpke := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)

	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")
	populateWelcomeMessages(t, ctx, svc, instA, hpke, 3, "welA")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
			addSub(welcomeTopic(instA), 0),
		},
		MutateId: 7,
	}))

	groupMarkerPred := func(r *mlsv1.SubscribeResponse) bool {
		return topicsLiveHas(r, groupTopic(groupA))
	}
	welcomeMarkerPred := func(r *mlsv1.SubscribeResponse) bool {
		return topicsLiveHas(r, welcomeTopic(instA))
	}
	resps := waitForResponses(
		t,
		stream,
		10*time.Second,
		"TopicsLive for both topics + CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return frameIndex(rs, groupMarkerPred) >= 0 &&
				frameIndex(rs, welcomeMarkerPred) >= 0 &&
				len(catchupCompletesFrom(rs)) >= 1
		},
	)

	groupMarker := frameIndex(resps, groupMarkerPred)
	welcomeMarker := frameIndex(resps, welcomeMarkerPred)
	catchUpComplete := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		return r.GetV1().GetCatchupComplete() != nil
	})
	require.Equal(
		t,
		[]uint64{7},
		catchupCompletesFrom(resps),
		"the wave's CatchupComplete must echo its Mutate's id",
	)

	// Each topic's full history lands strictly before its marker, and both markers
	// precede the wave's CatchupComplete.
	lastGroupHist := lastFrameWithGroupMsgs(resps, groupA)
	lastWelcomeHist := lastFrameWithWelcomes(resps, instA)
	require.GreaterOrEqual(t, lastGroupHist, 0, "group history must be delivered")
	require.GreaterOrEqual(t, lastWelcomeHist, 0, "welcome history must be delivered")
	require.Greater(t, groupMarker, lastGroupHist, "group marker must follow group history")
	require.Greater(
		t,
		welcomeMarker,
		lastWelcomeHist,
		"welcome marker must follow welcome history",
	)
	require.Greater(t, catchUpComplete, groupMarker)
	require.Greater(t, catchUpComplete, welcomeMarker)

	// Everything after the marker is live tail: a message published now lands after it.
	publishGroup(t, ctx, svc, validationSvc, groupA, "liveA")
	resps = waitForResponses(
		t,
		stream,
		5*time.Second,
		"live message after marker",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsFrom(rs), "liveA")) >= 1
		},
	)
	liveIdx := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		return len(groupMsgsWithData(r.GetV1().GetMessages().GetGroupMessages(), "liveA")) > 0
	})
	require.Greater(t, liveIdx, groupMarker, "live frames must follow the marker")

	// A topic added mid-stream is its own catch-up wave: it gets a marker after its
	// history — the signal the client-side fan-out routes to whoever cares — and the
	// wave ends with its own CatchupComplete (echoing its mutate_id) after that marker.
	groupB := []byte(test.RandomString(32))
	for i := 0; i < 3; i++ {
		publishGroup(t, ctx, svc, validationSvc, groupB, fmt.Sprintf("histB-%d", i))
	}
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupB), 0)},
		MutateId: 8,
	}))
	markerBPred := func(r *mlsv1.SubscribeResponse) bool {
		return topicsLiveHas(r, groupTopic(groupB))
	}
	resps = waitForResponses(
		t,
		stream,
		10*time.Second,
		"TopicsLive + CatchupComplete for the mid-stream wave",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return frameIndex(rs, markerBPred) >= 0 && len(catchupCompletesFrom(rs)) >= 2
		},
	)
	require.Len(
		t,
		groupMsgsWithData(groupMsgsFrom(resps), "histB-0"),
		1,
		"mid-stream add must deliver its history",
	)
	markerB := frameIndex(resps, markerBPred)
	require.Greater(
		t,
		markerB,
		lastFrameWithGroupMsgs(resps, groupB),
		"mid-stream marker must follow that topic's history",
	)
	require.Equal(
		t,
		[]uint64{7, 8},
		catchupCompletesFrom(resps),
		"each adding mutate is a wave whose CatchupComplete echoes its mutate_id",
	)
	lastCatchUpComplete := -1
	for i, r := range resps {
		if r.GetV1().GetCatchupComplete() != nil {
			lastCatchUpComplete = i
		}
	}
	require.Greater(
		t,
		lastCatchUpComplete,
		markerB,
		"the wave's CatchupComplete must follow its marker",
	)

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_HistoryOnlyDeliversNoLive verifies that a history_only Mutate catches its
// topics up — history, marker, wave CatchupComplete — without ever registering them for
// live delivery.
func TestSubscribe_HistoryOnlyDeliversNoLive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	sentinelId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// Wave 1: a live sentinel subscription. Wave 2: groupA, history only.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(sentinelId), 0),
		},
		MutateId: 1,
	}))
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
		},
		HistoryOnly: true,
		MutateId:    2,
	}))

	markerAPred := func(r *mlsv1.SubscribeResponse) bool {
		return topicsLiveHas(r, groupTopic(groupA))
	}
	waitForResponses(
		t,
		stream,
		10*time.Second,
		"history, marker, and both waves' CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsFrom(rs)) >= 5 &&
				frameIndex(rs, markerAPred) >= 0 &&
				len(catchupCompletesFrom(rs)) >= 2
		},
	)

	// Publish live to the history-only topic (must NOT deliver), then to the live
	// sentinel. The dbWorker dispatches strictly in id order, so once the sentinel
	// arrives the history-only group's message has already been (not) routed.
	publishGroup(t, ctx, svc, validationSvc, groupA, "a-live")
	publishGroup(t, ctx, svc, validationSvc, sentinelId, "sentinel-live")
	resps := waitForResponses(
		t,
		stream,
		5*time.Second,
		"sentinel live message",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsFrom(rs), "sentinel-live")) >= 1
		},
	)
	require.Empty(
		t,
		groupMsgsWithData(groupMsgsFrom(resps), "a-live"),
		"history_only topics must not receive live delivery",
	)

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_HistoryOnlyOnLiveTopicRejected verifies that a history_only add targeting a
// topic already live on the same stream is rejected with InvalidArgument. There is one cursor
// floor per topic, so a one-shot bounded read on a tailed topic is contradictory — and on the
// replay path (cursor below the floor) the old code would dropTopic without re-registering,
// silently severing a live subscription the client was still tailing.
func TestSubscribe_HistoryOnlyOnLiveTopicRejected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// Subscribe groupA live and wait until its catch-up completes — the cursor floor is now
	// above 0, so the history_only re-add below lands on the replay path.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 1,
	}))
	waitForResponses(
		t,
		stream,
		10*time.Second,
		"groupA live catch-up complete",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return frameIndex(rs, func(r *mlsv1.SubscribeResponse) bool {
				return topicsLiveHas(r, groupTopic(groupA))
			}) >= 0 && len(catchupCompletesFrom(rs)) >= 1
		},
	)

	// A history_only add for the same, already-live topic (cursor 0, below the floor) is
	// contradictory and must be rejected rather than silently severing the live tail.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
		},
		HistoryOnly: true,
		MutateId:    2,
	}))

	select {
	case err := <-errCh:
		require.Equal(
			t,
			codes.InvalidArgument,
			status.Code(err),
			"history_only targeting an already-live topic must be rejected",
		)
	case <-time.After(5 * time.Second):
		t.Fatal("history_only add on a live topic was not rejected")
	}
}

// TestSubscribe_AddOverInflightHistoryOnlyRejected verifies that a second add for a topic whose
// history_only catch-up is still in flight is rejected: a competing wave would re-deliver history
// the first wave already sent (and orphan it). The gate holds the history_only wave mid-fetch.
func TestSubscribe_AddOverInflightHistoryOnlyRejected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "hist")
	gate := make(chan struct{})
	defer close(gate)
	svc.readOnlyStore = &fakeReadStore{ReadMlsStore: svc.readOnlyStore, groupGate: gate}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// history_only add: its catch-up wave blocks on the gate, staying in flight.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
		},
		HistoryOnly: true,
		MutateId:    1,
	}))
	// A second add for the same topic while that wave is in flight must be rejected.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 2,
	}))

	select {
	case err := <-errCh:
		require.Equal(t, codes.InvalidArgument, status.Code(err),
			"an add overlapping an in-flight history_only wave must be rejected")
	case <-time.After(5 * time.Second):
		t.Fatal("overlapping add was not rejected")
	}
}

// TestSubscribe_HalfCloseDrainsCatchUpThenCloses is the bounded catch-up ("catchUpOnce")
// shape end to end: Mutate{history_only} then immediately half-close; the server finishes
// the wave — history, marker, CatchupComplete — and then closes the stream itself.
func TestSubscribe_HalfCloseDrainsCatchUpThenCloses(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 25, "histA")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
		},
		HistoryOnly: true,
		MutateId:    99,
	}))
	stream.closeSend()

	select {
	case err := <-errCh:
		require.NoError(t, err, "the drained stream must close cleanly")
	case <-time.After(10 * time.Second):
		t.Fatal("server did not close the stream after draining the wave")
	}

	resps := stream.responses()
	require.Len(t, groupMsgsFrom(resps), 25, "the full history must be delivered before close")
	markerA := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		return topicsLiveHas(r, groupTopic(groupA))
	})
	require.Greater(
		t,
		markerA,
		lastFrameWithGroupMsgs(resps, groupA),
		"marker must follow the history",
	)
	require.Equal(
		t,
		[]uint64{99},
		catchupCompletesFrom(resps),
		"exactly the one wave's CatchupComplete, echoing its mutate_id",
	)
}

// TestSubscribe_StalePongDoesNotKeepStreamAlive verifies a Pong must echo the outstanding
// ping nonce: a stale or unsolicited Pong cannot suppress the missed-pong reap.
func TestSubscribe_StalePongDoesNotKeepStreamAlive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.pingInterval = 150 * time.Millisecond
	svc.pongDeadline = 300 * time.Millisecond

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// Continuously answer EVERY server ping with the WRONG nonce. With the nonce check the
	// outstanding ping is never satisfied, so the stream is reaped on schedule; without it,
	// each wrong Pong would clear awaitingPong and the stream would live forever.
	done := make(chan struct{})
	defer close(done)
	go func() {
		sent := 0
		for {
			select {
			case <-done:
				return
			default:
			}
			pings := pingsFrom(stream.responses())
			for ; sent < len(pings); sent++ {
				stream.send(subReqPong(pings[sent] + 100000)) // never the real nonce
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case err := <-errCh:
		require.Equal(
			t,
			codes.DeadlineExceeded,
			status.Code(err),
			"a client that only ever answers with the wrong nonce must still be reaped",
		)
	case <-time.After(3 * time.Second):
		t.Fatal("stream not reaped despite only wrong-nonce Pongs")
	}
}

// TestSubscribe_RecvErrorFailsStream verifies a transport Recv error fails the RPC rather
// than being mistaken for a clean half-close (which would report success).
func TestSubscribe_RecvErrorFailsStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	waitForResponses(
		t,
		stream,
		5*time.Second,
		"Started",
		func(rs []*mlsv1.SubscribeResponse) bool { return hasStarted(rs) },
	)
	stream.failRecv(status.Error(codes.Internal, "connection reset"))

	select {
	case err := <-errCh:
		require.Equal(
			t,
			codes.Unavailable,
			status.Code(err),
			"a transport Recv error must fail the stream, not return nil",
		)
	case <-time.After(5 * time.Second):
		t.Fatal("stream did not fail after a transport Recv error")
	}
}

// TestSubscribe_DuplicateAddsDeduped verifies a topic repeated within one Mutate's adds is
// collapsed: it is announced live (and counted) exactly once, not once per occurrence.
func TestSubscribe_DuplicateAddsDeduped(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
			addSub(groupTopic(groupA), 0), // same topic, twice
		},
		MutateId: 5,
	}))
	resps := waitForResponses(
		t,
		stream,
		10*time.Second,
		"history + CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsFrom(rs)) >= 5 && len(catchupCompletesFrom(rs)) >= 1
		},
	)

	liveTopicCount := 0
	for _, r := range resps {
		for _, tp := range r.GetV1().GetTopicsLive().GetTopics() {
			if string(tp) == string(groupTopic(groupA)) {
				liveTopicCount++
			}
		}
	}
	require.Equal(
		t,
		1,
		liveTopicCount,
		"a duplicated add must announce the topic live exactly once",
	)
	require.Len(t, groupMsgsFrom(resps), 5, "a duplicated add must not duplicate history")
	require.Equal(t, []uint64{5}, catchupCompletesFrom(resps), "one wave, one CatchupComplete")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_ReplayAfterRemove verifies removing a topic clears its cursor floor, so a
// later re-add replays the history again (XIP-83 replay-after-remove) — and that each
// removes-only Mutate is acked with an immediate CatchupComplete echoing its mutate_id,
// including mutate_id 0.
func TestSubscribe_ReplayAfterRemove(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 1,
	}))
	waitForResponses(
		t,
		stream,
		10*time.Second,
		"first catch-up",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsFrom(rs)) >= 5 && len(catchupCompletesFrom(rs)) >= 1
		},
	)

	// Remove, then re-add from cursor 0: the cleared floor lets the history replay. The
	// removes-only Mutate spawns no wave, so it is acked immediately, echoing its id.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Removes:  [][]byte{groupTopic(groupA)},
		MutateId: 5,
	}))
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 2,
	}))
	resps := waitForResponses(
		t,
		stream,
		10*time.Second,
		"replayed catch-up (wave 2)",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 3 },
	)
	require.Len(t, groupMsgsFrom(resps), 10, "remove + re-add must replay the full history")
	require.Equal(t, []uint64{1, 5, 2}, catchupCompletesFrom(resps))

	// A removes-only Mutate with mutate_id 0 is valid (0 only rejects adds) and is acked
	// with a CatchupComplete echoing 0.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Removes: [][]byte{groupTopic(groupA)},
	}))
	resps = waitForResponses(
		t,
		stream,
		5*time.Second,
		"ack for the mutate_id-0 remove",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 4 },
	)
	require.Equal(t, []uint64{1, 5, 2, 0}, catchupCompletesFrom(resps))

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_CursorAboveInt64ReturnsNoHistory verifies a cursor above MaxInt64 clamps to
// "no rows" rather than wrapping negative and replaying the whole history.
func TestSubscribe_CursorAboveInt64ReturnsNoHistory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), math.MaxUint64),
		},
		MutateId: 1,
	}))
	resps := waitForResponses(
		t,
		stream,
		10*time.Second,
		"CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 },
	)
	require.Empty(
		t,
		groupMsgsFrom(resps),
		"a cursor above MaxInt64 must return no rows, not replay history",
	)

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_InboundFramesDoNotSuppressLivenessPing verifies that inbound client frames
// which produce no response (here, unsolicited Pongs) do not defer the liveness Ping, so a
// peer that sends but never reads is still probed and reaped.
func TestSubscribe_InboundFramesDoNotSuppressLivenessPing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.pingInterval = 200 * time.Millisecond
	svc.pongDeadline = 2 * time.Second

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	waitForResponses(t, stream, 5*time.Second, "Started",
		func(rs []*mlsv1.SubscribeResponse) bool { return hasStarted(rs) })

	// The liveness Ping probes the client's RECEIVE path and is driven by SEND-side idleness
	// only. A client streaming response-free frames (unsolicited Pongs — every Mutate now
	// yields a CatchupComplete) must NOT suppress it — otherwise a peer that sends but never
	// reads could never be reaped.
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
			stream.send(subReqPong(999999)) // never a nonce the server is waiting on
			time.Sleep(40 * time.Millisecond)
		}
	}()

	waitForResponses(t, stream, 3*time.Second, "server ping despite steady inbound traffic",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(pingsFrom(rs)) >= 1 })
}

// TestSubscribe_GroupAndWelcomeSameIdentifierNoCollision verifies a group and a welcome that
// share an identifier are tracked independently: neither suppresses the other when one
// advances its high-water mark (the per-topic key is kind-distinct).
func TestSubscribe_GroupAndWelcomeSameIdentifierNoCollision(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	id := []byte(test.RandomString(32)) // used as BOTH a group id and an installation key
	hpke := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, id)

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(id), 0),
			addSub(welcomeTopic(id), 0),
		},
		MutateId: 1,
	}))
	waitForResponses(
		t,
		stream,
		5*time.Second,
		"CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 },
	)

	// Advance the welcome topic's high-water mark first; then publish a group message whose
	// (independent) id may be lower. With a shared key the group message would be deduped
	// away; with kind-distinct keys it is delivered.
	populateWelcomeMessages(t, ctx, svc, id, hpke, 3, "wel")
	publishGroup(t, ctx, svc, validationSvc, id, "grp-1")

	resps := waitForResponses(
		t,
		stream,
		5*time.Second,
		"the welcomes and the group message",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(welcomesForKey(welcomeMsgsFrom(rs), id)) >= 3 &&
				len(groupMsgsWithData(groupMsgsFrom(rs), "grp-1")) >= 1
		},
	)
	require.Len(
		t,
		groupMsgsWithData(groupMsgsFrom(resps), "grp-1"),
		1,
		"the group message must not be suppressed by the same-identifier welcome topic",
	)
	require.GreaterOrEqual(t, len(welcomesForKey(welcomeMsgsFrom(resps), id)), 3)

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestQueryGroupMessagesWaveScan_CursorAboveInt64ReturnsNothing exercises clampCursor
// directly: the Subscribe-level test is masked by the writer's high-water mark, so guard
// the store here. The ceiling above MaxInt64 clamps too (to "everything"), so the empty
// result is attributable to the cursor alone.
func TestQueryGroupMessagesWaveScan_CursorAboveInt64ReturnsNothing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "hist")

	msgs, err := svc.readOnlyStore.QueryGroupMessagesWaveScan(
		ctx,
		[]mlsstore.GroupCatchup{{GroupID: groupA, IdCursor: math.MaxUint64}},
		0,
		math.MaxUint64,
		50,
	)
	require.NoError(t, err)
	require.Empty(t, msgs, "a cursor above MaxInt64 must clamp to no rows, not wrap negative")
}

// TestSubscribe_CurrentCursorWaveSkipsScan pins the wave short-circuit: a wave whose
// every cursor already sits at the ceiling can replay nothing, so it must complete —
// TopicsLive and all — without a single wave-scan query. Reconnecting clients are
// overwhelmingly in this state, and before the short-circuit each such wave still cost
// a DB round trip. A second, stale-cursor wave on the same stream then proves the
// counter observes real scans (the zero above is not vacuous).
func TestSubscribe_CurrentCursorWaveSkipsScan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	groupB := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")
	validationSvc.ExpectedCalls = nil
	mockValidateGroupMessages(validationSvc, groupB)
	populateGroupMessages(t, ctx, svc, groupB, 3, "histB")

	// The ceiling is the global newest id, so a "current" topic must be cursored there.
	latest, err := svc.readOnlyStore.Queries().GetLatestGroupMessageID(ctx)
	require.NoError(t, err)

	fake := &fakeReadStore{ReadMlsStore: svc.readOnlyStore}
	svc.readOnlyStore = fake

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), uint64(latest)),
		},
		MutateId: 1,
	}))
	resps := waitForResponses(t, stream, 10*time.Second, "current-cursor wave to go live",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return frameIndex(rs, func(r *mlsv1.SubscribeResponse) bool {
				return topicsLiveHas(r, groupTopic(groupA))
			}) >= 0
		})
	require.Zero(t, fake.groupScans.Load(), "a current-cursor wave must not query the store")
	require.Empty(t, groupMsgsFrom(resps), "a current-cursor wave has nothing to replay")

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupB), 0),
		},
		MutateId: 2,
	}))
	waitForResponses(t, stream, 10*time.Second, "stale-cursor wave replay",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(groupMsgsFrom(rs)) >= 3 })
	require.Positive(t, fake.groupScans.Load(), "a stale-cursor wave must scan")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_WaveRotatesInBatches pins the batch rotation with shrunken
// tunables (batch of 2 topics, page of 2 rows) over three groups: no scan query
// ever carries more than a batch of topics, a full page rotates its batch to
// the back of the queue (so a topic is re-queried across turns rather than
// replayed in one burst), and the replay still delivers every message exactly
// once, ascending within each topic — the per-topic guarantee the writer's
// high-water dedup relies on now that turns are not globally id-ordered.
func TestSubscribe_WaveRotatesInBatches(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.scanBatchTopics = 2
	svc.scanTopicPageLimit = 2

	groups := [][]byte{
		[]byte(test.RandomString(32)),
		[]byte(test.RandomString(32)),
		[]byte(test.RandomString(32)),
	}
	for i, g := range groups {
		validationSvc.ExpectedCalls = nil
		mockValidateGroupMessages(validationSvc, g)
		populateGroupMessages(t, ctx, svc, g, 3, fmt.Sprintf("hist%d", i))
	}

	fake := &fakeReadStore{ReadMlsStore: svc.readOnlyStore}
	svc.readOnlyStore = fake

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	adds := make([]*mlsv1.SubscribeRequest_V1_Mutate_Subscription, len(groups))
	for i, g := range groups {
		adds[i] = addSub(groupTopic(g), 0)
	}
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{Adds: adds, MutateId: 1}))

	resps := waitForResponses(t, stream, 10*time.Second, "full rotated replay",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(groupMsgsFrom(rs)) >= 9 })
	got := groupMsgsFrom(resps)
	require.Len(t, got, 9, "every message exactly once, no duplicates from rotation")

	perTopic := map[string][]uint64{}
	for _, m := range got {
		key := string(m.GetV1().GetGroupId())
		perTopic[key] = append(perTopic[key], m.GetV1().GetId())
	}
	require.Len(t, perTopic, 3, "all three topics replayed")
	for topic, ids := range perTopic {
		require.Len(t, ids, 3, "topic %x complete", topic)
		for i := 1; i < len(ids); i++ {
			require.Greater(t, ids[i], ids[i-1], "topic %x ids must ascend", topic)
		}
	}

	fake.scanMu.Lock()
	calls := fake.groupScanFilters
	fake.scanMu.Unlock()
	seenTurns := map[string]int{}
	for _, call := range calls {
		require.LessOrEqual(t, len(call), 2, "a scan query must never carry more than a batch")
		for _, topic := range call {
			seenTurns[topic]++
		}
	}
	rotated := false
	for _, n := range seenTurns {
		if n >= 2 {
			rotated = true
		}
	}
	require.True(t, rotated, "at least one topic must be re-queried across turns (rotation)")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_MultiPageCatchUp exercises catch-up pagination across more than
// a topic's per-turn limit: the cursor continuation plus wave completion on the final
// short page, with exactly one TopicsLive and one CatchupComplete.
func TestSubscribe_MultiPageCatchUp(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.scanTopicPageLimit = 50 // small pages so a modest history spans several

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	const history = 50*2 + 17
	populateGroupMessages(t, ctx, svc, groupA, history, "hist")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 1,
	}))
	resps := waitForResponses(t, stream, 15*time.Second,
		fmt.Sprintf("%d history messages + CatchupComplete", history),
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsFrom(rs)) >= history && len(catchupCompletesFrom(rs)) >= 1
		})

	msgs := groupMsgsFrom(resps)
	require.Len(t, msgs, history, "every page of history delivered exactly once")
	validateMessageOrdering(t, msgs)
	validateNoDuplicates(t, msgs)
	require.Equal(t, []uint64{1}, catchupCompletesFrom(resps))
	markerCount := 0
	for _, r := range resps {
		if topicsLiveHas(r, groupTopic(groupA)) {
			markerCount++
		}
	}
	require.Equal(t, 1, markerCount, "a multi-page topic opens (and is announced) exactly once")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_NonZeroStartingCursor verifies the id>cursor / seed-floor boundary: a
// subscription from a non-zero in-range cursor delivers only strictly-greater ids.
func TestSubscribe_NonZeroStartingCursor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 10, "hist")

	// Learn the real ids by catching up from 0.
	s1 := newFakeSubscribeStream(ctx)
	e1 := make(chan error, 1)
	go func() { e1 <- svc.Subscribe(s1) }()
	s1.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 1,
	}))
	r1 := waitForResponses(t, s1, 10*time.Second, "all 10",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(groupMsgsFrom(rs)) >= 10 })
	var ids []uint64
	for _, m := range groupMsgsFrom(r1) {
		ids = append(ids, m.GetV1().GetId())
	}
	require.Len(t, ids, 10)
	s1.closeSend()
	require.NoError(t, <-e1)

	// Subscribe from the 5th id: only the 5 strictly-greater ids must arrive (no off-by-one).
	cursor := ids[4]
	s2 := newFakeSubscribeStream(ctx)
	e2 := make(chan error, 1)
	go func() { e2 <- svc.Subscribe(s2) }()
	s2.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), cursor),
		},
		MutateId: 2,
	}))
	r2 := waitForResponses(t, s2, 10*time.Second, "messages after the cursor",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })
	got := groupMsgsFrom(r2)
	require.Len(t, got, 5, "exactly the 5 messages with id > cursor")
	for _, m := range got {
		require.Greater(t, m.GetV1().GetId(), cursor, "no message at or below the cursor")
	}
	s2.closeSend()
	require.NoError(t, <-e2)
}

// fakeReadStore wraps a ReadMlsStore to inject catch-up faults: groupErr makes
// QueryGroupMessagesWaveScan fail, and groupGate / welcomeGate (if non-nil) block the
// corresponding wave scan until released, letting a test hold a topic mid-catch-up.
// gateGroup / gateInstallation (if set) narrow the gate to scans covering that topic, so
// other waves on the same stream proceed. scanParked (if non-nil) is closed when a gated
// scan first parks — the wave's ceiling is pinned by then, so a test can publish
// deterministically above it. corruptWelcomeIDs (if set) simulates welcome rows the store
// skips as unparseable (an unknown message_type from a future writer): they are dropped
// from the parsed slice but keep their raw-scan slot (lastRawID / rawCount), matching the
// real store's paging contract. All other methods delegate to the embedded store
// (including Queries(), so the wave's ceiling lookup is unaffected).
type fakeReadStore struct {
	mlsstore.ReadMlsStore
	// groupScans counts QueryGroupMessagesWaveScan calls, letting a test assert a
	// wave completed without touching the store (the current-cursor pruning), and
	// groupScanFilters records each call's group ids so a test can assert batch
	// rotation (bounded batches; a deep topic re-queried across turns).
	groupScans        atomic.Int32
	scanMu            sync.Mutex
	groupScanFilters  [][]string
	groupErr          error
	groupGate         chan struct{}
	gateGroup         []byte
	welcomeGate       chan struct{}
	gateInstallation  []byte
	scanParked        chan struct{}
	parkedOnce        sync.Once
	corruptWelcomeIDs map[uint64]bool
}

func filtersInclude(filters []mlsstore.GroupCatchup, groupID []byte) bool {
	for _, flt := range filters {
		if string(flt.GroupID) == string(groupID) {
			return true
		}
	}
	return false
}

func welcomeFiltersInclude(filters []mlsstore.WelcomeCatchup, installationKey []byte) bool {
	for _, flt := range filters {
		if string(flt.InstallationKey) == string(installationKey) {
			return true
		}
	}
	return false
}

func (f *fakeReadStore) parked() {
	if f.scanParked != nil {
		f.parkedOnce.Do(func() { close(f.scanParked) })
	}
}

func (f *fakeReadStore) QueryGroupMessagesWaveScan(
	ctx context.Context,
	filters []mlsstore.GroupCatchup,
	scanCursor uint64,
	ceiling uint64,
	limit int32,
) ([]*mlsv1.GroupMessage, error) {
	f.groupScans.Add(1)
	topics := make([]string, len(filters))
	for i, flt := range filters {
		topics[i] = string(flt.GroupID)
	}
	f.scanMu.Lock()
	f.groupScanFilters = append(f.groupScanFilters, topics)
	f.scanMu.Unlock()
	if f.groupGate != nil && (f.gateGroup == nil || filtersInclude(filters, f.gateGroup)) {
		f.parked()
		select {
		case <-f.groupGate:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.groupErr != nil {
		return nil, f.groupErr
	}
	return f.ReadMlsStore.QueryGroupMessagesWaveScan(ctx, filters, scanCursor, ceiling, limit)
}

func (f *fakeReadStore) QueryWelcomeMessagesWaveScan(
	ctx context.Context,
	filters []mlsstore.WelcomeCatchup,
	scanCursor uint64,
	ceiling uint64,
	limit int32,
) ([]*mlsv1.WelcomeMessage, map[string]mlsstore.WaveScanTopicProgress, error) {
	if f.welcomeGate != nil &&
		(f.gateInstallation == nil || welcomeFiltersInclude(filters, f.gateInstallation)) {
		f.parked()
		select {
		case <-f.welcomeGate:
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		}
	}
	msgs, progress, err := f.ReadMlsStore.QueryWelcomeMessagesWaveScan(
		ctx, filters, scanCursor, ceiling, limit,
	)
	if err == nil && len(f.corruptWelcomeIDs) > 0 {
		kept := msgs[:0]
		for _, m := range msgs {
			if !f.corruptWelcomeIDs[welcomeID(m)] {
				kept = append(kept, m)
			}
		}
		msgs = kept
	}
	return msgs, progress, err
}

// TestSubscribe_NonReadingClientIsReaped verifies a connected-but-non-reading client (a
// stalled stream.Send) is still reaped: the sender goroutine absorbs the block so the writer
// stays free to run the ping/pong reap.
func TestSubscribe_NonReadingClientIsReaped(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.pingInterval = 100 * time.Millisecond
	svc.pongDeadline = 200 * time.Millisecond

	stream := newFakeSubscribeStream(ctx)
	stream.blockSend = make(chan struct{}) // client never reads: every Send blocks
	defer close(stream.blockSend)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	select {
	case err := <-errCh:
		require.Equal(
			t,
			codes.DeadlineExceeded,
			status.Code(err),
			"a non-reading client must be reaped; the writer must not be parked on Send",
		)
	case <-time.After(5 * time.Second):
		t.Fatal("non-reading client was not reaped (writer parked on stream.Send?)")
	}
}

// TestSubscribe_HalfCloseFlushTimeoutFailsNotOK verifies a bounded catch-up (history_only +
// half-close) whose queued frames cannot be delivered — the client stopped reading, so the
// sender goroutine is wedged in stream.Send — fails with DeadlineExceeded rather than returning
// OK with a silently truncated catch-up. flush() must report whether the drain actually
// finished, not just that it waited.
func TestSubscribe_HalfCloseFlushTimeoutFailsNotOK(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.pongDeadline = 300 * time.Millisecond // the bound flush waits for the sender to drain

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	// Small history: every queued frame fits the send buffer, so the writer reaches the
	// half-close flush() rather than stalling earlier in send().
	populateGroupMessages(t, ctx, svc, groupA, 3, "hist")

	stream := newFakeSubscribeStream(ctx)
	stream.blockSend = make(
		chan struct{},
	) // client never reads: the sender wedges on the first Send
	defer close(stream.blockSend)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// Bounded catch-up: catch groupA up, then half-close so the server finishes the wave and
	// closes the stream itself. The frames queue but are never delivered (sender is blocked).
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
		},
		HistoryOnly: true,
		MutateId:    1,
	}))
	stream.closeSend()

	select {
	case err := <-errCh:
		require.Equal(
			t,
			codes.DeadlineExceeded,
			status.Code(err),
			"a half-close drain that cannot deliver its queued frames must fail, not return OK",
		)
	case <-time.After(5 * time.Second):
		t.Fatal("half-close flush did not return after the drain timed out")
	}
}

// TestSubscribe_CatchUpFetchErrorTearsDownUnavailable verifies a catch-up fetch error tears
// the stream down with Unavailable (so the client reconnects) rather than hanging.
func TestSubscribe_CatchUpFetchErrorTearsDownUnavailable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "hist")
	svc.readOnlyStore = &fakeReadStore{
		ReadMlsStore: svc.readOnlyStore,
		groupErr:     errors.New("db boom"),
	}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 1,
	}))

	select {
	case err := <-errCh:
		require.Equal(
			t,
			codes.Unavailable,
			status.Code(err),
			"a catch-up fetch error must tear down Unavailable",
		)
	case <-time.After(10 * time.Second):
		t.Fatal("stream did not tear down on a catch-up fetch error")
	}
}

// TestSubscribe_RemoveMidCatchUpFreesWaveSlot verifies removing a topic whose catch-up is
// still in flight completes its wave immediately (from the remove) without waiting on the
// blocked fetch, and the orphaned page is dropped.
func TestSubscribe_RemoveMidCatchUpFreesWaveSlot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "hist")
	gate := make(chan struct{})
	svc.readOnlyStore = &fakeReadStore{ReadMlsStore: svc.readOnlyStore, groupGate: gate}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// Subscribe (catch-up will block on the gate), then remove before it returns.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 7,
	}))
	stream.send(
		subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{Removes: [][]byte{groupTopic(groupA)}}),
	)

	resps := waitForResponses(t, stream, 5*time.Second, "CatchupComplete from the remove",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 2 })
	require.Equal(
		t,
		[]uint64{7, 0},
		catchupCompletesFrom(resps),
		"the removed topic's wave completes from the remove (echoing its mutate_id), then the removes-only Mutate is acked",
	)
	require.Empty(
		t,
		groupMsgsFrom(resps),
		"catch-up was still blocked, so no history was delivered",
	)

	close(gate) // release the orphaned fetch; its stale page must be dropped
	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_ScanSlotsQueueWaves pins both halves of the scan-concurrency
// cap with one slot and three waves:
//
//   - a second stale wave must not start its scan until the first wave's scan
//     finishes — deterministic because wave 1's done-batch enters the FIFO
//     catch-up channel before its fetcher releases the slot, and wave 2's
//     fetcher acquires the slot only after that release;
//   - a fully-current wave must NOT queue: it prunes slot-free and completes
//     while the sole slot is still parked inside the store, so its
//     CatchupComplete arrives first of the three.
//
// A broken cap flips the first property (wave 2 overtakes the parked wave 1);
// a current wave wrongly routed through the slot hangs the third
// CatchupComplete until the gate opens and flips the {3, ...} prefix.
func TestSubscribe_ScanSlotsQueueWaves(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.scanMaxConcurrent = 1

	groupA := []byte(test.RandomString(32))
	groupB := []byte(test.RandomString(32))
	groupC := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 2, "histA")
	validationSvc.ExpectedCalls = nil
	mockValidateGroupMessages(validationSvc, groupB)
	populateGroupMessages(t, ctx, svc, groupB, 2, "histB")

	latest, err := svc.readOnlyStore.Queries().GetLatestGroupMessageID(ctx)
	require.NoError(t, err)

	gate := make(chan struct{})
	parked := make(chan struct{})
	fake := &fakeReadStore{
		ReadMlsStore: svc.readOnlyStore,
		groupGate:    gate,
		scanParked:   parked,
	}
	svc.readOnlyStore = fake

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 1,
	}))
	select {
	case <-parked:
	case <-time.After(10 * time.Second):
		t.Fatal("wave 1's scan never reached the store")
	}

	// Wave 2 is stale too: its scan needs the slot wave 1 is sitting on.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupB), 0)},
		MutateId: 2,
	}))
	// Wave 3 is fully current: it must prune slot-free and complete NOW, while
	// the sole slot is parked in the store and wave 2 queues behind it.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupC), uint64(latest)),
		},
		MutateId: 3,
	}))
	waitForResponses(t, stream, 10*time.Second, "current wave to bypass the slot queue",
		func(rs []*mlsv1.SubscribeResponse) bool {
			completes := catchupCompletesFrom(rs)
			return len(completes) >= 1 && completes[0] == 3
		})
	// Give a hypothetically uncapped wave 2 ample time to (wrongly) reach the
	// store and park at the gate before it opens. With the cap this only
	// delays the gate.
	time.Sleep(500 * time.Millisecond)
	require.Equal(t, int32(1), fake.groupScans.Load(),
		"only the parked wave 1 may query while it holds the sole slot")
	close(gate)

	resps := waitForResponses(t, stream, 10*time.Second, "all three waves to complete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 3 })
	require.Equal(t, []uint64{3, 1, 2}, catchupCompletesFrom(resps),
		"current wave bypasses; the queued stale wave completes only after the slot holder")
	require.Len(t, groupMsgsFrom(resps), 4, "both stale waves' history must be delivered in full")
	require.Equal(t, int32(2), fake.groupScans.Load(),
		"one turn per stale wave; the current wave never queries")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_CatchUpByteBudgetBackpressures starves the catch-up byte
// budget (1 byte: every non-empty turn is over-budget) so each turn must be
// admitted alone through the CAS-from-zero arm, with the group and welcome
// scans contending for the same budget across many small rotation turns. The
// replay must still deliver everything exactly once, ascending per topic, and
// complete — a lost free or a lost wakeup deadlocks the wave and times the
// test out.
func TestSubscribe_CatchUpByteBudgetBackpressures(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.scanMaxPendingBytes = 1
	svc.scanTopicPageLimit = 2
	svc.scanBatchTopics = 2
	// Record the lane's peak occupancy at reservation time: with a 1-byte
	// budget every admission must go through the empty-lane arm, so occupancy
	// may never exceed a single turn — turns must not stack.
	var peakLane atomic.Int64
	svc.scanBytesObserved = func(n int64) {
		for {
			cur := peakLane.Load()
			if n <= cur || peakLane.CompareAndSwap(cur, n) {
				return
			}
		}
	}

	groupA := []byte(test.RandomString(32))
	groupB := []byte(test.RandomString(32))
	instA := []byte(test.RandomString(32))
	hpke := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")
	validationSvc.ExpectedCalls = nil
	mockValidateGroupMessages(validationSvc, groupB)
	populateGroupMessages(t, ctx, svc, groupB, 5, "histB")
	populateWelcomeMessages(t, ctx, svc, instA, hpke, 5, "welA")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
			addSub(groupTopic(groupB), 0),
			addSub(welcomeTopic(instA), 0),
		},
		MutateId: 1,
	}))

	resps := waitForResponses(t, stream, 10*time.Second, "full starved replay",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsFrom(rs)) >= 10 && len(welcomeMsgsFrom(rs)) >= 5 &&
				len(catchupCompletesFrom(rs)) >= 1
		})
	require.Equal(t, []uint64{1}, catchupCompletesFrom(resps))

	perTopic := map[string][]uint64{}
	for _, m := range groupMsgsFrom(resps) {
		key := string(m.GetV1().GetGroupId())
		perTopic[key] = append(perTopic[key], m.GetV1().GetId())
	}
	require.Len(t, perTopic, 2, "both group topics replayed")
	for topic, ids := range perTopic {
		require.Len(t, ids, 5, "topic %x exactly once", topic)
		for i := 1; i < len(ids); i++ {
			require.Greater(t, ids[i], ids[i-1], "topic %x ids must ascend", topic)
		}
	}
	welcomes := welcomeMsgsFrom(resps)
	require.Len(t, welcomes, 5, "welcome history exactly once")
	for i := 1; i < len(welcomes); i++ {
		require.Greater(t, welcomeID(welcomes[i]), welcomeID(welcomes[i-1]),
			"welcome ids must ascend")
	}

	// The largest possible turn here is a full group batch: 2 topics × 2 rows
	// of len("histX_n") = 7 payload bytes. Anything above that means two turns
	// were admitted concurrently past the starved budget.
	require.Positive(t, peakLane.Load(), "the observer must have seen reservations")
	require.LessOrEqual(t, peakLane.Load(), int64(28),
		"a starved budget must admit turns one at a time, never stacked")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_PendingBufferCapAborts verifies the maxPendingBytes guard: live messages
// buffered while a topic is stuck catching up cannot grow without bound.
func TestSubscribe_PendingBufferCapAborts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.maxPendingBytes = 50 // tiny: a couple of buffered live messages exceed it

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	// One history row keeps the ceiling above the cursor: a fully-current topic
	// is pruned without ever reaching the gated scan, and this test needs the
	// wave parked mid-catch-up.
	populateGroupMessages(t, ctx, svc, groupA, 1, "hist")
	gate := make(chan struct{})
	defer close(gate)
	svc.readOnlyStore = &fakeReadStore{ReadMlsStore: svc.readOnlyStore, groupGate: gate}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 1,
	}))

	// Catch-up is gated, so the topic stays in catch-up; live messages pile into pending.
	for i := 0; i < 20; i++ {
		publishGroup(t, ctx, svc, validationSvc, groupA, fmt.Sprintf("live-message-%d", i))
	}

	select {
	case err := <-errCh:
		require.Equal(t, codes.ResourceExhausted, status.Code(err),
			"exceeding maxPendingBytes must abort the stream")
	case <-time.After(10 * time.Second):
		t.Fatal("pending-buffer cap was not enforced")
	}
}

// TestSubscribe_FramesSplitByMaxFrameBytes verifies the per-frame byte cap splits a batch of
// messages across multiple frames.
func TestSubscribe_FramesSplitByMaxFrameBytes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.maxFrameBytes = 1 // every non-empty message lands in its own frame

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	const n = 5
	populateGroupMessages(t, ctx, svc, groupA, n, "hist")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 1,
	}))
	resps := waitForResponses(t, stream, 10*time.Second, "all history",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(groupMsgsFrom(rs)) >= n })

	frames := 0
	for _, r := range resps {
		if msgs := r.GetV1().GetMessages(); msgs != nil && len(msgs.GetGroupMessages()) > 0 {
			frames++
		}
	}
	require.Equal(t, n, frames, "at maxFrameBytes=1 each message must be sent in its own frame")
	require.Len(t, groupMsgsFrom(resps), n)

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_LiveBurstCoalescesWhenQueued pins the live-lane coalescing drain: a burst of
// live messages queued behind the one the writer is processing is packed into fewer frames
// (bounded by liveCoalesceMax / maxFrameBytes) instead of one tiny frame each, while every
// message is still delivered exactly once, in ascending id order, tagged live (mutate_id 0).
//
// Determinism: publishing only after CatchupComplete keeps every message on the live path (not
// the gated catch-up lane). Freezing the sender first caps in-flight frames at sendQueueDepth,
// so a 64-message burst cannot be flushed as 64 singles — the writer must coalesce the
// remainder — making the coalescing assertion deterministic rather than timing-dependent.
func TestSubscribe_LiveBurstCoalescesWhenQueued(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupId)

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// No history: subscribe, then wait for catch-up to complete so the topic is fully live
	// before publishing — otherwise early messages would ride the gated catch-up lane and be
	// tagged with the wave rather than delivered live.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupId), 0)},
		MutateId: 1,
	}))
	waitForResponses(t, stream, 5*time.Second, "CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })

	// Quiesce the sender so the burst queues on the subscription channel and the writer must
	// coalesce it (in-flight frames are bounded by sendQueueDepth), then release to flush.
	release := stream.freezeSends()
	const n = 64
	populateGroupMessages(t, ctx, svc, groupId, n, "live")
	release()

	resps := waitForResponses(
		t,
		stream,
		10*time.Second,
		fmt.Sprintf("%d live group messages", n),
		func(rs []*mlsv1.SubscribeResponse) bool { return len(groupMsgsTagged(rs, 0)) >= n },
	)

	// Correctness: every message delivered exactly once, in order, tagged live (0).
	live := groupMsgsTagged(resps, 0)
	require.Len(t, live, n, "every live message delivered exactly once, tagged mutate_id 0")
	validateMessageOrdering(t, live)
	validateNoDuplicates(t, live)

	// Coalescing: with the sender frozen during the burst, at least one live frame must carry
	// several messages — proving the drain packed a queued burst rather than one-per-frame.
	maxPerFrame := 0
	for _, r := range resps {
		if m := r.GetV1().GetMessages(); m != nil && m.GetMutateId() == 0 {
			if k := len(m.GetGroupMessages()); k > maxPerFrame {
				maxPerFrame = k
			}
		}
	}
	require.Greater(
		t,
		maxPerFrame,
		1,
		"a queued live burst must coalesce into at least one multi-message frame",
	)

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// --- XIP-83 delivery tagging & ordering (server requirements 3 and 4) ---

// groupMsgsTagged returns the group messages carried by Messages frames whose mutate_id
// equals tag, in receive order.
func groupMsgsTagged(resps []*mlsv1.SubscribeResponse, tag uint64) []*mlsv1.GroupMessage {
	var out []*mlsv1.GroupMessage
	for _, r := range resps {
		if msgs := r.GetV1().GetMessages(); msgs != nil && msgs.GetMutateId() == tag {
			out = append(out, msgs.GetGroupMessages()...)
		}
	}
	return out
}

// welcomeMsgsTagged is the welcome analogue of groupMsgsTagged.
func welcomeMsgsTagged(resps []*mlsv1.SubscribeResponse, tag uint64) []*mlsv1.WelcomeMessage {
	var out []*mlsv1.WelcomeMessage
	for _, r := range resps {
		if msgs := r.GetV1().GetMessages(); msgs != nil && msgs.GetMutateId() == tag {
			out = append(out, msgs.GetWelcomeMessages()...)
		}
	}
	return out
}

// requireAscendingGroupIds asserts the messages' ids strictly ascend in the given order —
// the total-order shape both delivery lanes guarantee per kind.
func requireAscendingGroupIds(t *testing.T, msgs []*mlsv1.GroupMessage, desc string) {
	t.Helper()
	for i := 1; i < len(msgs); i++ {
		require.Greater(
			t,
			msgs[i].GetV1().GetId(),
			msgs[i-1].GetV1().GetId(),
			"%s: ids must strictly ascend (index %d)",
			desc,
			i,
		)
	}
}

// requireAscendingWelcomeIds is the welcome analogue of requireAscendingGroupIds.
func requireAscendingWelcomeIds(t *testing.T, msgs []*mlsv1.WelcomeMessage, desc string) {
	t.Helper()
	for i := 1; i < len(msgs); i++ {
		require.Greater(
			t,
			welcomeID(msgs[i]),
			welcomeID(msgs[i-1]),
			"%s: welcome ids must strictly ascend (index %d)",
			desc,
			i,
		)
	}
}

// TestSubscribe_ReplayTaggedWithWaveLiveTaggedZero pins delivery tagging under overlapping
// waves: every replay frame carries exactly the mutate_id of the wave that produced it —
// even when a later wave completes first — and live frames carry 0.
func TestSubscribe_ReplayTaggedWithWaveLiveTaggedZero(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	instB := []byte(test.RandomString(32))
	hpke := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")
	populateWelcomeMessages(t, ctx, svc, instB, hpke, 3, "welB")

	// Gate group scans so wave 7 stays in flight while wave 8 (welcomes) overtakes it.
	gate := make(chan struct{})
	svc.readOnlyStore = &fakeReadStore{ReadMlsStore: svc.readOnlyStore, groupGate: gate}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 7,
	}))
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(welcomeTopic(instB), 0)},
		MutateId: 8,
	}))

	// Wave 8 completes while wave 7 is still gated mid-fetch.
	resps := waitForResponses(t, stream, 10*time.Second, "wave 8 catch-up",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })
	require.Equal(t, []uint64{8}, catchupCompletesFrom(resps), "wave 8 overtakes the gated wave 7")
	require.Len(t, welcomeMsgsTagged(resps, 8), 3, "wave 8's replay is tagged 8")
	require.Empty(t, groupMsgsTagged(resps, 7), "wave 7 is still gated: no frames yet")

	// Release wave 7: its replay arrives tagged 7, never mixed into another tag.
	close(gate)
	resps = waitForResponses(t, stream, 10*time.Second, "wave 7 catch-up",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 2 })
	require.Equal(t, []uint64{8, 7}, catchupCompletesFrom(resps))
	require.Len(t, groupMsgsTagged(resps, 7), 5, "wave 7's replay is tagged 7")
	require.Empty(t, groupMsgsTagged(resps, 8), "no group frame may carry wave 8's tag")
	require.Empty(t, groupMsgsTagged(resps, 0), "no replay frame may carry the live tag")

	// Live tail on both topics is tagged 0.
	publishGroup(t, ctx, svc, validationSvc, groupA, "a-live")
	populateWelcomeMessages(t, ctx, svc, instB, hpke, 1, "welB-live")
	resps = waitForResponses(t, stream, 10*time.Second, "live tail",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsTagged(rs, 0)) >= 1 && len(welcomeMsgsTagged(rs, 0)) >= 1
		})
	require.Len(t, groupMsgsWithData(groupMsgsTagged(resps, 0), "a-live"), 1)
	require.Len(t, welcomeMsgsTagged(resps, 0), 1)

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_InflightMutateIdCollisionRejected verifies that a Mutate reusing the
// mutate_id of a wave still in flight is rejected with InvalidArgument: the frame tag and
// the CatchupComplete echo are the only keys correlating frames to mutations, so an
// in-flight collision would make the two waves' replay and completions indistinguishable
// (XIP-83 server requirement 3). The gate holds wave 7 mid-fetch; both an adds-bearing
// Mutate and a removes-only Mutate (whose immediate CatchupComplete would be ambiguous
// with the in-flight wave's) must be rejected.
func TestSubscribe_InflightMutateIdCollisionRejected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	instB := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "hist")
	gate := make(chan struct{})
	defer close(gate)
	svc.readOnlyStore = &fakeReadStore{ReadMlsStore: svc.readOnlyStore, groupGate: gate}

	// Each rejection fails its stream, so each colliding Mutate gets a fresh stream with
	// its own gated wave 7 in flight.
	rejected := func(desc string, colliding *mlsv1.SubscribeRequest_V1_Mutate) {
		t.Helper()
		stream := newFakeSubscribeStream(ctx)
		errCh := make(chan error, 1)
		go func() { errCh <- svc.Subscribe(stream) }()

		// Wave 7's catch-up blocks on the gate, staying in flight.
		stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
			Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
				addSub(groupTopic(groupA), 0),
			},
			MutateId: 7,
		}))
		stream.send(subReqMutate(colliding))

		select {
		case err := <-errCh:
			require.Equal(t, codes.InvalidArgument, status.Code(err), desc)
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: colliding Mutate was not rejected", desc)
		}
	}

	rejected("adds reusing an in-flight mutate_id must be rejected",
		&mlsv1.SubscribeRequest_V1_Mutate{
			Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
				addSub(welcomeTopic(instB), 0),
			},
			MutateId: 7,
		})
	rejected("a removes-only Mutate reusing an in-flight mutate_id must be rejected",
		&mlsv1.SubscribeRequest_V1_Mutate{
			Removes:  [][]byte{welcomeTopic(instB)},
			MutateId: 7,
		})
}

// TestSubscribe_MutateIdReuseAfterCompletionAccepted pins the boundary of the in-flight
// collision rule: once a wave's CatchupComplete has been emitted its mutate_id is free
// again, so a later adds-Mutate reusing it replays and completes normally.
func TestSubscribe_MutateIdReuseAfterCompletionAccepted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	instB := []byte(test.RandomString(32))
	hpke := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")
	populateWelcomeMessages(t, ctx, svc, instB, hpke, 3, "welB")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	// Wave 7 runs to completion.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 7,
	}))
	waitForResponses(t, stream, 10*time.Second, "wave 7 catch-up",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })

	// The id is no longer in flight: a fresh wave reusing 7 replays and completes again.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(welcomeTopic(instB), 0)},
		MutateId: 7,
	}))
	resps := waitForResponses(t, stream, 10*time.Second, "reused wave 7 catch-up",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 2 })
	require.Equal(t, []uint64{7, 7}, catchupCompletesFrom(resps),
		"each wave completes with its own CatchupComplete echoing the reused id")
	require.Len(t, groupMsgsTagged(resps, 7), 5, "the first wave's replay is tagged 7")
	require.Len(t, welcomeMsgsTagged(resps, 7), 3, "the reused wave's replay is tagged 7")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_WaveReplayMergedAcrossTopics pins wave order: one wave covering two groups
// whose histories interleave by id must replay the union in ascending id order — one merged
// cursor-ordered pass — not one topic's burst then the other's.
func TestSubscribe_WaveReplayMergedAcrossTopics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	groupB := []byte(test.RandomString(32))
	// Alternate publishes so the two groups' ids interleave globally.
	const perGroup = 6
	for i := 0; i < perGroup; i++ {
		publishGroup(t, ctx, svc, validationSvc, groupA, fmt.Sprintf("a-%d", i))
		publishGroup(t, ctx, svc, validationSvc, groupB, fmt.Sprintf("b-%d", i))
	}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
			addSub(groupTopic(groupB), 0),
		},
		MutateId: 3,
	}))
	resps := waitForResponses(t, stream, 10*time.Second, "merged replay + CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsFrom(rs)) >= 2*perGroup && len(catchupCompletesFrom(rs)) >= 1
		})

	replay := groupMsgsTagged(resps, 3)
	require.Len(t, replay, 2*perGroup, "the wave delivers both groups' history, tagged 3")
	requireAscendingGroupIds(t, replay, "wave replay across interleaved topics")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_LiveTotalOrderPerKind pins live order: live (mutate_id 0) group messages
// across all live topics on the stream arrive in ascending group id order, and live
// welcomes across all live installations in ascending welcome id order — each kind totally
// ordered on its own id sequence, independent of the other's.
func TestSubscribe_LiveTotalOrderPerKind(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	groupB := []byte(test.RandomString(32))
	instA := []byte(test.RandomString(32))
	instB := []byte(test.RandomString(32))
	hpke := []byte(test.RandomString(32))

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
			addSub(groupTopic(groupB), 0),
			addSub(welcomeTopic(instA), 0),
			addSub(welcomeTopic(instB), 0),
		},
		MutateId: 1,
	}))
	waitForResponses(t, stream, 10*time.Second, "CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })

	// Interleave the kinds so their publishes overlap in time; the two id sequences are
	// independent, and each must ascend on its own.
	const perGroup = 5
	for i := 0; i < perGroup; i++ {
		publishGroup(t, ctx, svc, validationSvc, groupA, fmt.Sprintf("a-live-%d", i))
		populateWelcomeMessages(t, ctx, svc, instA, hpke, 1, fmt.Sprintf("wa-live-%d", i))
		publishGroup(t, ctx, svc, validationSvc, groupB, fmt.Sprintf("b-live-%d", i))
		populateWelcomeMessages(t, ctx, svc, instB, hpke, 1, fmt.Sprintf("wb-live-%d", i))
	}
	resps := waitForResponses(t, stream, 10*time.Second, "all live messages",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsTagged(rs, 0)) >= 2*perGroup &&
				len(welcomeMsgsTagged(rs, 0)) >= 2*perGroup
		})

	live := groupMsgsTagged(resps, 0)
	require.Len(t, live, 2*perGroup)
	requireAscendingGroupIds(t, live, "live lane across topics")

	liveWelcomes := welcomeMsgsTagged(resps, 0)
	require.Len(t, liveWelcomes, 2*perGroup)
	requireAscendingWelcomeIds(t, liveWelcomes, "live welcome lane across installations")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_SeamHoldsLiveUntilCatchupComplete pins the seam: while a wave replays a
// topic, that topic speaks live (mutate_id 0) only after the wave's CatchupComplete —
// messages published mid-wave are folded into the wave (tagged), delivered exactly once —
// while live frames for topics of other subscriptions keep flowing throughout.
func TestSubscribe_SeamHoldsLiveUntilCatchupComplete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	sentinelId := []byte(test.RandomString(32))
	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")

	// Gate ONLY groupA's wave, so the sentinel's wave and live tail flow freely.
	gate := make(chan struct{})
	svc.readOnlyStore = &fakeReadStore{
		ReadMlsStore: svc.readOnlyStore,
		groupGate:    gate,
		gateGroup:    groupA,
	}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(sentinelId), 0),
		},
		MutateId: 1,
	}))
	waitForResponses(t, stream, 10*time.Second, "sentinel live",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })

	// Wave 9 starts (gated mid-fetch); groupA messages published now are gated behind it,
	// while the sentinel keeps receiving live frames.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 9,
	}))
	publishGroup(t, ctx, svc, validationSvc, groupA, "a-mid-wave")
	publishGroup(t, ctx, svc, validationSvc, sentinelId, "s-live")
	resps := waitForResponses(t, stream, 10*time.Second, "sentinel live during the wave",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsTagged(rs, 0), "s-live")) >= 1
		})
	require.Empty(t, groupMsgsTagged(resps, 9), "wave 9 is still gated")
	require.Empty(
		t,
		groupMsgsWithData(groupMsgsTagged(resps, 0), "a-mid-wave"),
		"no live frame for the wave's topic before its CatchupComplete",
	)

	// Release the wave: history and the folded mid-wave message arrive tagged 9, in
	// ascending id order, then CatchupComplete; only frames after it speak live for groupA.
	close(gate)
	resps = waitForResponses(t, stream, 10*time.Second, "wave 9 complete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 2 })
	replay := groupMsgsTagged(resps, 9)
	require.Len(t, replay, 6, "5 history + 1 folded mid-wave message, all tagged 9")
	require.Len(t, groupMsgsWithData(replay, "a-mid-wave"), 1, "folded exactly once")
	requireAscendingGroupIds(t, replay, "wave 9 replay including the folded tail")
	require.Empty(t, groupMsgsWithData(groupMsgsTagged(resps, 0), "a-mid-wave"),
		"the folded message must not also be delivered live")

	// Live tail for groupA now flows tagged 0, strictly after the wave's CatchupComplete.
	publishGroup(t, ctx, svc, validationSvc, groupA, "a-live")
	resps = waitForResponses(t, stream, 10*time.Second, "groupA live tail",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsTagged(rs, 0), "a-live")) >= 1
		})
	ccIdx := -1
	for i, r := range resps {
		if cc := r.GetV1().GetCatchupComplete(); cc != nil && cc.GetMutateId() == 9 {
			ccIdx = i
		}
	}
	require.GreaterOrEqual(t, ccIdx, 0)
	liveIdx := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		msgs := r.GetV1().GetMessages()
		return msgs.GetMutateId() == 0 &&
			len(groupMsgsWithData(msgs.GetGroupMessages(), "a-live")) > 0
	})
	require.Greater(t, liveIdx, ccIdx, "groupA's live tail begins after CatchupComplete(9)")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_AddsRequireNonzeroMutateId pins the request-side rule the tag depends on:
// a Mutate with adds and mutate_id 0 (the live tag) fails the stream with InvalidArgument.
func TestSubscribe_AddsRequireNonzeroMutateId(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic([]byte(test.RandomString(32))), 0),
		},
	}))

	select {
	case err := <-errCh:
		require.Equal(t, codes.InvalidArgument, status.Code(err),
			"adds with mutate_id 0 must be rejected")
	case <-time.After(5 * time.Second):
		t.Fatal("a Mutate with adds and mutate_id 0 was not rejected")
	}
}

// TestSubscribe_MutateAddsCapRejected verifies the per-Mutate adds cap (maxMutateAdds): a
// Mutate carrying more adds than the cap fails the stream with ResourceExhausted, and the
// rejection is atomic — it lands before the Mutate's removes or adds take any effect, so
// the offending Mutate produces no frames at all. A fresh stream with exactly the cap
// catches up to CatchupComplete.
func TestSubscribe_MutateAddsCapRejected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.maxMutateAdds = 3 // tiny: a handful of adds exceeds it

	// Stream 1: subscribe a live topic, then send a Mutate that pairs removes:[liveTopic]
	// with 4 adds (one over the cap).
	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	liveTopic := groupTopic([]byte(test.RandomString(32)))
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(liveTopic, 0)},
		MutateId: 1,
	}))
	waitForResponses(
		t,
		stream,
		5*time.Second,
		"wave 1 CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 },
	)

	overCap := &mlsv1.SubscribeRequest_V1_Mutate{Removes: [][]byte{liveTopic}, MutateId: 2}
	for i := 0; i < 4; i++ {
		overCap.Adds = append(overCap.Adds, addSub(groupTopic([]byte(test.RandomString(32))), 0))
	}
	stream.send(subReqMutate(overCap))

	select {
	case err := <-errCh:
		require.Equal(t, codes.ResourceExhausted, status.Code(err),
			"a Mutate with more adds than maxMutateAdds must be rejected")
	case <-time.After(5 * time.Second):
		t.Fatal("over-cap Mutate was not rejected")
	}
	// Atomicity: the rejection precedes the remove and the adds, so the offending Mutate
	// delivered nothing — the stream's frames are exactly wave 1's (its TopicsLive and its
	// CatchupComplete), with no CatchupComplete echoing mutate_id 2 and no later TopicsLive.
	resps := stream.responses()
	require.Equal(t, []uint64{1}, catchupCompletesFrom(resps),
		"the rejected Mutate must not produce a CatchupComplete")
	topicsLiveFrames := 0
	for _, r := range resps {
		if r.GetV1().GetTopicsLive() != nil {
			topicsLiveFrames++
		}
	}
	require.Equal(t, 1, topicsLiveFrames,
		"the rejected Mutate must not produce a TopicsLive frame")

	// Stream 2: exactly the cap is allowed and catches up.
	stream2 := newFakeSubscribeStream(ctx)
	errCh2 := make(chan error, 1)
	go func() { errCh2 <- svc.Subscribe(stream2) }()

	atCap := &mlsv1.SubscribeRequest_V1_Mutate{MutateId: 3}
	for i := 0; i < 3; i++ {
		atCap.Adds = append(atCap.Adds, addSub(groupTopic([]byte(test.RandomString(32))), 0))
	}
	stream2.send(subReqMutate(atCap))
	resps2 := waitForResponses(
		t,
		stream2,
		10*time.Second,
		"at-cap wave CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 },
	)
	require.Equal(t, []uint64{3}, catchupCompletesFrom(resps2),
		"a Mutate with exactly maxMutateAdds adds must catch up normally")

	stream2.closeSend()
	require.NoError(t, <-errCh2)
}

// TestSubscribe_ResetMidWaveHandsTopicToNewWave pins the seam when a topic moves between
// overlapping waves: a remove+re-add mid-wave hands the topic (gate ownership, pending
// buffer, cursor floor) to the new wave — the old wave replays only its surviving topics,
// and the reset topic's full history plus its discarded mid-wave live message arrive
// exactly once, under the new wave's tag.
func TestSubscribe_ResetMidWaveHandsTopicToNewWave(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	t1 := []byte(test.RandomString(32))
	t2 := []byte(test.RandomString(32))
	sentinelId := []byte(test.RandomString(32))
	for i := 0; i < 5; i++ {
		publishGroup(t, ctx, svc, validationSvc, t1, fmt.Sprintf("h1-%d", i))
		publishGroup(t, ctx, svc, validationSvc, t2, fmt.Sprintf("h2-%d", i))
	}

	// Gate ONLY scans covering t1 (both waves below include it), so the sentinel's wave
	// and live tail flow freely.
	gate := make(chan struct{})
	svc.readOnlyStore = &fakeReadStore{
		ReadMlsStore: svc.readOnlyStore,
		groupGate:    gate,
		gateGroup:    t1,
	}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(sentinelId), 0),
		},
		MutateId: 11,
	}))
	waitForResponses(t, stream, 10*time.Second, "sentinel live",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })

	// Wave 1 (gated mid-fetch) owns both topics; a live t1 message lands in its pending.
	// The dbWorker dispatches strictly in id order, so the sentinel's arrival proves "mid"
	// has been routed into the gated buffer.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(t1), 0),
			addSub(groupTopic(t2), 0),
		},
		MutateId: 1,
	}))
	publishGroup(t, ctx, svc, validationSvc, t1, "mid")
	publishGroup(t, ctx, svc, validationSvc, sentinelId, "s-live")
	waitForResponses(t, stream, 10*time.Second, "sentinel live during wave 1",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsTagged(rs, 0), "s-live")) >= 1
		})

	// Reset t1 mid-wave: wave 1 drops it (pending discarded), wave 2 owns it gated.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Removes:  [][]byte{groupTopic(t1)},
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(t1), 0)},
		MutateId: 2,
	}))

	close(gate)
	resps := waitForResponses(t, stream, 10*time.Second, "both waves complete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 3 })

	// Wave 1 replays only its surviving topic.
	wave1 := groupMsgsTagged(resps, 1)
	require.Len(t, wave1, 5, "wave 1 delivers only t2's history")
	for _, m := range wave1 {
		require.Equal(t, string(t2), string(m.GetV1().GetGroupId()),
			"no t1 frame may carry wave 1's tag")
	}
	requireAscendingGroupIds(t, wave1, "wave 1 replay")

	// Wave 2 delivers ALL of t1 — the history plus the reset-discarded "mid" message, which
	// its own scan (whose ceiling postdates the publish) picks back up — exactly once.
	wave2 := groupMsgsTagged(resps, 2)
	require.Len(t, wave2, 6, "wave 2 delivers t1's full history plus the mid-wave message")
	for _, m := range wave2 {
		require.Equal(t, string(t1), string(m.GetV1().GetGroupId()))
	}
	require.Len(t, groupMsgsWithData(wave2, "mid"), 1, "the mid-wave message arrives exactly once")
	requireAscendingGroupIds(t, wave2, "wave 2 replay")

	// t1 never speaks live before wave 2 completes (it was gated throughout).
	for _, m := range groupMsgsTagged(resps, 0) {
		require.NotEqual(t, string(t1), string(m.GetV1().GetGroupId()),
			"no t1 frame may be tagged live")
	}
	// No message is delivered under two tags.
	seen := make(map[uint64]bool)
	for _, tag := range []uint64{0, 1, 2, 11} {
		for _, m := range groupMsgsTagged(resps, tag) {
			require.False(t, seen[m.GetV1().GetId()], "a message must appear under exactly one tag")
			seen[m.GetV1().GetId()] = true
		}
	}

	// Each wave's TopicsLive announces only what it still owned at completion (the marker
	// is sent in the same batch as its CatchupComplete, so it is the frame just before).
	cc1 := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		cc := r.GetV1().GetCatchupComplete()
		return cc != nil && cc.GetMutateId() == 1
	})
	cc2 := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		cc := r.GetV1().GetCatchupComplete()
		return cc != nil && cc.GetMutateId() == 2
	})
	require.Greater(t, cc1, 0)
	require.Greater(t, cc2, 0)
	require.Equal(t, [][]byte{groupTopic(t2)}, resps[cc1-1].GetV1().GetTopicsLive().GetTopics(),
		"wave 1 announces only t2 live")
	require.Equal(t, [][]byte{groupTopic(t1)}, resps[cc2-1].GetV1().GetTopicsLive().GetTopics(),
		"wave 2 announces only t1 live")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_RemoveOneTopicMidMultiTopicWave verifies removing one of a wave's topics
// mid-flight: the wave completes with only the survivor — the removed topic's history and
// buffered live messages are dropped — and the removes-only Mutate itself is acked
// immediately, echoing its own mutate_id.
func TestSubscribe_RemoveOneTopicMidMultiTopicWave(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	t1 := []byte(test.RandomString(32))
	t2 := []byte(test.RandomString(32))
	sentinelId := []byte(test.RandomString(32))
	for i := 0; i < 5; i++ {
		publishGroup(t, ctx, svc, validationSvc, t1, fmt.Sprintf("g1-%d", i))
		publishGroup(t, ctx, svc, validationSvc, t2, fmt.Sprintf("g2-%d", i))
	}

	gate := make(chan struct{})
	svc.readOnlyStore = &fakeReadStore{
		ReadMlsStore: svc.readOnlyStore,
		groupGate:    gate,
		gateGroup:    t1,
	}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(sentinelId), 0),
		},
		MutateId: 1,
	}))
	waitForResponses(t, stream, 10*time.Second, "sentinel live",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })

	// Wave 3 (gated mid-fetch) owns both topics; put a live t1 message in its pending
	// (sentinel-sequenced), then remove t1 while the wave is still in flight.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(t1), 0),
			addSub(groupTopic(t2), 0),
		},
		MutateId: 3,
	}))
	publishGroup(t, ctx, svc, validationSvc, t1, "g1-mid")
	publishGroup(t, ctx, svc, validationSvc, sentinelId, "s-live")
	waitForResponses(t, stream, 10*time.Second, "sentinel live during wave 3",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsTagged(rs, 0), "s-live")) >= 1
		})
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Removes:  [][]byte{groupTopic(t1)},
		MutateId: 4,
	}))

	close(gate)
	resps := waitForResponses(t, stream, 10*time.Second, "wave 3 complete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 3 })
	require.Equal(t, []uint64{1, 4, 3}, catchupCompletesFrom(resps),
		"the remove is acked immediately; the wave completes exactly once")

	wave := groupMsgsTagged(resps, 3)
	require.Len(t, wave, 5, "the wave delivers only the surviving topic's history")
	for _, m := range wave {
		require.Equal(t, string(t2), string(m.GetV1().GetGroupId()))
	}
	requireAscendingGroupIds(t, wave, "wave 3 replay")

	// The removed topic must not surface at all — not its history, not its pending.
	for _, r := range resps {
		for _, m := range r.GetV1().GetMessages().GetGroupMessages() {
			require.NotEqual(t, string(t1), string(m.GetV1().GetGroupId()),
				"no t1 message may be delivered under any tag")
		}
		require.False(t, topicsLiveHas(r, groupTopic(t1)), "no marker may announce t1 live")
	}
	cc3 := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		cc := r.GetV1().GetCatchupComplete()
		return cc != nil && cc.GetMutateId() == 3
	})
	require.Greater(t, cc3, 0)
	require.Equal(t, [][]byte{groupTopic(t2)}, resps[cc3-1].GetV1().GetTopicsLive().GetTopics(),
		"the wave announces only the surviving topic")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_WelcomeFoldAndMergedOrder pins the welcome side of wave completion: a
// two-installation welcome wave replays merged in ascending welcome id order, live welcomes
// gated mid-wave fold into the wave (tagged, exactly once, in the merged order), and a
// post-completion welcome is live tail.
func TestSubscribe_WelcomeFoldAndMergedOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()

	instA := []byte(test.RandomString(32))
	instB := []byte(test.RandomString(32))
	instC := []byte(test.RandomString(32)) // live sentinel: sequences the gated publishes
	hpke := []byte(test.RandomString(32))
	// Alternate publishes so the two installations' ids interleave globally.
	for i := 0; i < 3; i++ {
		populateWelcomeMessages(t, ctx, svc, instA, hpke, 1, fmt.Sprintf("wA-%d", i))
		populateWelcomeMessages(t, ctx, svc, instB, hpke, 1, fmt.Sprintf("wB-%d", i))
	}

	// Gate ONLY scans covering instA (the wave below), so the sentinel's wave flows freely.
	// scanParked pins the sequencing: once the wave parks, its ceiling is snapshotted, so
	// welcomes published after it MUST reach the client through the pending fold.
	gate := make(chan struct{})
	parked := make(chan struct{})
	svc.readOnlyStore = &fakeReadStore{
		ReadMlsStore:     svc.readOnlyStore,
		welcomeGate:      gate,
		gateInstallation: instA,
		scanParked:       parked,
	}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(welcomeTopic(instC), 0),
		},
		MutateId: 1,
	}))
	waitForResponses(t, stream, 10*time.Second, "sentinel live",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(welcomeTopic(instA), 0),
			addSub(welcomeTopic(instB), 0),
		},
		MutateId: 5,
	}))
	select {
	case <-parked:
	case <-time.After(5 * time.Second):
		t.Fatal("the welcome wave never reached its gated scan")
	}

	// One live welcome per installation, gated into the wave's pending — instB FIRST, so
	// its welcome gets the lower id and the fold's wave-topic-order concatenation
	// [instA's, instB's] is descending: only the fold's sort restores the merged order.
	// The welcome dispatch is strictly id-ordered, so the sentinel's arrival proves both
	// were routed.
	populateWelcomeMessages(t, ctx, svc, instB, hpke, 1, "wB-live")
	populateWelcomeMessages(t, ctx, svc, instA, hpke, 1, "wA-live")
	populateWelcomeMessages(t, ctx, svc, instC, hpke, 1, "wC-live")
	waitForResponses(t, stream, 10*time.Second, "sentinel welcome during the wave",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(welcomesForKey(welcomeMsgsTagged(rs, 0), instC)) >= 1
		})

	close(gate)
	resps := waitForResponses(t, stream, 10*time.Second, "wave 5 complete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 2 })

	wave := welcomeMsgsTagged(resps, 5)
	require.Len(t, wave, 8, "6 history + 2 folded live welcomes, all tagged 5")
	require.Len(t, welcomesForKey(wave, instA), 4)
	require.Len(t, welcomesForKey(wave, instB), 4)
	requireAscendingWelcomeIds(t, wave, "welcome wave replay including the folded tail")

	// Before the wave's CatchupComplete, no welcome for its installations may be live.
	cc5 := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		cc := r.GetV1().GetCatchupComplete()
		return cc != nil && cc.GetMutateId() == 5
	})
	require.Greater(t, cc5, 0)
	for _, r := range resps[:cc5] {
		if msgs := r.GetV1().GetMessages(); msgs != nil && msgs.GetMutateId() == 0 {
			for _, m := range msgs.GetWelcomeMessages() {
				key := welcomeInstallationKey(m)
				require.NotEqual(t, string(instA), string(key),
					"no live frame for a wave installation before its CatchupComplete")
				require.NotEqual(t, string(instB), string(key),
					"no live frame for a wave installation before its CatchupComplete")
			}
		}
	}

	// A post-completion welcome flows live, tagged 0.
	populateWelcomeMessages(t, ctx, svc, instA, hpke, 1, "wA-after")
	waitForResponses(t, stream, 10*time.Second, "post-completion live welcome",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(welcomesForKey(welcomeMsgsTagged(rs, 0), instA)) >= 1
		})

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_FoldSortsAcrossTopics pins the fold's merge order on the group side: live
// messages gated for DIFFERENT topics of one wave land in per-topic pending buffers, and
// completeWave concatenates those buffers in wave-topic order — so when the publish order
// inverts the topic order, only the fold's sort restores the wave's ascending-id replay.
func TestSubscribe_FoldSortsAcrossTopics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	t1 := []byte(test.RandomString(32))
	t2 := []byte(test.RandomString(32))
	sentinelId := []byte(test.RandomString(32))
	for i := 0; i < 3; i++ {
		publishGroup(t, ctx, svc, validationSvc, t1, fmt.Sprintf("h1-%d", i))
		publishGroup(t, ctx, svc, validationSvc, t2, fmt.Sprintf("h2-%d", i))
	}

	// Gate ONLY scans covering t1 (the two-topic wave below), so the sentinel's wave and
	// live tail flow freely. scanParked pins the sequencing: once the wave parks, its
	// ceiling is snapshotted, so messages published after it MUST reach the client through
	// the pending fold.
	gate := make(chan struct{})
	parked := make(chan struct{})
	svc.readOnlyStore = &fakeReadStore{
		ReadMlsStore: svc.readOnlyStore,
		groupGate:    gate,
		gateGroup:    t1,
		scanParked:   parked,
	}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(sentinelId), 0),
		},
		MutateId: 1,
	}))
	waitForResponses(t, stream, 10*time.Second, "sentinel live",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })

	// Wave 6 owns t1 THEN t2, so completeWave folds pending as [t1's, t2's].
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(t1), 0),
			addSub(groupTopic(t2), 0),
		},
		MutateId: 6,
	}))
	select {
	case <-parked:
	case <-time.After(5 * time.Second):
		t.Fatal("the two-topic wave never reached its gated scan")
	}

	// Publish to t2 FIRST, then t1: t2's gated message gets the LOWER id, so the fold's
	// wave-topic-order concatenation [t1's, t2's] is descending — only the sort restores
	// the replay's order. The dbWorker dispatches strictly in id order, so the sentinel's
	// arrival proves both were routed into the gated pending buffers.
	publishGroup(t, ctx, svc, validationSvc, t2, "t2-mid")
	publishGroup(t, ctx, svc, validationSvc, t1, "t1-mid")
	publishGroup(t, ctx, svc, validationSvc, sentinelId, "s-live")
	waitForResponses(t, stream, 10*time.Second, "sentinel live during the wave",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsTagged(rs, 0), "s-live")) >= 1
		})

	close(gate)
	resps := waitForResponses(t, stream, 10*time.Second, "wave 6 complete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 2 })

	wave := groupMsgsTagged(resps, 6)
	require.Len(t, wave, 8, "6 history + 2 folded live messages, all tagged 6")
	require.Len(t, groupMsgsWithData(wave, "t1-mid"), 1, "t1's folded message arrives exactly once")
	require.Len(t, groupMsgsWithData(wave, "t2-mid"), 1, "t2's folded message arrives exactly once")
	requireAscendingGroupIds(t, wave, "wave 6 replay folded across topics")
	require.Empty(t, groupMsgsWithData(groupMsgsTagged(resps, 0), "t1-mid"),
		"the folded messages must not also be delivered live")
	require.Empty(t, groupMsgsWithData(groupMsgsTagged(resps, 0), "t2-mid"),
		"the folded messages must not also be delivered live")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_WelcomeWaveScanSkipsUnparseableRows pins the welcome fetcher's paging
// contract at the Subscribe level: a row the store cannot parse (an unknown message_type
// from a future writer) still consumes its topic's raw LIMIT slot, so the fetcher must
// advance the topic's cursor from its raw scan position and retire it only on a short RAW
// count — paging by the parsed slice would silently truncate the wave at the first
// skipped row inside a full turn.
func TestSubscribe_WelcomeWaveScanSkipsUnparseableRows(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, _, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.scanTopicPageLimit = 4 // small pages so the corrupt row sits inside a full page

	inst := []byte(test.RandomString(32))
	hpke := []byte(test.RandomString(32))
	const total = 10
	populateWelcomeMessages(t, ctx, svc, inst, hpke, total, "wel")

	// Learn the real ids, then mark one INSIDE the first full raw turn (not the final
	// turn) as unparseable: it keeps its raw slot but drops from the parsed slice.
	all, progress, err := svc.readOnlyStore.QueryWelcomeMessagesWaveScan(
		ctx,
		[]mlsstore.WelcomeCatchup{{InstallationKey: inst}},
		0,
		math.MaxUint64,
		total,
	)
	require.NoError(t, err)
	require.Equal(t, total, progress[string(inst)].RawCount)
	corruptID := welcomeID(all[2])
	svc.readOnlyStore = &fakeReadStore{
		ReadMlsStore:      svc.readOnlyStore,
		corruptWelcomeIDs: map[uint64]bool{corruptID: true},
	}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(welcomeTopic(inst), 0)},
		MutateId: 6,
	}))
	resps := waitForResponses(t, stream, 10*time.Second, "CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })
	require.Equal(t, []uint64{6}, catchupCompletesFrom(resps))

	wave := welcomeMsgsTagged(resps, 6)
	require.Len(t, wave, total-1,
		"every parseable welcome must arrive; the wave must not truncate at the skipped row")
	requireAscendingWelcomeIds(t, wave, "welcome wave replay around the skipped row")
	for _, m := range wave {
		require.NotEqual(t, corruptID, welcomeID(m), "the unparseable row must not be delivered")
	}

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_EveryHistoryPageCarriesWaveTag pins multi-page tag stamping: every
// data-bearing frame before a wave's CatchupComplete carries the wave's mutate_id — no
// page of a paginated replay may leak out tagged live.
func TestSubscribe_EveryHistoryPageCarriesWaveTag(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()
	svc.scanTopicPageLimit = 50 // small pages so a modest history spans several

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	const history = 117
	populateGroupMessages(t, ctx, svc, groupA, history, "hist")

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 1,
	}))
	resps := waitForResponses(t, stream, 15*time.Second, "full history + CatchupComplete",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsFrom(rs)) >= history && len(catchupCompletesFrom(rs)) >= 1
		})
	require.Equal(t, []uint64{1}, catchupCompletesFrom(resps))

	ccIdx := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		return r.GetV1().GetCatchupComplete() != nil
	})
	dataFrames := 0
	for _, r := range resps[:ccIdx] {
		msgs := r.GetV1().GetMessages()
		if msgs == nil ||
			(len(msgs.GetGroupMessages()) == 0 && len(msgs.GetWelcomeMessages()) == 0) {
			continue
		}
		dataFrames++
		require.Equal(t, uint64(1), msgs.GetMutateId(),
			"every data-bearing frame before CatchupComplete must carry the wave's tag")
	}
	require.Greater(t, dataFrames, 2, "the replay must have spanned multiple pages")

	// The live tail after completion is tagged 0.
	publishGroup(t, ctx, svc, validationSvc, groupA, "a-live")
	resps = waitForResponses(t, stream, 5*time.Second, "live tail",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsFrom(rs), "a-live")) >= 1
		})
	liveIdx := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		return len(groupMsgsWithData(r.GetV1().GetMessages().GetGroupMessages(), "a-live")) > 0
	})
	require.Equal(t, uint64(0), resps[liveIdx].GetV1().GetMessages().GetMutateId(),
		"the live frame must carry the live tag")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_HistoryOnlyMidReplayPublishNotDelivered pins the history_only seam: a
// message published while a history_only replay is in flight is never delivered — the
// topic has no live registration and the wave's ceiling predates the publish — while live
// flow for other topics is uninterrupted.
func TestSubscribe_HistoryOnlyMidReplayPublishNotDelivered(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	sentinelId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "histA")

	gate := make(chan struct{})
	parked := make(chan struct{})
	svc.readOnlyStore = &fakeReadStore{
		ReadMlsStore: svc.readOnlyStore,
		groupGate:    gate,
		gateGroup:    groupA,
		scanParked:   parked,
	}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(sentinelId), 0),
		},
		MutateId: 1,
	}))
	waitForResponses(t, stream, 10*time.Second, "sentinel live",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
		},
		HistoryOnly: true,
		MutateId:    2,
	}))
	// The wave pins its ceiling before parking on its first page; publish only after, so
	// "mid-replay" is deterministically above the ceiling.
	select {
	case <-parked:
	case <-time.After(5 * time.Second):
		t.Fatal("the history_only wave never reached its gated scan")
	}
	publishGroup(t, ctx, svc, validationSvc, groupA, "mid-replay")
	publishGroup(t, ctx, svc, validationSvc, sentinelId, "s-live")
	// The dbWorker dispatches strictly in id order, so the sentinel's arrival proves the
	// dispatcher already handled (and, lacking a live registration, dropped) "mid-replay".
	waitForResponses(t, stream, 10*time.Second, "sentinel live during the replay",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsTagged(rs, 0), "s-live")) >= 1
		})

	close(gate)
	resps := waitForResponses(t, stream, 10*time.Second, "history_only wave complete",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 2 })
	require.Equal(t, []uint64{1, 2}, catchupCompletesFrom(resps))

	require.Len(t, groupMsgsTagged(resps, 2), 5, "exactly the history, tagged with the wave")
	require.Empty(t, groupMsgsWithData(groupMsgsFrom(resps), "mid-replay"),
		"a mid-replay publish to a history_only topic must not be delivered under any tag")

	markerIdx := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		return topicsLiveHas(r, groupTopic(groupA))
	})
	cc2 := frameIndex(resps, func(r *mlsv1.SubscribeResponse) bool {
		cc := r.GetV1().GetCatchupComplete()
		return cc != nil && cc.GetMutateId() == 2
	})
	require.GreaterOrEqual(t, markerIdx, 0, "the history_only wave must announce its marker")
	require.Greater(t, cc2, markerIdx, "the marker must precede the wave's CatchupComplete")
	require.Len(t, groupMsgsWithData(groupMsgsTagged(resps, 0), "s-live"), 1,
		"the sentinel's live flow must be uninterrupted")

	stream.closeSend()
	require.NoError(t, <-errCh)
}

// TestSubscribe_DisconnectMidWaveCleansUp verifies teardown with a gated wave and nonempty
// pending: cancelling the client context returns the handler promptly, and the goroutines
// the stream spawned (sender, frame reader, fetcher) all exit.
func TestSubscribe_DisconnectMidWaveCleansUp(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	sentinelId := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "hist")

	gate := make(chan struct{})
	svc.readOnlyStore = &fakeReadStore{
		ReadMlsStore: svc.readOnlyStore,
		groupGate:    gate,
		gateGroup:    groupA,
	}

	before := runtime.NumGoroutine()

	streamCtx, streamCancel := context.WithCancel(ctx)
	defer streamCancel()
	stream := newFakeSubscribeStream(streamCtx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()

	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(sentinelId), 0),
		},
		MutateId: 1,
	}))
	waitForResponses(t, stream, 10*time.Second, "sentinel live",
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })

	// Wave 2 parks on the gate; a live message lands in its pending (sentinel-sequenced).
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds:     []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
		MutateId: 2,
	}))
	publishGroup(t, ctx, svc, validationSvc, groupA, "a-mid")
	publishGroup(t, ctx, svc, validationSvc, sentinelId, "s-live")
	waitForResponses(t, stream, 10*time.Second, "sentinel live during the wave",
		func(rs []*mlsv1.SubscribeResponse) bool {
			return len(groupMsgsWithData(groupMsgsTagged(rs, 0), "s-live")) >= 1
		})

	// The client disconnects while the fetcher is parked and pending is nonempty.
	streamCancel()
	select {
	case err := <-errCh:
		require.NoError(t, err, "a client cancellation is not a server error")
	case <-time.After(2 * time.Second):
		t.Fatal("Subscribe did not return after the client disconnected mid-wave")
	}
	close(gate) // release the parked fetcher only after the handler returned

	// Every stream goroutine must exit. No leak-check helper exists in this repo, so bound
	// the goroutine count instead: generous slack, retried for a few seconds.
	require.Eventually(t, func() bool {
		return runtime.NumGoroutine() <= before+2
	}, 5*time.Second, 50*time.Millisecond, "stream goroutines did not exit after disconnect")
}
