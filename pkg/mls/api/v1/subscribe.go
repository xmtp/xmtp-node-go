package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	v1proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// subscribePingInterval is how long a Subscribe stream may be idle before the
	// node sends a liveness Ping (XIP-83 server requirement 4; RECOMMENDED <= 30s).
	subscribePingInterval = 30 * time.Second
	// subscribePongDeadline is how long the node waits for the client's Pong before
	// reaping the stream (XIP-83 RECOMMENDED <= the ping interval).
	subscribePongDeadline = subscribePingInterval
	// subscribeBacklog is the message-channel buffer for a Subscribe connection.
	subscribeBacklog = 4096
	// sendQueueDepth buffers frame batches between the writer and the sender goroutine, so the
	// writer is not parked by a slow stream.Send. Small: it only smooths Send latency.
	sendQueueDepth = 8

	// Batched catch-up tuning (XIP-83). chunk = groups per DB round-trip;
	// perGroupLimit = rows per group per round; concurrency = chunks in flight.
	// The composite (group_id, id) index makes each per-group seek cheap, so the
	// binding constraint is payload bytes, not the planner. All tunable.
	catchUpChunkSize     = 256
	catchUpPerGroupLimit = 50
	catchUpConcurrency   = 4
	// catchUpChannelBuffer smooths handoff from the catch-up fetchers to the writer;
	// when full, fetchers block (backpressure) rather than racing ahead of the client.
	catchUpChannelBuffer = 64

	// maxPendingBytes caps live messages buffered while topics catch up. Exceeding it
	// drops the stream (the client reconnects from its cursor) rather than risk OOM.
	maxPendingBytes = 64 << 20 // 64 MiB
	// maxFrameBytes targets server message frames well under gRPC's default 4 MiB limit. It is a
	// best-effort cap: messages are packed up to this size but never split, so a single message
	// larger than maxFrameBytes is still emitted in its own frame. Individual message size is
	// bounded below gRPC's hard limit by the publish path, so a single-message frame still fits.
	maxFrameBytes = 2 << 20 // 2 MiB
)

// Topic kinds shared with the decentralized backend (XIP-49 §3.3.2): the first
// byte of a wire topic is the kind, the remainder is the identifier.
const (
	topicKindGroupMessagesV1   = 0x00 // identifier = group_id
	topicKindWelcomeMessagesV1 = 0x01 // identifier = installation_key
)

// splitTopic validates a kind-prefixed wire topic and returns (kind, identifier).
func splitTopic(t []byte) (byte, []byte, error) {
	if len(t) < 2 {
		return 0, nil, fmt.Errorf("topic must be a kind byte plus an identifier")
	}
	kind := t[0]
	if kind != topicKindGroupMessagesV1 && kind != topicKindWelcomeMessagesV1 {
		return 0, nil, fmt.Errorf("unsupported topic kind %d", kind)
	}
	return kind, t[1:], nil
}

// buildMLSTopic maps a (kind, identifier) pair to the dispatcher/content topic
// string used as the canonical per-topic key throughout Subscribe. The kind is
// baked into the string (g- vs w- prefix) so a group and a welcome with the same
// identifier never collide.
func buildMLSTopic(kind byte, id []byte) string {
	if kind == topicKindWelcomeMessagesV1 {
		return topic.BuildMLSV1WelcomeTopic(id)
	}
	return topic.BuildMLSV1GroupTopic(id)
}

// catchUpBatch is one unit of fetched catch-up history handed from a fetcher goroutine to
// the single writer. opened lists internal topics that finished catch-up in this batch —
// the writer flushes their buffered live messages, opens their gate, and announces them in
// a TopicsLive frame; openedWire carries the matching kind-prefixed wire topics the
// announcement echoes back to the client. wave identifies the Mutate whose adds this batch
// serves; the writer emits that wave's CatchupComplete once all its topics have opened. A
// non-nil err means the fetch failed and the writer should tear the stream down (the
// client reconnects from its cursor) rather than emit a misleading CatchupComplete or a
// history gap.
type catchUpBatch struct {
	groupMsgs   []*mlsv1.GroupMessage
	welcomeMsgs []*mlsv1.WelcomeMessage
	opened      []string
	openedWire  [][]byte
	wave        int
	err         error
}

// waveState tracks one Mutate's in-flight catch-up: how many of its topics have
// yet to open, and the client's correlation id to echo on its CatchupComplete.
type waveState struct {
	remaining int
	mutateID  uint64
}

// Subscribe is the XIP-83 bidirectional subscription. One long-lived stream that the
// client mutates in place via add/remove group & welcome deltas (no reconnect on
// membership change), with a WebSocket-style ping/pong so the client detects silent
// stream death and the node reaps a peer that has gone away. Group and welcome messages
// share the stream. See XIP-83.
//
// Concurrency model: SINGLE WRITER. The select loop below is the sole owner of every piece
// of mutable state (the high-water marks, the catch-up gate, the pending buffer, the ping
// bookkeeping). It is the only goroutine that decides WHAT to send and in what order; the
// actual stream.Send runs on one dedicated sender goroutine fed by an ordered channel, so a
// slow client can never park the writer (it stays free to run the ping/pong reap). Every
// other goroutine — the frame reader, the catch-up fetchers, the sender — is a pure producer
// or consumer that touches no writer-owned state. There are no mutexes: serialization is by
// single-threadedness, like a Rust task that owns its socket and an mpsc receiver. Ordering
// (history before live, no dupes) falls out of the order the writer enqueues, backstopped by
// the per-topic high-water mark.
func (s *Service) Subscribe(stream mlsv1.MlsApi_SubscribeServer) error {
	log := s.log.Named("subscribe")
	log.Info("subscription started")
	defer log.Info("subscription stopped")

	// Tonic-based clients need an initial header (see SubscribeGroupMessages).
	_ = stream.SendHeader(metadata.Pairs("subscribed", "true"))

	ctx := stream.Context()

	wrap := func(v1 *mlsv1.SubscribeResponse_V1) *mlsv1.SubscribeResponse {
		return &mlsv1.SubscribeResponse{Version: &mlsv1.SubscribeResponse_V1_{V1: v1}}
	}

	// ----- Writer-owned state. Touched ONLY by the select loop below. -----
	highWaterMarks := make(map[string]uint64)       // topic (kind-prefixed) -> last id sent
	catchingUp := make(map[string]bool)             // topic -> catch-up in progress
	pending := make(map[string][]*v1proto.Envelope) // live held while a topic catches up
	subscribed := make(map[string]struct{})         // topics live-registered on the dispatcher
	pendingBytes := 0
	lastActivity := time.Now()
	var awaitingPong bool
	var pingNonce uint64
	var pingSentAt time.Time
	subscribedTopics := 0
	// Catch-up wave bookkeeping: each Mutate that adds subscriptions is a wave, and the
	// writer emits one CatchupComplete (echoing the Mutate's mutate_id) per wave once all
	// of its topics have opened. topicWave maps a not-yet-opened topic to its wave so a
	// remove mid-catch-up can free the wave slot it would otherwise wait on forever.
	mutateSeen := false
	nextWave := 0
	waves := make(map[int]waveState)
	topicWave := make(map[string]int)
	// halfClosed: the client closed its send direction; finish in-flight waves, then close.
	halfClosed := false

	defer func() {
		if subscribedTopics > 0 {
			metrics.EmitUnsubscribeTopics(ctx, log, subscribedTopics)
		}
	}()

	// One mutable subscription for the whole stream, grown/shrunk in place.
	sub := s.subDispatcher.NewSubscription(subscribeBacklog)
	defer sub.Unsubscribe()

	// stream.Send runs on a dedicated sender goroutine, NOT the writer, so a client that stops
	// reading (a stalled Send) can never park the writer — it stays free to run the ping/pong
	// reap and tear the stream down. The writer hands frame batches to the sender over a
	// bounded channel; the sender is the SOLE caller of stream.Send (order preserved by the
	// single channel). An async Send error surfaces on sendErrCh (observed in send() and in the
	// main select); a handoff that blocks past the pong deadline (sender wedged on a non-reading
	// client and the buffer filled) fails the stream too.
	outbound := make(chan []*mlsv1.SubscribeResponse, sendQueueDepth)
	sendErrCh := make(chan error, 1)
	senderDone := make(chan struct{})
	go func() {
		defer close(senderDone)
		for batch := range outbound {
			for _, resp := range batch {
				if err := stream.Send(resp); err != nil {
					select {
					case sendErrCh <- err:
					default:
					}
					return
				}
			}
		}
	}()
	var closeOutboundOnce sync.Once
	closeOutbound := func() { closeOutboundOnce.Do(func() { close(outbound) }) }
	defer closeOutbound() // sender exits when this drains (or when a Send error returns it)

	// flush closes the outbound queue and waits for the sender to drain it, so a GRACEFUL
	// completion (the bounded-catch-up half-close) delivers every queued frame before the
	// handler returns and gRPC closes the stream. Without it the tail of a drain would be
	// lost. Bounded by the pong deadline / ctx so a client that stopped reading mid-drain
	// cannot wedge the close; if that bound trips the drain did NOT finish, so flush returns
	// DeadlineExceeded (never a false OK). Only the clean-completion paths call this; error
	// teardowns just return (the sender exits when the stream closes).
	flush := func() error {
		closeOutbound()
		select {
		case <-senderDone:
			return nil // the sender drained every queued frame
		case <-ctx.Done():
			return nil // client disconnected; gRPC surfaces the cancellation
		case <-time.After(s.pongDeadline):
			// The sender is still blocked in stream.Send (a slow or non-reading client),
			// so the queued frames — the bounded catch-up's history tail and its
			// CatchupComplete — were NOT delivered. This is NOT a successful completion:
			// fail rather than return OK and mislead the client into believing the drain
			// finished with a truncated catch-up.
			return status.Errorf(
				codes.DeadlineExceeded,
				"flush timed out waiting for sender to drain",
			)
		}
	}

	// send is the ONLY path frames take to the client (this includes Started). The single
	// writer calls it, in order. lastActivity advances when a batch is accepted by the sender
	// and is what gates the liveness Ping, so the Ping probes the client's receive path on a
	// send-idle schedule (it is deliberately NOT reset by inbound frames — see requestChannel).
	send := func(batches ...[]*mlsv1.SubscribeResponse) error {
		var flat []*mlsv1.SubscribeResponse
		for _, batch := range batches {
			flat = append(flat, batch...)
		}
		if len(flat) == 0 {
			return nil
		}
		select {
		case outbound <- flat:
			lastActivity = time.Now()
			return nil
		case err := <-sendErrCh:
			return err
		case <-ctx.Done():
			return nil
		case <-s.ctx.Done():
			return status.Errorf(codes.Unavailable, "service is shutting down")
		case <-time.After(s.pongDeadline):
			return status.Errorf(codes.Unavailable, "send stalled; client not reading")
		}
	}

	// buildGroupFrames dedups by group id (advancing the high-water mark, so no duplicates
	// across catch-up/live) and packs the survivors into <=maxFrameBytes frames.
	buildGroupFrames := func(msgs []*mlsv1.GroupMessage) []*mlsv1.SubscribeResponse {
		var frames []*mlsv1.SubscribeResponse
		var frame []*mlsv1.GroupMessage
		frameBytes := 0
		flush := func() {
			if len(frame) == 0 {
				return
			}
			frames = append(frames, wrap(&mlsv1.SubscribeResponse_V1{
				Response: &mlsv1.SubscribeResponse_V1_Messages_{
					Messages: &mlsv1.SubscribeResponse_V1_Messages{GroupMessages: frame},
				},
			}))
			frame = nil
			frameBytes = 0
		}
		for _, m := range msgs {
			key := topic.BuildMLSV1GroupTopic(m.GetV1().GetGroupId())
			if highWaterMarks[key] >= m.GetV1().GetId() {
				continue
			}
			highWaterMarks[key] = m.GetV1().GetId()
			sz := len(m.GetV1().GetData())
			if frameBytes+sz > s.maxFrameBytes && len(frame) > 0 {
				flush()
			}
			frame = append(frame, m)
			frameBytes += sz
		}
		flush()
		return frames
	}

	// buildWelcomeFrames is the welcome-topic analogue of buildGroupFrames.
	buildWelcomeFrames := func(msgs []*mlsv1.WelcomeMessage) []*mlsv1.SubscribeResponse {
		var frames []*mlsv1.SubscribeResponse
		var frame []*mlsv1.WelcomeMessage
		frameBytes := 0
		flush := func() {
			if len(frame) == 0 {
				return
			}
			frames = append(frames, wrap(&mlsv1.SubscribeResponse_V1{
				Response: &mlsv1.SubscribeResponse_V1_Messages_{
					Messages: &mlsv1.SubscribeResponse_V1_Messages{WelcomeMessages: frame},
				},
			}))
			frame = nil
			frameBytes = 0
		}
		for _, m := range msgs {
			key := topic.BuildMLSV1WelcomeTopic(welcomeInstallationKey(m))
			id := welcomeID(m)
			if highWaterMarks[key] >= id {
				continue
			}
			highWaterMarks[key] = id
			sz := len(welcomeData(m))
			if frameBytes+sz > s.maxFrameBytes && len(frame) > 0 {
				flush()
			}
			frame = append(frame, m)
			frameBytes += sz
		}
		flush()
		return frames
	}

	// buildOpenGateFrames drains a topic's live messages buffered during catch-up (deduped)
	// into frames and clears its gate, so subsequent live messages send directly.
	buildOpenGateFrames := func(topicStr string) []*mlsv1.SubscribeResponse {
		buffered := pending[topicStr]
		delete(pending, topicStr)
		delete(catchingUp, topicStr)
		if topic.IsMLSV1Welcome(topicStr) {
			welcomes := make([]*mlsv1.WelcomeMessage, 0, len(buffered))
			for _, env := range buffered {
				pendingBytes -= len(env.Message)
				if m, err := getWelcomeMessageFromEnvelope(env); err == nil {
					welcomes = append(welcomes, m)
				}
			}
			return buildWelcomeFrames(welcomes)
		}
		groups := make([]*mlsv1.GroupMessage, 0, len(buffered))
		for _, env := range buffered {
			pendingBytes -= len(env.Message)
			if m, err := getGroupMessageFromEnvelope(env); err == nil {
				groups = append(groups, m)
			}
		}
		return buildGroupFrames(groups)
	}

	startedFrame := func() []*mlsv1.SubscribeResponse {
		return []*mlsv1.SubscribeResponse{wrap(&mlsv1.SubscribeResponse_V1{
			Response: &mlsv1.SubscribeResponse_V1_Started_{
				Started: &mlsv1.SubscribeResponse_V1_Started{
					KeepaliveIntervalMs: uint32(s.pingInterval / time.Millisecond),
				},
			},
		})}
	}
	catchupCompleteFrame := func(mutateID uint64) []*mlsv1.SubscribeResponse {
		return []*mlsv1.SubscribeResponse{wrap(&mlsv1.SubscribeResponse_V1{
			Response: &mlsv1.SubscribeResponse_V1_CatchupComplete_{
				CatchupComplete: &mlsv1.SubscribeResponse_V1_CatchupComplete{
					MutateId: mutateID,
				},
			},
		})}
	}

	// dropTopic removes one topic from the stream: it stops live delivery, clears the
	// per-stream cursor floor (so a later re-add can replay from a lower cursor — XIP-83),
	// discards any live messages buffered during its catch-up, and — if the topic was still
	// mid catch-up — frees its slot in the owning wave, emitting that wave's CatchupComplete
	// if it was the last outstanding topic. Writer-goroutine only; never runs after
	// half-close (mutations stop then).
	dropTopic := func(t string) error {
		sub.Remove(t)
		delete(highWaterMarks, t)
		// The gauge tracks live topics, so it (and the metric) move only when a genuinely
		// live topic leaves — keeping subscribedTopics == len(subscribed) through removes,
		// resets, and history_only-over-live, and balanced against the teardown defer.
		if _, live := subscribed[t]; live {
			delete(subscribed, t)
			subscribedTopics--
			metrics.EmitUnsubscribeTopics(ctx, log, 1)
		}
		if catchingUp[t] {
			for _, env := range pending[t] {
				pendingBytes -= len(env.Message)
			}
			delete(pending, t)
			delete(catchingUp, t)
		}
		if wave, ok := topicWave[t]; ok {
			delete(topicWave, t)
			if w, ok := waves[wave]; ok {
				w.remaining--
				if w.remaining > 0 {
					waves[wave] = w
				} else {
					delete(waves, wave)
					if err := send(catchupCompleteFrame(w.mutateID)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	// Started must be the first frame, before any catch-up, so proxied/buffered transports
	// keep the connection open (XIP-83 server requirement 1). Sent here on the sole
	// goroutine, before any producer is spawned.
	if err := send(startedFrame()); err != nil {
		return err
	}

	// ----- Producers (no writer-owned state, no stream.Send) -----

	catchUpCh := make(chan catchUpBatch, catchUpChannelBuffer)
	forward := func(b catchUpBatch) {
		select {
		case catchUpCh <- b:
		case <-ctx.Done():
		case <-s.ctx.Done():
		}
	}

	// catchUpGroups fetches catch-up history for the given groups — chunked across the DB
	// with bounded concurrency — and forwards each page to the writer. It owns no shared
	// state and never sends to the stream; it just queries and hands results over.
	catchUpGroups := func(adds []mlsstore.GroupCatchup, topics []string, wire [][]byte, wave int) {
		processChunk := func(chunk []mlsstore.GroupCatchup, chunkTopics []string, chunkWire [][]byte) {
			cursors := make([]uint64, len(chunk))
			for i := range chunk {
				cursors[i] = chunk[i].IdCursor
			}
			active := make([]int, len(chunk))
			for i := range chunk {
				active[i] = i
			}
			for len(active) > 0 {
				select {
				case <-ctx.Done():
					return
				case <-s.ctx.Done():
					return
				default:
				}
				filters := make([]mlsstore.GroupCatchup, len(active))
				for j, idx := range active {
					filters[j] = mlsstore.GroupCatchup{
						GroupID:  chunk[idx].GroupID,
						IdCursor: cursors[idx],
					}
				}
				msgs, err := s.readOnlyStore.QueryGroupMessagesBatch(
					ctx,
					filters,
					catchUpPerGroupLimit,
				)
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						log.Error("batch catch-up (group)", zap.Error(err))
						forward(catchUpBatch{err: err})
					}
					return
				}
				counts := make(map[string]int)
				lastID := make(map[string]uint64)
				for _, m := range msgs {
					gid := string(m.GetV1().GetGroupId())
					counts[gid]++
					lastID[gid] = m.GetV1().GetId()
				}
				var opened []string
				var openedWire [][]byte
				var next []int
				for _, idx := range active {
					gid := string(chunk[idx].GroupID)
					if counts[gid] == catchUpPerGroupLimit {
						cursors[idx] = lastID[gid]
						next = append(next, idx)
					} else {
						opened = append(opened, chunkTopics[idx])
						openedWire = append(openedWire, chunkWire[idx])
					}
				}
				forward(catchUpBatch{
					groupMsgs:  msgs,
					opened:     opened,
					openedWire: openedWire,
					wave:       wave,
				})
				active = next
			}
		}

		sem := make(chan struct{}, catchUpConcurrency)
		var chunkWg sync.WaitGroup
		for start := 0; start < len(adds); start += catchUpChunkSize {
			end := start + catchUpChunkSize
			if end > len(adds) {
				end = len(adds)
			}
			chunk, chunkTopics, chunkWire := adds[start:end], topics[start:end], wire[start:end]
			chunkWg.Add(1)
			sem <- struct{}{}
			go func(chunk []mlsstore.GroupCatchup, chunkTopics []string, chunkWire [][]byte) {
				defer chunkWg.Done()
				defer func() { <-sem }()
				processChunk(chunk, chunkTopics, chunkWire)
			}(chunk, chunkTopics, chunkWire)
		}
		chunkWg.Wait()
	}

	// catchUpWelcomes is the welcome-topic analogue of catchUpGroups.
	catchUpWelcomes := func(adds []mlsstore.WelcomeCatchup, topics []string, wire [][]byte, wave int) {
		processChunk := func(chunk []mlsstore.WelcomeCatchup, chunkTopics []string, chunkWire [][]byte) {
			cursors := make([]uint64, len(chunk))
			for i := range chunk {
				cursors[i] = chunk[i].IdCursor
			}
			active := make([]int, len(chunk))
			for i := range chunk {
				active[i] = i
			}
			for len(active) > 0 {
				select {
				case <-ctx.Done():
					return
				case <-s.ctx.Done():
					return
				default:
				}
				filters := make([]mlsstore.WelcomeCatchup, len(active))
				for j, idx := range active {
					filters[j] = mlsstore.WelcomeCatchup{
						InstallationKey: chunk[idx].InstallationKey,
						IdCursor:        cursors[idx],
					}
				}
				msgs, err := s.readOnlyStore.QueryWelcomeMessagesBatch(
					ctx,
					filters,
					catchUpPerGroupLimit,
				)
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						log.Error("batch catch-up (welcome)", zap.Error(err))
						forward(catchUpBatch{err: err})
					}
					return
				}
				counts := make(map[string]int)
				lastID := make(map[string]uint64)
				for _, m := range msgs {
					key := string(welcomeInstallationKey(m))
					counts[key]++
					lastID[key] = welcomeID(m)
				}
				var opened []string
				var openedWire [][]byte
				var next []int
				for _, idx := range active {
					key := string(chunk[idx].InstallationKey)
					if counts[key] == catchUpPerGroupLimit {
						cursors[idx] = lastID[key]
						next = append(next, idx)
					} else {
						opened = append(opened, chunkTopics[idx])
						openedWire = append(openedWire, chunkWire[idx])
					}
				}
				forward(catchUpBatch{
					welcomeMsgs: msgs,
					opened:      opened,
					openedWire:  openedWire,
					wave:        wave,
				})
				active = next
			}
		}

		sem := make(chan struct{}, catchUpConcurrency)
		var chunkWg sync.WaitGroup
		for start := 0; start < len(adds); start += catchUpChunkSize {
			end := start + catchUpChunkSize
			if end > len(adds) {
				end = len(adds)
			}
			chunk, chunkTopics, chunkWire := adds[start:end], topics[start:end], wire[start:end]
			chunkWg.Add(1)
			sem <- struct{}{}
			go func(chunk []mlsstore.WelcomeCatchup, chunkTopics []string, chunkWire [][]byte) {
				defer chunkWg.Done()
				defer func() { <-sem }()
				processChunk(chunk, chunkTopics, chunkWire)
			}(chunk, chunkTopics, chunkWire)
		}
		chunkWg.Wait()
	}

	// Read client frames in a dedicated goroutine (gRPC Recv blocks) and forward them to
	// the writer. This producer touches no writer-owned state. The channel is buffered so a
	// forwarded Pong is never left blocking here while the writer is busy in a Send — which
	// would otherwise let the ping deadline reap a stream whose client did answer.
	//
	// recvErr distinguishes a clean half-close (io.EOF — the client called CloseSend, the
	// bounded catch-up flow) from a transport failure; on the latter the writer fails the
	// RPC instead of reporting a false clean completion. The goroutine writes recvErr
	// strictly before close(requestChannel), and the writer reads it only after observing
	// the channel closed, so the close establishes the happens-before (no data race).
	requestChannel := make(chan *mlsv1.SubscribeRequest, 16)
	var recvErr error
	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				switch e, ok := status.FromError(err); {
				case ok && e.Code() == codes.Canceled:
					// client cancelled; ctx.Done covers teardown — not an error to surface
				case err == io.EOF || err == context.Canceled:
					// clean half-close
				default:
					log.Debug("reading subscription", zap.Error(err))
					recvErr = err
				}
				close(requestChannel)
				return
			}
			select {
			case requestChannel <- req:
			case <-ctx.Done():
				return
			case <-s.ctx.Done():
				return
			}
		}
	}()

	pingTicker := time.NewTicker(s.pingInterval)
	defer pingTicker.Stop()

	// ----- The writer. Single goroutine; owns all state and the socket. -----
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-s.ctx.Done():
			return status.Errorf(codes.Unavailable, "service is shutting down")

		case err := <-sendErrCh:
			// A stream.Send failed on the sender goroutine after the writer had moved on;
			// surface it here so the RPC ends in error rather than hanging.
			return err

		case req, ok := <-requestChannel:
			if !ok {
				// The frame reader closed the channel. A transport error means the stream
				// broke mid-flight: fail the RPC so the client reconnects from its cursor,
				// rather than report it as a successful completion.
				if recvErr != nil {
					return status.Errorf(codes.Unavailable, "stream recv failed: %v", recvErr)
				}
				// Otherwise it is a clean half-close (io.EOF): no more mutations or pongs can
				// arrive. Finish any in-flight catch-up waves first (the bounded catch-up
				// flow: Mutate then CloseSend, the server hangs up after the last wave), and
				// stop pinging — a half-closed peer cannot answer, and the bounded drain plus
				// gRPC's own transport timeouts cover liveness.
				if len(waves) == 0 {
					return flush()
				}
				halfClosed = true
				awaitingPong = false
				requestChannel = nil // a nil channel never fires; this case goes dormant
				continue
			}
			// NB: lastActivity is updated ONLY by a successful Send. The idle timer drives the
			// liveness Ping, which probes the client's RECEIVE path; inbound frames prove only
			// the send path, so they must NOT defer the Ping — else a client that streams frames
			// but never reads could suppress the probe (and the reap) forever.
			v1 := req.GetV1()
			if v1 == nil {
				// Unrecognized request version arm: fail rather than silently ignore, so a
				// forward-version client is not left waiting on a response (XIP-83 req 8).
				return status.Errorf(codes.InvalidArgument, "unrecognized SubscribeRequest version")
			}
			switch {
			case v1.GetMutate() != nil:
				m := v1.GetMutate()
				// Removes are applied before adds, so a topic appearing in both is reset:
				// removed (clearing its cursor floor), then re-added with a fresh catch-up.
				// dropTopic owns the gauge/metric for any topic that was actually live, so a
				// duplicate or unknown remove is harmless rather than drifting the count.
				for _, wire := range m.GetRemoves() {
					kind, id, err := splitTopic(wire)
					if err != nil {
						return status.Errorf(codes.InvalidArgument, "remove: %v", err)
					}
					if err := dropTopic(buildMLSTopic(kind, id)); err != nil {
						return err
					}
				}

				// history_only adds never touch the dispatcher: no live registration,
				// no gate, no pending buffer — a pure batched read with markers.
				historyOnly := m.GetHistoryOnly()
				// Collapse duplicate topics within this Mutate's adds, lowest id_cursor
				// winning, so a repeated topic resolves deterministically and a lower cursor
				// still drives the replay path below.
				type addReq struct {
					wire   []byte
					kind   byte
					id     []byte
					cursor uint64
				}
				addOrder := make([]string, 0, len(m.GetAdds()))
				addByTopic := make(map[string]*addReq, len(m.GetAdds()))
				for _, add := range m.GetAdds() {
					wire := add.GetTopic()
					kind, id, err := splitTopic(wire)
					if err != nil {
						return status.Errorf(codes.InvalidArgument, "add: %v", err)
					}
					t := buildMLSTopic(kind, id)
					cursor := add.GetIdCursor()
					if ex, ok := addByTopic[t]; ok {
						if cursor < ex.cursor {
							ex.cursor, ex.wire = cursor, wire
						}
						continue
					}
					addByTopic[t] = &addReq{wire: wire, kind: kind, id: id, cursor: cursor}
					addOrder = append(addOrder, t)
				}

				var groupAdds []mlsstore.GroupCatchup
				var groupTopics []string
				var groupWire [][]byte
				var welcomeAdds []mlsstore.WelcomeCatchup
				var welcomeTopics []string
				var welcomeWire [][]byte
				newlyLive := 0
				for _, t := range addOrder {
					a := addByTopic[t]
					// Re-adding a topic already active on this stream is a no-op unless its
					// cursor is below the current floor, which restarts catch-up to replay
					// (XIP-83). The floor seeds to the starting cursor on a fresh add (below),
					// so this compares against the topic's own start / last-sent id.
					if _, live := subscribed[t]; live || catchingUp[t] {
						// A history_only add is a one-shot bounded read that never registers
						// for live delivery. Targeting a topic already live (or catching up) on
						// this stream is contradictory: there is a single cursor floor per topic,
						// so honoring it would have to disturb the live subscription's floor — and
						// on the replay path (cursor below the floor) dropTopic would unsubscribe
						// the topic without re-registering it, silently severing a live tail.
						// Reject rather than guess (XIP-83 req 8 stance: fail contradictory input).
						if historyOnly {
							return status.Errorf(
								codes.InvalidArgument,
								"history_only add targets a topic already subscribed on this stream",
							)
						}
						if a.cursor >= highWaterMarks[t] {
							continue
						}
						if err := dropTopic(t); err != nil {
							return err
						}
					}
					highWaterMarks[t] = a.cursor // explicit starting floor
					if !historyOnly {
						catchingUp[t] = true // gate BEFORE Add: no live escapes before buffering
						sub.Add(t)
						subscribed[t] = struct{}{}
						newlyLive++
					}
					switch a.kind {
					case topicKindGroupMessagesV1:
						groupAdds = append(
							groupAdds,
							mlsstore.GroupCatchup{GroupID: a.id, IdCursor: a.cursor},
						)
						groupTopics = append(groupTopics, t)
						groupWire = append(groupWire, a.wire)
					case topicKindWelcomeMessagesV1:
						welcomeAdds = append(
							welcomeAdds,
							mlsstore.WelcomeCatchup{InstallationKey: a.id, IdCursor: a.cursor},
						)
						welcomeTopics = append(welcomeTopics, t)
						welcomeWire = append(welcomeWire, a.wire)
					}
				}
				if newlyLive > 0 {
					subscribedTopics += newlyLive
					metrics.EmitSubscribeTopics(ctx, log, newlyLive)
				}

				// Each mutate that catches up subscriptions starts a wave; its
				// CatchupComplete (echoing mutate_id) is emitted once all of the wave's
				// topics are live. A mutate whose adds were all already-live (no-ops) or that
				// added nothing still gets an immediate CatchupComplete — both so the client's
				// mutate_id is answered and so a client that subscribed nothing learns it is
				// live. A pure remove-only mutate after the first yields neither.
				adds := len(groupAdds) + len(welcomeAdds)
				switch {
				case adds > 0:
					wave := nextWave
					nextWave++
					waves[wave] = waveState{remaining: adds, mutateID: m.GetMutateId()}
					for _, t := range groupTopics {
						topicWave[t] = wave
					}
					for _, t := range welcomeTopics {
						topicWave[t] = wave
					}
					if len(groupAdds) > 0 {
						go catchUpGroups(groupAdds, groupTopics, groupWire, wave)
					}
					if len(welcomeAdds) > 0 {
						go catchUpWelcomes(welcomeAdds, welcomeTopics, welcomeWire, wave)
					}
				case len(m.GetAdds()) > 0 || !mutateSeen:
					if err := send(catchupCompleteFrame(m.GetMutateId())); err != nil {
						return err
					}
				}
				mutateSeen = true

			case v1.GetPing() != nil:
				nonce := v1.GetPing().GetNonce()
				if err := send([]*mlsv1.SubscribeResponse{wrap(&mlsv1.SubscribeResponse_V1{
					Response: &mlsv1.SubscribeResponse_V1_Pong{Pong: &mlsv1.Pong{Nonce: nonce}},
				})}); err != nil {
					return err
				}

			case v1.GetPong() != nil:
				// Only a Pong echoing the outstanding nonce clears the liveness deadline; a
				// stale or unsolicited Pong must not keep a half-dead stream alive.
				if v1.GetPong().GetNonce() == pingNonce {
					awaitingPong = false
				}
			}

		case b := <-catchUpCh:
			if b.err != nil {
				// A fetch error means catch-up is incomplete; fail fast so the client
				// reconnects from its cursor rather than receive a gap or a false
				// CATCHUP_COMPLETE.
				return status.Errorf(codes.Unavailable, "catch-up failed: %v", b.err)
			}
			// A topic can be removed (or reset) while its catch-up page is in flight. wanted
			// reports whether this batch's wave still owns the topic; history, the gate
			// flush, TopicsLive and the wave count all skip topics no longer wanted, so a
			// removed topic never gets stale history, a phantom marker, or a miscounted wave.
			wanted := func(t string) bool {
				wv, ok := topicWave[t]
				return ok && wv == b.wave
			}
			var history []*mlsv1.SubscribeResponse
			if len(b.groupMsgs) > 0 {
				kept := b.groupMsgs[:0]
				for _, m := range b.groupMsgs {
					if wanted(topic.BuildMLSV1GroupTopic(m.GetV1().GetGroupId())) {
						kept = append(kept, m)
					}
				}
				history = buildGroupFrames(kept)
			} else if len(b.welcomeMsgs) > 0 {
				kept := b.welcomeMsgs[:0]
				for _, m := range b.welcomeMsgs {
					if wanted(topic.BuildMLSV1WelcomeTopic(welcomeInstallationKey(m))) {
						kept = append(kept, m)
					}
				}
				history = buildWelcomeFrames(kept)
			}
			var openFrames []*mlsv1.SubscribeResponse
			var liveMarker []*mlsv1.SubscribeResponse
			openedCount := 0
			if len(b.opened) > 0 {
				marker := &mlsv1.SubscribeResponse_V1_TopicsLive{}
				for i, t := range b.opened {
					if wv, ok := topicWave[t]; !ok || wv != b.wave {
						continue // removed or reset mid-catch-up; its remove already settled it
					}
					delete(topicWave, t)
					openFrames = append(openFrames, buildOpenGateFrames(t)...)
					marker.Topics = append(marker.Topics, b.openedWire[i])
					openedCount++
				}
				if openedCount > 0 {
					liveMarker = []*mlsv1.SubscribeResponse{wrap(&mlsv1.SubscribeResponse_V1{
						Response: &mlsv1.SubscribeResponse_V1_TopicsLive_{TopicsLive: marker},
					})}
				}
			}
			// Order is just program order here: history, then the flushed pending buffer
			// (live messages that queued behind the catch-up — equally historical from the
			// client's perspective), and only then the TopicsLive marker, so every frame a
			// client sees after the marker really is live tail.
			if err := send(history, openFrames, liveMarker); err != nil {
				return err
			}
			if w, ok := waves[b.wave]; ok {
				w.remaining -= openedCount
				if w.remaining > 0 {
					waves[b.wave] = w
				} else {
					// The wave's last topic just went live; its CatchupComplete (echoing
					// the Mutate's id) follows the wave's final TopicsLive in program order.
					delete(waves, b.wave)
					if err := send(catchupCompleteFrame(w.mutateID)); err != nil {
						return err
					}
					if halfClosed && len(waves) == 0 {
						// Everything the client asked for is queued; return OK only if the
						// sender actually drains it (else flush returns DeadlineExceeded).
						return flush()
					}
				}
			}

		case env, open := <-sub.MessagesCh:
			if !open {
				return status.Errorf(codes.Aborted, "subscription closed: consumer too slow")
			}
			t := env.ContentTopic
			if catchingUp[t] {
				pending[t] = append(pending[t], env)
				pendingBytes += len(env.Message)
				if pendingBytes > s.maxPendingBytes {
					return status.Errorf(
						codes.ResourceExhausted,
						"catch-up buffer exceeded; reconnect from cursor",
					)
				}
				continue
			}
			var frames []*mlsv1.SubscribeResponse
			if topic.IsMLSV1Welcome(t) {
				if m, err := getWelcomeMessageFromEnvelope(env); err == nil {
					frames = buildWelcomeFrames([]*mlsv1.WelcomeMessage{m})
				} else {
					log.Error("error parsing welcome message", zap.Error(err))
				}
			} else {
				if m, err := getGroupMessageFromEnvelope(env); err == nil {
					frames = buildGroupFrames([]*mlsv1.GroupMessage{m})
				} else {
					log.Error("error parsing group message", zap.Error(err))
				}
			}
			if err := send(frames); err != nil {
				return err
			}

		case <-pingTicker.C:
			if halfClosed {
				continue // a half-closed peer cannot Pong; the wave drain bounds the stream
			}
			switch {
			case awaitingPong:
				if time.Since(pingSentAt) >= s.pongDeadline {
					return status.Errorf(codes.DeadlineExceeded, "no Pong within deadline")
				}
			case time.Since(lastActivity) >= s.pingInterval:
				pingNonce++
				if err := send([]*mlsv1.SubscribeResponse{wrap(&mlsv1.SubscribeResponse_V1{
					Response: &mlsv1.SubscribeResponse_V1_Ping{Ping: &mlsv1.Ping{Nonce: pingNonce}},
				})}); err != nil {
					return err
				}
				awaitingPong = true
				pingSentAt = time.Now()
			}
		}
	}
}

// welcomeInstallationKey / welcomeID / welcomeData read a WelcomeMessage regardless of
// which version (V1 or WelcomePointer) it carries.
func welcomeInstallationKey(m *mlsv1.WelcomeMessage) []byte {
	if v1 := m.GetV1(); v1 != nil {
		return v1.GetInstallationKey()
	}
	return m.GetWelcomePointer().GetInstallationKey()
}

func welcomeID(m *mlsv1.WelcomeMessage) uint64 {
	if v1 := m.GetV1(); v1 != nil {
		return v1.GetId()
	}
	return m.GetWelcomePointer().GetId()
}

func welcomeData(m *mlsv1.WelcomeMessage) []byte {
	if v1 := m.GetV1(); v1 != nil {
		return v1.GetData()
	}
	return m.GetWelcomePointer().GetWelcomePointer()
}
