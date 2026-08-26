package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"sync/atomic"
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
	// node sends a liveness Ping (XIP-83 server requirement 6; RECOMMENDED <= 30s).
	subscribePingInterval = 30 * time.Second
	// subscribePongDeadline is how long the node waits for the client's Pong before
	// reaping the stream (XIP-83 RECOMMENDED <= the ping interval).
	subscribePongDeadline = subscribePingInterval
	// subscribeBacklog is the message-channel buffer for a Subscribe connection.
	subscribeBacklog = 4096
	// sendQueueDepth buffers frame batches between the writer and the sender goroutine, so the
	// writer is not parked by a slow stream.Send. Small: it only smooths Send latency.
	sendQueueDepth = 8
	// liveCoalesceMax bounds how many already-queued live messages the writer drains into
	// one send after receiving one, packing a burst of small per-message frames into fewer
	// (byte-batched) frames without blocking. Purely a fairness cap so a hot producer cannot
	// keep the writer in the live branch and starve ping/mutation handling; frame size is
	// separately bounded by maxFrameBytes.
	liveCoalesceMax = 256

	// Wave-scan catch-up tuning (XIP-83 server requirement 4, amended for batch
	// rotation). A wave prunes topics whose cursor already sits at the ceiling,
	// then replays the rest in rotating batches: one query per turn returns up
	// to catchUpTopicPageLimit rows for EACH topic in the batch, every fetched
	// row is delivered (turns never discard and re-fetch work), and ids ascend
	// within every turn and per topic across the wave. The per-topic guarantee
	// is the one the writer's high-water dedup and the client's replay guards
	// rely on; only the live lane needs total cursor order. The ceiling (the
	// newest id at wave start) pins every turn so the wave terminates under
	// sustained publishing.
	//
	// A turn is bounded by catchUpBatchTopics × catchUpTopicPageLimit = 16,384
	// rows: generous for real deep replays, which stream out in few round
	// trips, while still capping what a hostile subscription — many topics,
	// all far behind — can force into one query result. Topics retire from
	// the rotation individually the moment a turn returns fewer than their
	// per-topic limit. The lane's other two axes are bounded by the scan-slot
	// and byte-budget constants below; the channel buffer is sized in
	// turns-of-16,384, not messages.
	catchUpTopicPageLimit = 64
	catchUpBatchTopics    = 256
	// catchUpChannelBuffer smooths handoff from the catch-up fetchers to the writer;
	// when full, fetchers block (backpressure) rather than racing ahead of the
	// client. Sized in TURNS, each of which may now carry up to the full
	// per-turn row cap (16,384 messages): 4 buffered turns bounds the channel
	// at ~65k messages, the same order as the 64 × 512-row pages it previously
	// buffered. Typical turns are tiny or empty, so the smaller depth costs
	// real replays nothing.
	catchUpChannelBuffer = 4
	// catchUpMaxConcurrentScans bounds how many of one stream's catch-up scans
	// fetch at once. Each add-bearing Mutate's wave contributes up to two scans
	// (group + welcome); a scan with rows to replay owns a slot from before its
	// first wave-scan query until its done marker is in the channel, while a
	// fully-current wave — the dominant reconnect case — prunes slot-free and
	// completes without queuing (a parked current wave would only hold its
	// topics gated, live traffic accumulating against maxPendingBytes, for no
	// replay at all). Excess scans park on the slot channel (FIFO) and start
	// as slots free, so a burst of stale, topic-dense Mutates cannot multiply
	// full-turn row hands or per-stream query concurrency without bound. A
	// parked wave holds only its topic floors (state its waveState carries
	// anyway). Waves still overlap up to the cap (XIP-83 server requirement 8
	// promises concurrent waves, not unboundedly many), and cross-wave
	// completion order was never part of the contract.
	catchUpMaxConcurrentScans = 4
	// catchUpMaxPendingBytes caps the payload bytes catch-up scans have
	// fetched but the writer has not yet consumed — the catch-up mirror of
	// maxPendingBytes, which covers only the gated live lane. A scan reserves
	// a turn's bytes after its query returns (sizes are unknowable beforehand)
	// and parks until the writer frees room: backpressure, like the channel,
	// never a stream drop. An empty lane admits any single turn, so one turn
	// larger than the whole budget still replays (alone) instead of
	// deadlocking. The budget thus bounds fetched-but-unconsumed bytes at
	// max(budget, one turn); on top of that sit not-yet-reserved query
	// results in flight (≤ catchUpMaxConcurrentScans turns) and the
	// writer→sender pipeline, which the budget deliberately does not cover —
	// the writer frees a batch when it takes it, and delivery is separately
	// turn-bounded by sendQueueDepth plus the writer's and sender's hands.
	catchUpMaxPendingBytes = 64 << 20 // 64 MiB
	// maxMutateAdds bounds one Mutate's raw adds (pre-dedup — a stateless
	// check). Batch rotation keeps any single query small regardless, but the
	// wave still holds every topic's floor in memory for its lifetime. A
	// client with a larger set splits it across Mutates, whose waves run
	// concurrently (XIP-83 server requirement 8).
	maxMutateAdds = 100_000

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

// catchUpBatch is one unit of fetched catch-up history handed from a wave's fetcher
// goroutine to the single writer, in scan order. wave identifies the Mutate whose adds
// the batch serves; the writer stamps its frames with that wave's mutate_id (XIP-83
// server requirement 3). done marks the end of one kind's scan for the wave (the group
// and welcome scans run independently); when the wave's last scan is done the writer
// flushes the wave's gated live buffer, announces TopicsLive, emits CatchupComplete,
// and opens the gates. A non-nil err means the fetch failed and the writer should tear
// the stream down (the client reconnects from its cursor) rather than emit a misleading
// CatchupComplete or a history gap.
type catchUpBatch struct {
	groupMsgs   []*mlsv1.GroupMessage
	welcomeMsgs []*mlsv1.WelcomeMessage
	wave        int
	// bytes is the turn's payload size, reserved against the stream's catch-up
	// byte budget by the fetcher; the writer frees it when it takes the batch.
	bytes int
	done  bool
	err   error
}

// waveState tracks one Mutate's in-flight catch-up. A wave completes when its last
// scan drains (scansLeft reaches 0) — or early, when every topic it owned has been
// removed (owned reaches 0). topics/wire hold what its completion TopicsLive will
// announce, filtered against topicWave at completion so topics removed or reset
// mid-wave are not announced.
type waveState struct {
	mutateID  uint64
	scansLeft int
	owned     int
	topics    []string
	wire      [][]byte
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
			// The sender finished — but it may have stopped early on a Send error, leaving the
			// wave's terminal frames (history tail + CatchupComplete) unsent. Surface that rather
			// than a false OK that misleads the client into believing the drain completed.
			select {
			case err := <-sendErrCh:
				return err
			default:
				return nil
			}
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
	// sendTimer bounds a single send(); reused across calls (send runs only on the writer
	// goroutine) instead of allocating a time.After timer — live for the whole deadline — on
	// every send under high throughput.
	sendTimer := time.NewTimer(s.pongDeadline)
	stopTimer(sendTimer)
	defer sendTimer.Stop()
	send := func(batches ...[]*mlsv1.SubscribeResponse) error {
		var flat []*mlsv1.SubscribeResponse
		for _, batch := range batches {
			flat = append(flat, batch...)
		}
		if len(flat) == 0 {
			return nil
		}
		sendTimer.Reset(s.pongDeadline)
		defer stopTimer(sendTimer)
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
		case <-sendTimer.C:
			return status.Errorf(codes.Unavailable, "send stalled; client not reading")
		}
	}

	// buildGroupFrames dedups by group id (advancing the high-water mark, so no duplicates
	// across catch-up/live) and packs the survivors into <=maxFrameBytes frames, each
	// stamped with the catch-up wave that produced it — a Mutate's mutate_id for wave
	// replay, 0 for live tail (XIP-83 server requirement 3). A frame is exactly one or
	// the other; callers never mix lanes in one call.
	buildGroupFrames := func(msgs []*mlsv1.GroupMessage, mutateID uint64) []*mlsv1.SubscribeResponse {
		var frames []*mlsv1.SubscribeResponse
		var frame []*mlsv1.GroupMessage
		frameBytes := 0
		flush := func() {
			if len(frame) == 0 {
				return
			}
			frames = append(frames, wrap(&mlsv1.SubscribeResponse_V1{
				Response: &mlsv1.SubscribeResponse_V1_Messages_{
					Messages: &mlsv1.SubscribeResponse_V1_Messages{
						GroupMessages: frame,
						MutateId:      mutateID,
					},
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
	buildWelcomeFrames := func(msgs []*mlsv1.WelcomeMessage, mutateID uint64) []*mlsv1.SubscribeResponse {
		var frames []*mlsv1.SubscribeResponse
		var frame []*mlsv1.WelcomeMessage
		frameBytes := 0
		flush := func() {
			if len(frame) == 0 {
				return
			}
			frames = append(frames, wrap(&mlsv1.SubscribeResponse_V1{
				Response: &mlsv1.SubscribeResponse_V1_Messages_{
					Messages: &mlsv1.SubscribeResponse_V1_Messages{
						WelcomeMessages: frame,
						MutateId:        mutateID,
					},
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
				w.owned--
				if w.owned > 0 {
					waves[wave] = w
				} else {
					// Every topic the wave owned is gone; complete it now (there is
					// nothing left to announce) rather than wait for its scans, whose
					// remaining batches the writer drops once the wave is deleted.
					delete(waves, wave)
					if err := send(catchupCompleteFrame(w.mutateID)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	// completeWave finishes a wave whose scans have all drained: it flushes the live
	// messages buffered while its topics were gated — merged back into ascending id
	// order and stamped with the wave's mutate_id, since the wave folds them in (their
	// ids sit above the scan's ceiling) — then announces the surviving topics in one
	// TopicsLive and emits the wave's CatchupComplete. Only then do the gates open, so
	// a live (mutate_id 0) frame for a wave's topic is never delivered before its
	// CatchupComplete (XIP-83 server requirement 4: the seam). Topics removed or reset
	// mid-wave were settled by dropTopic and are skipped here.
	completeWave := func(wave int, w waveState) error {
		var groups []*mlsv1.GroupMessage
		var welcomes []*mlsv1.WelcomeMessage
		marker := &mlsv1.SubscribeResponse_V1_TopicsLive{}
		for i, t := range w.topics {
			if wv, ok := topicWave[t]; !ok || wv != wave {
				continue // removed or reset mid-wave; its remove already settled it
			}
			delete(topicWave, t)
			marker.Topics = append(marker.Topics, w.wire[i])
			if !catchingUp[t] {
				// history_only: never gated, nothing buffered — and no live registration
				// follows (both live-over-history_only directions are rejected in the Mutate
				// handler), so drop the floor here or one-shot reads leak it forever.
				delete(highWaterMarks, t)
				continue
			}
			delete(catchingUp, t)
			for _, env := range pending[t] {
				pendingBytes -= len(env.Message)
				if topic.IsMLSV1Welcome(t) {
					if m, err := getWelcomeMessageFromEnvelope(env); err == nil {
						welcomes = append(welcomes, m)
					} else {
						log.Error("error parsing welcome message", zap.Error(err))
					}
				} else {
					if m, err := getGroupMessageFromEnvelope(env); err == nil {
						groups = append(groups, m)
					} else {
						log.Error("error parsing group message", zap.Error(err))
					}
				}
			}
			delete(pending, t)
		}
		// Each topic's buffer is in dispatch (= id) order, but the wave's replay must
		// stay totally ordered across its topics: merge before framing. The frame
		// builders' high-water dedup then drops anything the scan already delivered —
		// sound only under the same id-visibility-order invariant the ceiling pin rests
		// on (see catchUpGroups).
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].GetV1().GetId() < groups[j].GetV1().GetId()
		})
		sort.Slice(welcomes, func(i, j int) bool {
			return welcomeID(welcomes[i]) < welcomeID(welcomes[j])
		})
		var liveMarker []*mlsv1.SubscribeResponse
		if len(marker.Topics) > 0 {
			liveMarker = []*mlsv1.SubscribeResponse{wrap(&mlsv1.SubscribeResponse_V1{
				Response: &mlsv1.SubscribeResponse_V1_TopicsLive_{TopicsLive: marker},
			})}
		}
		delete(waves, wave)
		return send(
			buildGroupFrames(groups, w.mutateID),
			buildWelcomeFrames(welcomes, w.mutateID),
			liveMarker,
			catchupCompleteFrame(w.mutateID),
		)
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

	// Catch-up lane bounds (see the catchUpMaxConcurrentScans /
	// catchUpMaxPendingBytes comments). Both are per-stream and shared by the
	// group and welcome fetchers; neither is writer-owned state — the writer
	// only ever frees, which cannot block. Guard non-positive tunables the
	// same way the fetchers guard theirs: a zero cap would park every scan
	// forever.
	maxScans := s.scanMaxConcurrent
	if maxScans < 1 {
		maxScans = 1
	}
	scanSlots := make(chan struct{}, maxScans)
	byteBudget := s.scanMaxPendingBytes
	if byteBudget < 1 {
		byteBudget = 1
	}
	var catchUpBytes atomic.Int64
	catchUpBytesFreed := make(chan struct{}, 1)

	// acquireScanSlot parks a scan until one of the stream's fetch slots frees
	// (or the stream dies — the false return, on which the fetcher just
	// abandons the scan like any other teardown path). Parked acquirers are
	// served FIFO by the channel.
	acquireScanSlot := func() bool {
		select {
		case scanSlots <- struct{}{}:
			return true
		case <-ctx.Done():
		case <-s.ctx.Done():
		}
		return false
	}
	releaseScanSlot := func() { <-scanSlots }

	// reserveCatchUpBytes parks until n payload bytes fit the catch-up budget,
	// returning false only when the stream dies first. The CAS-from-zero arm
	// admits exactly one turn into an empty lane whatever its size, so an
	// over-budget turn replays alone rather than deadlocking.
	reserveCatchUpBytes := func(n int) bool {
		for {
			cur := catchUpBytes.Load()
			if cur == 0 || cur+int64(n) <= int64(byteBudget) {
				if catchUpBytes.CompareAndSwap(cur, cur+int64(n)) {
					if s.scanBytesObserved != nil {
						s.scanBytesObserved(cur + int64(n))
					}
					return true
				}
				continue
			}
			select {
			case <-catchUpBytesFreed:
			case <-ctx.Done():
				return false
			case <-s.ctx.Done():
				return false
			}
		}
	}
	// freeCatchUpBytes releases a batch's reservation once the writer has
	// consumed it (sent, filtered, or dropped with its wave) and wakes one
	// parked fetcher to re-check the budget. Fetchers that parked without a
	// wakeup slot re-check on the next free — whenever the lane is over
	// budget, reserved batches exist for the writer to consume, so frees (and
	// wakeups) keep coming.
	freeCatchUpBytes := func(n int) {
		if n == 0 {
			return
		}
		catchUpBytes.Add(-int64(n))
		select {
		case catchUpBytesFreed <- struct{}{}:
		default:
		}
	}

	// catchUpGroups replays one wave's group topics in rotating batches (see
	// catchUpBatchTopics). Topics whose cursor already sits at the ceiling are
	// pruned without a query — a topic can only contribute rows in
	// (cursor, ceiling], and reconnect waves are overwhelmingly fully current,
	// so most waves complete right there. Each turn queries one batch for up
	// to the per-topic limit of rows per topic: a topic that fills its limit
	// may have more, so it advances its own cursor and rotates to the back of
	// the queue; every other topic is fully replayed up to the ceiling and
	// retires. Every fetched row is delivered — turns never discard and
	// re-fetch work. Anything newer than the ceiling reaches the client
	// through the gated live path and is folded into the wave when it
	// completes. It owns no shared state and never sends to the stream; it
	// just queries and hands results over.
	catchUpGroups := func(adds []mlsstore.GroupCatchup, wave int) {
		// The ceiling pin assumes v3's id-visibility-order invariant: ids become visible to
		// readers in id order (the live poller in worker.go advances from raw rows on the same
		// assumption — a row committing out of order behind a reader is undeliverable
		// stream-wide, a pre-existing v3 property). This first snapshot runs
		// slot-free — it is one index-tail probe — so the dominant
		// fully-current reconnect wave prunes to nothing and completes without
		// queuing behind a deep scan: parked with no replay to do, such a wave
		// would only hold its topics gated while live traffic accumulates
		// against maxPendingBytes.
		ceiling, err := s.readOnlyStore.Queries().GetLatestGroupMessageID(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error("wave-scan ceiling (group)", zap.Error(err))
				forward(catchUpBatch{err: err})
			}
			return
		}
		queue := make([]mlsstore.GroupCatchup, 0, len(adds))
		for _, a := range adds {
			if a.IdCursor < uint64(ceiling) {
				queue = append(queue, a)
			}
		}
		if len(queue) > 0 {
			// Only a wave with rows to replay competes for a scan slot; the
			// slot is held until the wave's done marker is in the channel.
			if !acquireScanSlot() {
				return
			}
			defer releaseScanSlot()
			// Re-snapshot under the slot: the wait may have been long, and a
			// fresher ceiling moves rows from the gated pending fold into the
			// scan (never the reverse — the ceiling only grows, and the queue
			// needs no re-prune: every queued cursor sits below the old
			// ceiling, hence below the new). A refresh failure is benign; the
			// first snapshot stays a valid pin.
			c, cerr := s.readOnlyStore.Queries().GetLatestGroupMessageID(ctx)
			if cerr == nil && c > ceiling {
				ceiling = c
			}
		}
		// Guard against non-positive tunables: a zero batch would spin the
		// loop forever without consuming the queue, and a zero limit would
		// make every topic look drained.
		batchTopics := s.scanBatchTopics
		if batchTopics < 1 {
			batchTopics = 1
		}
		limit := s.scanTopicPageLimit
		if limit < 1 {
			limit = 1
		}
		for len(queue) > 0 {
			select {
			case <-ctx.Done():
				return
			case <-s.ctx.Done():
				return
			default:
			}
			batch := queue
			if len(batch) > batchTopics {
				batch = queue[:batchTopics]
			}
			queue = queue[len(batch):]
			msgs, err := s.readOnlyStore.QueryGroupMessagesWaveScan(
				ctx,
				batch,
				0,
				uint64(ceiling),
				limit,
			)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Error("wave-scan catch-up (group)", zap.Error(err))
					forward(catchUpBatch{err: err})
				}
				return
			}
			// Per-topic scan positions, captured BEFORE the writer takes
			// ownership of msgs (it filters the slice in place once handed
			// over). msgs ascend, so each topic's entry ends at its max id.
			counts := make(map[string]int, len(batch))
			lastIDs := make(map[string]uint64, len(batch))
			for _, m := range msgs {
				v1 := m.GetV1()
				k := string(v1.GetGroupId())
				counts[k]++
				lastIDs[k] = v1.GetId()
			}
			if len(msgs) > 0 {
				turnBytes := 0
				for _, m := range msgs {
					turnBytes += len(m.GetV1().GetData())
				}
				if !reserveCatchUpBytes(turnBytes) {
					return
				}
				forward(catchUpBatch{groupMsgs: msgs, wave: wave, bytes: turnBytes})
			}
			// A topic that filled its per-topic limit may have more rows:
			// advance its own cursor and rotate it to the back. Every other
			// topic is fully replayed up to the ceiling and retires.
			for _, b := range batch {
				k := string(b.GroupID)
				if counts[k] < int(limit) {
					continue
				}
				if b.IdCursor < lastIDs[k] {
					b.IdCursor = lastIDs[k]
				}
				queue = append(queue, b)
			}
		}
		forward(catchUpBatch{wave: wave, done: true})
	}

	// catchUpWelcomes is the welcome-topic analogue of catchUpGroups. Topic
	// truncation and rotation cursors come from the store's RAW per-topic
	// progress, not the parsed slice: the store skips rows with an unknown
	// message_type but they still consumed their topic's LIMIT slots, so
	// paging by parsed rows would silently truncate the replay at the first
	// skipped row.
	catchUpWelcomes := func(adds []mlsstore.WelcomeCatchup, wave int) {
		ceiling, err := s.readOnlyStore.Queries().GetLatestWelcomeMessageID(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error("wave-scan ceiling (welcome)", zap.Error(err))
				forward(catchUpBatch{err: err})
			}
			return
		}
		queue := make([]mlsstore.WelcomeCatchup, 0, len(adds))
		for _, a := range adds {
			if a.IdCursor < uint64(ceiling) {
				queue = append(queue, a)
			}
		}
		if len(queue) > 0 {
			// Same slot discipline and ceiling re-snapshot as catchUpGroups.
			if !acquireScanSlot() {
				return
			}
			defer releaseScanSlot()
			c, cerr := s.readOnlyStore.Queries().GetLatestWelcomeMessageID(ctx)
			if cerr == nil && c > ceiling {
				ceiling = c
			}
		}
		batchTopics := s.scanBatchTopics
		if batchTopics < 1 {
			batchTopics = 1
		}
		limit := s.scanTopicPageLimit
		if limit < 1 {
			limit = 1
		}
		for len(queue) > 0 {
			select {
			case <-ctx.Done():
				return
			case <-s.ctx.Done():
				return
			default:
			}
			batch := queue
			if len(batch) > batchTopics {
				batch = queue[:batchTopics]
			}
			queue = queue[len(batch):]
			msgs, progress, err := s.readOnlyStore.QueryWelcomeMessagesWaveScan(
				ctx,
				batch,
				0,
				uint64(ceiling),
				limit,
			)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Error("wave-scan catch-up (welcome)", zap.Error(err))
					forward(catchUpBatch{err: err})
				}
				return
			}
			if len(msgs) > 0 {
				turnBytes := 0
				for _, m := range msgs {
					turnBytes += len(welcomeData(m))
				}
				if !reserveCatchUpBytes(turnBytes) {
					return
				}
				forward(catchUpBatch{welcomeMsgs: msgs, wave: wave, bytes: turnBytes})
			}
			for _, b := range batch {
				p := progress[string(b.InstallationKey)]
				if p.RawCount < int(limit) {
					continue
				}
				if b.IdCursor < p.LastRawID {
					b.IdCursor = p.LastRawID
				}
				queue = append(queue, b)
			}
		}
		forward(catchUpBatch{wave: wave, done: true})
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
				// forward-version client is not left waiting on a response (XIP-83 req 10).
				return status.Errorf(codes.InvalidArgument, "unrecognized SubscribeRequest version")
			}
			switch {
			case v1.GetMutate() != nil:
				m := v1.GetMutate()
				// A wave's replay frames are stamped with its mutate_id, and 0 is the live
				// tag, so a wave of adds cannot ride on 0 (XIP-83 server requirement 3).
				// Validated before any state changes so the rejected frame is atomic.
				if len(m.GetAdds()) > 0 && m.GetMutateId() == 0 {
					return status.Errorf(
						codes.InvalidArgument,
						"a Mutate with adds requires a nonzero mutate_id",
					)
				}
				// Each add becomes a floor entry the wave holds for its lifetime
				// (see maxMutateAdds). Checked on the raw adds — stateless, before
				// any state changes — so the rejected frame is atomic.
				if len(m.GetAdds()) > s.maxMutateAdds {
					return status.Errorf(
						codes.ResourceExhausted,
						"adds-per-Mutate limit %d exceeded; split the adds across Mutates",
						s.maxMutateAdds,
					)
				}
				// The frame tag and the CatchupComplete echo are the only keys correlating
				// frames to mutations, so a mutate_id reused while its wave is still in
				// flight would make the two waves' replay and completions indistinguishable
				// (XIP-83 server requirement 3). This applies to ANY Mutate — even a
				// removes-only one, whose immediate CatchupComplete would be ambiguous with
				// the in-flight wave's — and is checked before the removes' side effects.
				// In-flight ids are always nonzero (waves start only from adds-bearing
				// Mutates, rejected above on 0), so a mutate_id of 0 never collides. Reuse
				// after a wave's CatchupComplete stays legal.
				if m.GetMutateId() != 0 {
					for _, w := range waves {
						if w.mutateID == m.GetMutateId() {
							return status.Errorf(
								codes.InvalidArgument,
								"mutate_id %d is already in flight on this stream",
								m.GetMutateId(),
							)
						}
					}
				}
				// history_only adds never touch the dispatcher: no live registration,
				// no gate, no pending buffer — a pure batched read with markers.
				historyOnly := m.GetHistoryOnly()
				// Parse and kind-validate every add up front — pure parsing, no state
				// decisions — so a malformed add fails the stream BEFORE any remove's side
				// effects (dropTopic, including a freed wave's CatchupComplete) take hold.
				// The state-dependent no-op/reset decisions stay below the removes: a
				// same-Mutate remove+re-add must see the removed state.
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

				var groupAdds []mlsstore.GroupCatchup
				var groupTopics []string
				var groupWire [][]byte
				var welcomeAdds []mlsstore.WelcomeCatchup
				var welcomeTopics []string
				var welcomeWire [][]byte
				newlyLive := 0
				for _, t := range addOrder {
					a := addByTopic[t]
					// A topic with an in-flight history_only catch-up has an active wave but no
					// live registration (history_only never sets subscribed/catchingUp). A second
					// overlapping add for it would start a competing wave and reset the high-water
					// floor, re-delivering history the first wave already sent — reject. (This
					// covers the common no-remove overlap; a remove+re-add of such a topic is an
					// even more unusual sequence that dropTopic narrows but does not fully close.)
					if _, hasWave := topicWave[t]; hasWave {
						if _, live := subscribed[t]; !live && !catchingUp[t] {
							return status.Errorf(
								codes.InvalidArgument,
								"add targets a topic with an in-flight history_only catch-up",
							)
						}
					}
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
						// Reject rather than guess (XIP-83 req 11 stance: fail contradictory input).
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
				// CatchupComplete (echoing mutate_id) is emitted once its scans drain and
				// its gates open. Every other mutate — removes-only, empty, or adds that
				// were all no-ops — is acknowledged with an immediate CatchupComplete, so
				// every Mutate is answered by exactly one CatchupComplete echoing its
				// mutate_id.
				adds := len(groupAdds) + len(welcomeAdds)
				switch {
				case adds > 0:
					wave := nextWave
					nextWave++
					scans := 0
					if len(groupAdds) > 0 {
						scans++
					}
					if len(welcomeAdds) > 0 {
						scans++
					}
					waves[wave] = waveState{
						mutateID:  m.GetMutateId(),
						scansLeft: scans,
						owned:     adds,
						topics:    append(append([]string{}, groupTopics...), welcomeTopics...),
						wire:      append(append([][]byte{}, groupWire...), welcomeWire...),
					}
					for _, t := range groupTopics {
						topicWave[t] = wave
					}
					for _, t := range welcomeTopics {
						topicWave[t] = wave
					}
					if len(groupAdds) > 0 {
						go catchUpGroups(groupAdds, wave)
					}
					if len(welcomeAdds) > 0 {
						go catchUpWelcomes(welcomeAdds, wave)
					}
				default:
					if err := send(catchupCompleteFrame(m.GetMutateId())); err != nil {
						return err
					}
				}

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
			// Taking the batch consumes it — sent below, filtered, or dropped
			// with a dead wave — so its byte reservation is freed on every one
			// of those paths, here, before any of them branch. (The err return
			// above skips this safely: error batches always carry zero bytes.)
			freeCatchUpBytes(b.bytes)
			w, waveActive := waves[b.wave]
			if !waveActive {
				continue // the wave completed early (all topics removed); drop the straggler
			}
			// A topic can be removed (or reset) while its wave's scan page is in flight.
			// wanted reports whether this batch's wave still owns the topic, so a removed
			// topic never gets stale history. Dropping rows preserves the scan's order.
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
				history = buildGroupFrames(kept, w.mutateID)
			} else if len(b.welcomeMsgs) > 0 {
				kept := b.welcomeMsgs[:0]
				for _, m := range b.welcomeMsgs {
					if wanted(topic.BuildMLSV1WelcomeTopic(welcomeInstallationKey(m))) {
						kept = append(kept, m)
					}
				}
				history = buildWelcomeFrames(kept, w.mutateID)
			}
			if err := send(history); err != nil {
				return err
			}
			if b.done {
				w.scansLeft--
				if w.scansLeft > 0 {
					waves[b.wave] = w
				} else {
					// The wave's last scan just drained: flush its gated live buffer (still
					// the wave's replay), announce TopicsLive, emit its CatchupComplete, and
					// open the gates — in that order, so the seam holds.
					if err := completeWave(b.wave, w); err != nil {
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
			// Live tail: tagged 0. The dispatcher delivers each kind in ascending id
			// order and the writer sends in arrival order, so the live lane stays
			// totally ordered per kind (XIP-83 server requirement 4).
			//
			// Coalesce any messages already queued behind this one — without blocking, so
			// no added latency — into as few frames as possible. buildWelcome/GroupFrames
			// pack per kind up to maxFrameBytes, so a burst of small live messages that
			// would otherwise stream as one tiny frame each (pressuring the h2 inbound
			// data-frame guard on clients streaming many small welcomes/messages) goes out
			// in a handful of full frames. Messages for still-catching-up topics keep
			// buffering to the gated lane, exactly as before.
			var liveWelcomes []*mlsv1.WelcomeMessage
			var liveGroups []*mlsv1.GroupMessage
			route := func(e *v1proto.Envelope) error {
				t := e.ContentTopic
				if catchingUp[t] {
					pending[t] = append(pending[t], e)
					pendingBytes += len(e.Message)
					if pendingBytes > s.maxPendingBytes {
						return status.Errorf(
							codes.ResourceExhausted,
							"catch-up buffer exceeded; reconnect from cursor",
						)
					}
					return nil
				}
				if topic.IsMLSV1Welcome(t) {
					if m, err := getWelcomeMessageFromEnvelope(e); err == nil {
						liveWelcomes = append(liveWelcomes, m)
					} else {
						log.Error("error parsing welcome message", zap.Error(err))
					}
				} else {
					if m, err := getGroupMessageFromEnvelope(e); err == nil {
						liveGroups = append(liveGroups, m)
					} else {
						log.Error("error parsing group message", zap.Error(err))
					}
				}
				return nil
			}
			if err := route(env); err != nil {
				return err
			}
		drainLive:
			for i := 0; i < liveCoalesceMax; i++ {
				select {
				case next, ok := <-sub.MessagesCh:
					if !ok {
						return status.Errorf(codes.Aborted, "subscription closed: consumer too slow")
					}
					if err := route(next); err != nil {
						return err
					}
				default:
					break drainLive // channel momentarily empty: flush what we have
				}
			}
			if err := send(buildWelcomeFrames(liveWelcomes, 0), buildGroupFrames(liveGroups, 0)); err != nil {
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

// stopTimer stops t and drains a pending fire if there is one, so a later Reset cannot observe a
// stale value. Safe on an already-stopped or already-fired-and-drained timer.
func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}
