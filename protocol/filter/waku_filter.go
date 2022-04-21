package filter

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	libp2pProtocol "github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-msgio/protoio"
	"github.com/status-im/go-waku/waku/v2/metrics"
	"github.com/status-im/go-waku/waku/v2/protocol"
	goWakuFilter "github.com/status-im/go-waku/waku/v2/protocol/filter"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var (
	ErrNoPeersAvailable = errors.New("no suitable remote peers")
)

const (
	CHANNEL_SIZE = 32
)

type (
	WakuFilter struct {
		ctx        context.Context
		h          host.Host
		isFullNode bool
		MsgC       chan *protocol.Envelope
		wg         *sync.WaitGroup
		log        *zap.SugaredLogger

		filters     *goWakuFilter.FilterMap
		subscribers *Subscribers
	}
)

// NOTE This is just a start, the design of this protocol isn't done yet. It
// should be direct payload exchange (a la req-resp), not be coupled with the
// relay protocol.
const FilterID_v10beta1 = libp2pProtocol.ID("/xmtp/filter/1.0.0-beta1")

func NewWakuFilter(ctx context.Context, host host.Host, isFullNode bool, log *zap.SugaredLogger, opts ...goWakuFilter.Option) (*WakuFilter, error) {
	wf := new(WakuFilter)
	wf.log = log.Named("filter")

	ctx, err := tag.New(ctx, tag.Insert(metrics.KeyType, "filter"))
	if err != nil {
		wf.log.Error(err)
		return nil, errors.New("could not start waku filter")
	}

	params := new(goWakuFilter.FilterParameters)
	optList := goWakuFilter.DefaultOptions()
	optList = append(optList, opts...)
	for _, opt := range optList {
		opt(params)
	}

	wf.ctx = ctx
	wf.wg = &sync.WaitGroup{}
	wf.MsgC = make(chan *protocol.Envelope, 1024)
	wf.h = host
	wf.isFullNode = isFullNode
	wf.filters = goWakuFilter.NewFilterMap()
	wf.subscribers = NewSubscribers(params.Timeout)

	wf.h.SetStreamHandlerMatch(FilterID_v10beta1, protocol.PrefixTextMatch(string(FilterID_v10beta1)), wf.onRequest)
	wf.log.Info("Registered stream for protocol", FilterID_v10beta1)
	wf.wg.Add(1)
	go wf.FilterListener()

	if wf.isFullNode {
		wf.log.Info("Filter protocol started")
	} else {
		wf.log.Info("Filter protocol started (only client mode)")
	}

	return wf, nil
}

func (wf *WakuFilter) onRequest(s network.Stream) {
	filterRPCRequest := &pb.FilterRPC{}

	reader := protoio.NewDelimitedReader(s, math.MaxInt32)

	err := reader.ReadMsg(filterRPCRequest)
	if err != nil {
		wf.log.Error("error reading request", err)
		return
	}

	wf.log.Info(fmt.Sprintf("%s: received request from %s", s.Conn().LocalPeer(), s.Conn().RemotePeer()))

	// We're on a light node.
	// This is a message push coming from a full node.
	if filterRPCRequest.Request != nil && wf.isFullNode {
		// We're on a full node.
		// This is a filter request coming from a light node.
		if filterRPCRequest.Request.Subscribe {
			subscriber := Subscriber{peer: s.Conn().RemotePeer(), requestId: filterRPCRequest.RequestId, filter: *filterRPCRequest.Request, ch: make(chan *pb.WakuMessage, CHANNEL_SIZE)}
			len := wf.subscribers.Append(subscriber)

			wf.log.Info("filter full node, add a filter subscriber: ", subscriber.peer)
			stats.Record(wf.ctx, metrics.FilterSubscriptions.M(int64(len)))

			go wf.streamMessagesToClient(s, subscriber)
		} else {
			// Technically I still support "unsubscribe" requests on the server for compatibility with the Spec and the possibility of needing it later (it was already there)
			// In practice, the only way the client can currently unsubscribe is to close the connection
			peerId := s.Conn().RemotePeer()
			wf.subscribers.RemoveContentFilters(peerId, filterRPCRequest.Request.ContentFilters)

			wf.log.Info("filter full node, remove a filter subscriber: ", peerId.Pretty())
			stats.Record(wf.ctx, metrics.FilterSubscriptions.M(int64(wf.subscribers.Length())))
		}
	} else {
		wf.log.Error("can't serve request")
		return
	}
}

func (wf *WakuFilter) streamMessagesToClient(stream network.Stream, subscriber Subscriber) {
	writer := protoio.NewDelimitedWriter(stream)
	// As soon as connection has an error close the stream on our end and remove the filter
	defer stream.Close()
	defer wf.subscribers.RemoveContentFilters(subscriber.peer, subscriber.filter.ContentFilters)
	for {
		select {
		case <-wf.ctx.Done():
			return
		case msg := <-subscriber.ch:
			wf.log.Info("Sending message to remote peer", subscriber.peer.String())

			messagePush := &pb.FilterRPC{
				RequestId: subscriber.requestId,
				Request:   nil,
				Push: &pb.MessagePush{
					Messages: []*pb.WakuMessage{msg},
				},
			}
			// Pretty sure just breaking on the first write error is going to be too brittle for real world usage.
			// Need to find a way to make this the right amount of resilient to errors
			if err := writer.WriteMsg(messagePush); err != nil {
				wf.log.Error("Error writing to stream. Breaking", err)
				return
			}
			wf.subscribers.FlagAsSuccess(subscriber.peer)
		}
	}
}

func (wf *WakuFilter) pushMessage(subscriber Subscriber, msg *pb.WakuMessage) {
	select {
	case subscriber.ch <- msg:
		wf.log.Info("Sent message to channel")
	default:
		wf.log.Warnf("Trying to push to full channel for subscriber %s", subscriber.peer)
	}
}

func (wf *WakuFilter) FilterListener() {
	defer wf.wg.Done()

	// This function is invoked for each message received
	// on the full node in context of Waku2-Filter
	handle := func(envelope *protocol.Envelope) error { // async
		msg := envelope.Message()
		topic := envelope.PubsubTopic()
		// Each subscriber is a light node that earlier on invoked
		// a FilterRequest on this node
		for subscriber := range wf.subscribers.Items() {
			if subscriber.filter.Topic != "" && subscriber.filter.Topic != topic {
				wf.log.Info("Subscriber's filter pubsubTopic does not match message topic", subscriber.filter.Topic, topic)
				continue
			}

			for _, filter := range subscriber.filter.ContentFilters {
				if msg.ContentTopic == filter.ContentTopic {
					// Do a message push to light node
					wf.log.Info("pushing messages to light node: ", subscriber.peer)
					wf.pushMessage(subscriber, msg)
				}
			}
		}

		return nil
	}

	for m := range wf.MsgC {
		if err := handle(m); err != nil {
			wf.log.Error("failed to handle message", err)
		}
	}
}

func (wf *WakuFilter) Unsubscribe(ctx context.Context, contentFilter goWakuFilter.ContentFilter, peer peer.ID) error {
	// We connect first so dns4 addresses are resolved (NewStream does not do it)
	err := wf.h.Connect(ctx, wf.h.Peerstore().PeerInfo(peer))
	if err != nil {
		return err
	}

	conn, err := wf.h.NewStream(ctx, peer, FilterID_v10beta1)
	if err != nil {
		return err
	}

	defer conn.Close()

	// This is the only successful path to subscription
	id := protocol.GenerateRequestId()

	var contentFilters []*pb.FilterRequest_ContentFilter
	for _, ct := range contentFilter.ContentTopics {
		contentFilters = append(contentFilters, &pb.FilterRequest_ContentFilter{ContentTopic: ct})
	}

	request := pb.FilterRequest{
		Subscribe:      false,
		Topic:          contentFilter.Topic,
		ContentFilters: contentFilters,
	}

	writer := protoio.NewDelimitedWriter(conn)
	filterRPC := &pb.FilterRPC{RequestId: hex.EncodeToString(id), Request: &request}
	err = writer.WriteMsg(filterRPC)
	if err != nil {
		return err
	}

	return nil
}

func (wf *WakuFilter) Stop() {
	close(wf.MsgC)

	wf.h.RemoveStreamHandler(FilterID_v10beta1)
	wf.filters.RemoveAll()
	wf.wg.Wait()
}

func (wf *WakuFilter) Subscribe(ctx context.Context, f goWakuFilter.ContentFilter, opts ...goWakuFilter.FilterSubscribeOption) (filterID string, theFilter goWakuFilter.Filter, err error) {
	params := new(goWakuFilter.FilterSubscribeParameters)
	params.Log = wf.log
	params.Host = wf.h

	optList := goWakuFilter.DefaultSubscribtionOptions()
	optList = append(optList, opts...)
	for _, opt := range optList {
		opt(params)
	}

	if params.SelectedPeer == "" {
		err = ErrNoPeersAvailable
		return
	}

	var contentFilters []*pb.FilterRequest_ContentFilter
	for _, ct := range f.ContentTopics {
		contentFilters = append(contentFilters, &pb.FilterRequest_ContentFilter{ContentTopic: ct})
	}

	// We connect first so dns4 addresses are resolved (NewStream does not do it)
	err = wf.h.Connect(ctx, wf.h.Peerstore().PeerInfo(params.SelectedPeer))
	if err != nil {
		return
	}

	request := pb.FilterRequest{
		Subscribe:      true,
		Topic:          f.Topic,
		ContentFilters: contentFilters,
	}

	var conn network.Stream
	conn, err = wf.h.NewStream(ctx, params.SelectedPeer, FilterID_v10beta1)
	if err != nil {
		return
	}

	// This is the only successful path to subscription
	requestID := hex.EncodeToString(protocol.GenerateRequestId())

	writer := protoio.NewDelimitedWriter(conn)
	filterRPC := &pb.FilterRPC{RequestId: requestID, Request: &request}
	wf.log.Info("sending filterRPC: ", filterRPC)
	err = writer.WriteMsg(filterRPC)
	if err != nil {
		wf.log.Error("failed to write message", err)
		return
	}

	theFilter = goWakuFilter.Filter{
		PeerID:         params.SelectedPeer,
		Topic:          f.Topic,
		ContentFilters: f.ContentTopics,
		Chan:           make(chan *protocol.Envelope, 1024), // To avoid blocking
	}

	go func() {
		reader := protoio.NewDelimitedReader(conn, math.MaxInt32)
		for {
			select {
			case <-ctx.Done():
				wf.log.Info("Channel is closed")
				conn.Close()
				close(theFilter.Chan)
				return
			default:
				filterRPCRequest := &pb.FilterRPC{}
				// If this is stuck reading forever, the done channel will never get read from
				if err := reader.ReadMsg(filterRPCRequest); err != nil {
					wf.log.Error("Error reading message", err)
					conn.Close()
					close(theFilter.Chan)
					return
				}

				if filterRPCRequest.Push != nil {
					for _, msg := range filterRPCRequest.Push.Messages {
						theFilter.Chan <- protocol.NewEnvelope(msg, f.Topic)
					}
				} else {
					wf.log.Warnf("Push is nil in message from peer %s", params.SelectedPeer)
				}
			}

		}

	}()

	return
}

// UnsubscribeFilterByID removes a subscription to a filter node completely
// using the filterID returned when the subscription was created
func (wf *WakuFilter) UnsubscribeFilterByID(ctx context.Context, filterID string) error {

	var f goWakuFilter.Filter
	var ok bool

	if f, ok = wf.filters.Get(filterID); !ok {
		return errors.New("filter not found")
	}

	cf := goWakuFilter.ContentFilter{
		Topic:         f.Topic,
		ContentTopics: f.ContentFilters,
	}

	err := wf.Unsubscribe(ctx, cf, f.PeerID)
	if err != nil {
		return err
	}

	wf.filters.Delete(filterID)

	return nil
}

// Unsubscribe filter removes content topics from a filter subscription. If all
// the contentTopics are removed the subscription is dropped completely
func (wf *WakuFilter) UnsubscribeFilter(ctx context.Context, cf goWakuFilter.ContentFilter) error {
	// Remove local filter
	var idsToRemove []string
	for filterMapItem := range wf.filters.Items() {
		f := filterMapItem.Value
		id := filterMapItem.Key

		if f.Topic != cf.Topic {
			continue
		}

		// Send message to full node in order to unsubscribe
		err := wf.Unsubscribe(ctx, cf, f.PeerID)
		if err != nil {
			return err
		}

		// Iterate filter entries to remove matching content topics
		// make sure we delete the content filter
		// if no more topics are left
		for _, cfToDelete := range cf.ContentTopics {
			for i, cf := range f.ContentFilters {
				if cf == cfToDelete {
					l := len(f.ContentFilters) - 1
					f.ContentFilters[l], f.ContentFilters[i] = f.ContentFilters[i], f.ContentFilters[l]
					f.ContentFilters = f.ContentFilters[:l]
					break
				}

			}
			if len(f.ContentFilters) == 0 {
				idsToRemove = append(idsToRemove, id)
			}
		}
	}

	for _, rId := range idsToRemove {
		wf.filters.Delete(rId)
	}

	return nil
}

func (wf *WakuFilter) MsgChannel() chan *protocol.Envelope {
	return wf.MsgC
}
