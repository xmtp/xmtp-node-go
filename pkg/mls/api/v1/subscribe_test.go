package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
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
	if f.blockSend != nil {
		<-f.blockSend
	}
	f.mu.Lock()
	f.sent = append(f.sent, resp)
	f.mu.Unlock()
	return nil
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
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupId), 0)},
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
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(sentinelId), 0)},
	}))
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{
			addSub(groupTopic(groupA), 0),
		},
		HistoryOnly: true,
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
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
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
// later re-add replays the history again (XIP-83 replay-after-remove).
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

	// Remove, then re-add from cursor 0: the cleared floor lets the history replay.
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Removes: [][]byte{groupTopic(groupA)},
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
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 2 },
	)
	require.Len(t, groupMsgsFrom(resps), 10, "remove + re-add must replay the full history")
	require.Equal(t, []uint64{1, 2}, catchupCompletesFrom(resps))

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

// TestSubscribe_InboundFramesResetIdleTimer verifies that inbound client frames which produce
// no response (here, no-op removes) still reset the idle timer, so the node does not ping or
// reap a stream the client is actively using.
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
	// only. A client streaming response-free frames (no-op removes) must NOT suppress it —
	// otherwise a peer that sends but never reads could never be reaped.
	done := make(chan struct{})
	defer close(done)
	go func() {
		dummy := groupTopic([]byte(test.RandomString(32)))
		for {
			select {
			case <-done:
				return
			default:
			}
			stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{Removes: [][]byte{dummy}}))
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

// TestQueryGroupMessagesBatch_CursorAboveInt64ReturnsNothing exercises clampCursor directly:
// the Subscribe-level test is masked by the writer's high-water mark, so guard the store here.
func TestQueryGroupMessagesBatch_CursorAboveInt64ReturnsNothing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	populateGroupMessages(t, ctx, svc, groupA, 5, "hist")

	msgs, err := svc.readOnlyStore.QueryGroupMessagesBatch(
		ctx,
		[]mlsstore.GroupCatchup{{GroupID: groupA, IdCursor: math.MaxUint64}},
		50,
	)
	require.NoError(t, err)
	require.Empty(t, msgs, "a cursor above MaxInt64 must clamp to no rows, not wrap negative")
}

// TestSubscribe_MultiPageCatchUp exercises catch-up pagination across more than
// catchUpPerGroupLimit messages: the active-loop continuation plus open-on-final-page wave
// accounting, with exactly one TopicsLive and one CatchupComplete.
func TestSubscribe_MultiPageCatchUp(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, _, validationSvc, cleanup := newTestService(t, ctx)
	defer cleanup()

	groupA := []byte(test.RandomString(32))
	mockValidateGroupMessages(validationSvc, groupA)
	const history = catchUpPerGroupLimit*2 + 17
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
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
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
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), cursor)},
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
// QueryGroupMessagesBatch fail, and groupGate (if non-nil) blocks it until released, letting
// a test hold a topic mid-catch-up. All other methods delegate to the embedded store.
type fakeReadStore struct {
	mlsstore.ReadMlsStore
	groupErr  error
	groupGate chan struct{}
}

func (f *fakeReadStore) QueryGroupMessagesBatch(
	ctx context.Context,
	filters []mlsstore.GroupCatchup,
	perGroupLimit int32,
) ([]*mlsv1.GroupMessage, error) {
	if f.groupGate != nil {
		select {
		case <-f.groupGate:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.groupErr != nil {
		return nil, f.groupErr
	}
	return f.ReadMlsStore.QueryGroupMessagesBatch(ctx, filters, perGroupLimit)
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
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
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
		func(rs []*mlsv1.SubscribeResponse) bool { return len(catchupCompletesFrom(rs)) >= 1 })
	require.Equal(t, []uint64{7}, catchupCompletesFrom(resps),
		"the removed topic's wave completes from the remove, echoing its mutate_id")
	require.Empty(
		t,
		groupMsgsFrom(resps),
		"catch-up was still blocked, so no history was delivered",
	)

	close(gate) // release the orphaned fetch; its stale page must be dropped
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
	gate := make(chan struct{})
	defer close(gate)
	svc.readOnlyStore = &fakeReadStore{ReadMlsStore: svc.readOnlyStore, groupGate: gate}

	stream := newFakeSubscribeStream(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Subscribe(stream) }()
	stream.send(subReqMutate(&mlsv1.SubscribeRequest_V1_Mutate{
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
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
		Adds: []*mlsv1.SubscribeRequest_V1_Mutate_Subscription{addSub(groupTopic(groupA), 0)},
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
