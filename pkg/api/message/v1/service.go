package api

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	apicontext "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/context"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	pb "google.golang.org/protobuf/proto"
)

const (
	validXMTPTopicPrefix = "/xmtp/0/"
	natsWildcardTopic    = "*"

	MaxContentTopicNameSize = 300

	// 1048576 - 300 - 62 = 1048214
	MaxMessageSize = pubsub.DefaultMaxMessageSize - MaxContentTopicNameSize - 62

	// maxQueriesPerBatch defines the maximum number of queries we can support per batch.
	maxQueriesPerBatch = 50

	// maxTopicsPerQueryRequest defines the maximum number of topics that can be queried in a single request.
	// the number is likely to be more than we want it to be, but would be a safe place to put it -
	// per Test_LargeQueryTesting, the request decoding already failing before it reaches th handler.
	// maxTopicsPerQueryRequest = 157733

	// maxTopicsPerBatchQueryRequest defines the maximum number of topics that can be queried in a batch query. This
	// limit is imposed in additional to the per-query limit maxTopicsPerRequest.
	// as a starting value, we've using the same value as above, since the entire request would be tossed
	// away before this is reached.
	// maxTopicsPerBatchQueryRequest = maxTopicsPerQueryRequest
)

type Service struct {
	proto.UnimplementedMessageApiServer

	// Configured as constructor options.
	log   *zap.Logger
	store *store.Store

	publishToWakuRelay func(context.Context, *wakupb.WakuMessage) error

	// Configured internally.
	ctx       context.Context
	ctxCancel func()
	wg        sync.WaitGroup

	ns *server.Server
	nc *nats.Conn

	subDispatcher *subscriptionDispatcher
}

func NewService(log *zap.Logger, store *store.Store, publishToWakuRelay func(context.Context, *wakupb.WakuMessage) error) (s *Service, err error) {
	s = &Service{
		log:                log.Named("message/v1"),
		store:              store,
		publishToWakuRelay: publishToWakuRelay,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	// Initialize nats for API subscribers.
	s.ns, err = server.NewServer(&server.Options{
		Port: server.RANDOM_PORT,
	})
	if err != nil {
		return nil, err
	}
	go s.ns.Start()
	if !s.ns.ReadyForConnections(4 * time.Second) {
		return nil, errors.New("nats not ready")
	}

	s.nc, err = nats.Connect(s.ns.ClientURL())
	if err != nil {
		return nil, err
	}

	s.subDispatcher, err = newSubscriptionDispatcher(s.nc, s.log)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Service) Close() {
	s.log.Info("closing")

	if s.ctxCancel != nil {
		s.ctxCancel()
	}
	s.subDispatcher.Shutdown()

	if s.nc != nil {
		s.nc.Close()
	}
	if s.ns != nil {
		s.ns.Shutdown()
	}

	s.wg.Wait()
	s.log.Info("closed")
}

func (s *Service) HandleIncomingWakuRelayMessage(msg *wakupb.WakuMessage) error {
	env := buildEnvelope(msg)

	envB, err := pb.Marshal(env)
	if err != nil {
		return err
	}

	err = s.nc.Publish(buildNatsSubject(env.ContentTopic), envB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Publish(ctx context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error) {
	for _, env := range req.Envelopes {
		log := s.log.Named("publish").With(zap.String("content_topic", env.ContentTopic))
		log.Debug("received message")

		if len(env.ContentTopic) > MaxContentTopicNameSize {
			return nil, status.Errorf(codes.InvalidArgument, "topic length too big")
		}

		if len(env.Message) > MaxMessageSize {
			return nil, status.Errorf(codes.InvalidArgument, "message too big")
		}

		if !topic.IsEphemeral(env.ContentTopic) {
			_, err := s.store.InsertMessage(env)
			if err != nil {
				return nil, status.Errorf(codes.Internal, err.Error())
			}
		}

		err := s.publishToWakuRelay(ctx, &wakupb.WakuMessage{
			ContentTopic: env.ContentTopic,
			Timestamp:    int64(env.TimestampNs),
			Payload:      env.Message,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		metrics.EmitPublishedEnvelope(ctx, log, env)
	}
	return &proto.PublishResponse{}, nil
}

func (s *Service) Subscribe(req *proto.SubscribeRequest, stream proto.MessageApi_SubscribeServer) error {
	log := s.log.Named("subscribe").With(zap.Strings("content_topics", req.ContentTopics))
	log.Debug("started")
	defer log.Debug("stopped")

	// Send a header (any header) to fix an issue with Tonic based GRPC clients.
	// See: https://github.com/xmtp/libxmtp/pull/58
	_ = stream.SendHeader(metadata.Pairs("subscribed", "true"))

	metrics.EmitSubscribeTopics(stream.Context(), log, len(req.ContentTopics))

	// create a topics map.
	topics := make(map[string]bool, len(req.ContentTopics))
	for _, topic := range req.ContentTopics {
		topics[topic] = true
	}
	sub := s.subDispatcher.Subscribe(topics)
	defer func() {
		if sub != nil {
			sub.Unsubscribe()
		}
		metrics.EmitUnsubscribeTopics(stream.Context(), log, len(req.ContentTopics))
	}()

	var streamLock sync.Mutex
	for exit := false; !exit; {
		select {
		case msg, open := <-sub.messagesCh:
			if open {
				func() {
					streamLock.Lock()
					defer streamLock.Unlock()
					err := stream.Send(msg)
					if err != nil {
						log.Error("sending envelope to subscribe", zap.Error(err))
					}
				}()
			} else {
				// channel got closed; likely due to backpressure of the sending channel.
				log.Info("stream closed due to backpressure")
				exit = true
			}
		case <-stream.Context().Done():
			log.Debug("stream closed")
			exit = true
		case <-s.ctx.Done():
			log.Info("service closed")
			exit = true
		}
	}
	return nil
}

func (s *Service) Subscribe2(stream proto.MessageApi_Subscribe2Server) error {
	log := s.log.Named("subscribe2")
	log.Debug("started")
	defer log.Debug("stopped")

	// Send a header (any header) to fix an issue with Tonic based GRPC clients.
	// See: https://github.com/xmtp/libxmtp/pull/58
	_ = stream.SendHeader(metadata.Pairs("subscribed", "true"))

	requestChannel := make(chan *proto.SubscribeRequest)
	go func() {
		for {
			select {
			case <-stream.Context().Done():
				return
			case <-s.ctx.Done():
				return
			default:
				req, err := stream.Recv()
				if err != nil {
					if e, ok := status.FromError(err); ok {
						if e.Code() != codes.Canceled {
							log.Error("reading subscription", zap.Error(err))
						}
					} else if err != io.EOF && err != context.Canceled {
						log.Error("reading subscription", zap.Error(err))
					}
					close(requestChannel)
					return
				}
				requestChannel <- req
			}
		}
	}()

	var streamLock sync.Mutex
	subscribedTopicCount := 0
	var currentSubscription *subscription
	defer func() {
		if currentSubscription != nil {
			currentSubscription.Unsubscribe()
			metrics.EmitUnsubscribeTopics(stream.Context(), log, subscribedTopicCount)
		}
	}()
	subscriptionChannel := make(chan *proto.Envelope, 1)
	for {
		select {
		case <-stream.Context().Done():
			log.Debug("stream closed")
			return nil
		case <-s.ctx.Done():
			log.Info("service closed")
			return nil
		case req := <-requestChannel:
			if req == nil {
				continue
			}

			// unsubscribe first.
			if currentSubscription != nil {
				currentSubscription.Unsubscribe()
				currentSubscription = nil
			}
			log.Info("updating subscription", zap.Int("num_content_topics", len(req.ContentTopics)))

			topics := map[string]bool{}
			for _, topic := range req.ContentTopics {
				topics[topic] = true
			}

			nextSubscription := s.subDispatcher.Subscribe(topics)
			if currentSubscription == nil {
				// on the first time, emit subscription
				metrics.EmitSubscribeTopics(stream.Context(), log, len(topics))
			} else {
				// otherwise, emit the change.
				metrics.EmitSubscriptionChange(stream.Context(), log, len(topics)-subscribedTopicCount)
			}
			subscribedTopicCount = len(topics)
			subscriptionChannel = nextSubscription.messagesCh
			currentSubscription = nextSubscription
		case msg, open := <-subscriptionChannel:
			if open {
				func() {
					streamLock.Lock()
					defer streamLock.Unlock()
					err := stream.Send(msg)
					if err != nil {
						log.Error("sending envelope to subscribe", zap.Error(err))
					}
				}()
			} else {
				// channel got closed; likely due to backpressure of the sending channel.
				log.Debug("stream closed due to backpressure")
				return nil
			}
		}
	}
}

func (s *Service) SubscribeAll(req *proto.SubscribeAllRequest, stream proto.MessageApi_SubscribeAllServer) error {
	log := s.log.Named("subscribeAll")
	log.Info("started")
	defer log.Debug("stopped")

	// Subscribe to all nats subjects via wildcard
	// https://docs.nats.io/nats-concepts/subjects#wildcards
	return s.Subscribe(&proto.SubscribeRequest{
		ContentTopics: []string{natsWildcardTopic},
	}, stream)
}

func (s *Service) Query(ctx context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
	log := s.log.Named("query").With(zap.Strings("content_topics", req.ContentTopics))
	log.Debug("received request")

	if len(req.ContentTopics) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "content topics required")
	}

	if len(req.ContentTopics) > 1 {
		// if len(req.ContentTopics) > maxTopicsPerQueryRequest {
		// 	return nil, status.Errorf(codes.InvalidArgument, "the number of content topics(%d) exceed the maximum topics per query request (%d)", len(req.ContentTopics), maxTopicsPerQueryRequest)
		// }
		ri := apicontext.NewRequesterInfo(ctx)
		log.Info("query with multiple topics", ri.ZapFields()...)
	} else {
		log = log.With(zap.String("topic_type", topic.Category(req.ContentTopics[0])))
	}
	log = log.With(logging.QueryParameters(req))

	if req.PagingInfo != nil && req.PagingInfo.Cursor != nil {
		cursor := req.PagingInfo.Cursor.GetIndex()
		if cursor != nil && cursor.SenderTimeNs == 0 && cursor.Digest == nil {
			log.Info("query with partial cursor", zap.Int("cursor_timestamp", int(cursor.SenderTimeNs)), zap.Any("cursor_digest", cursor.Digest))
		}
	}

	return s.store.Query(req)
}

func (s *Service) BatchQuery(ctx context.Context, req *proto.BatchQueryRequest) (*proto.BatchQueryResponse, error) {
	log := s.log.Named("batchQuery")
	log.Debug("batch query", zap.Int("num_queries", len(req.Requests)))

	// NOTE: in our implementation, we implicitly limit batch size to 50 requests (maxQueriesPerBatch = 50)
	if len(req.Requests) > maxQueriesPerBatch {
		return nil, status.Errorf(codes.InvalidArgument, "cannot exceed %d requests in single batch", maxQueriesPerBatch)
	}

	// calculate the total number of topics being requested in this batch request.
	// totalRequestedTopicsCount := 0
	// for _, query := range req.Requests {
	// 	totalRequestedTopicsCount += len(query.ContentTopics)
	// }

	// if totalRequestedTopicsCount == 0 {
	// 	return nil, status.Errorf(codes.InvalidArgument, "content topics required")
	// }

	// // are we still within limits ?
	// if totalRequestedTopicsCount > maxTopicsPerBatchQueryRequest {
	// 	return nil, status.Errorf(codes.InvalidArgument, "the total number of content topics(%d) exceed the maximum topics per batch query request(%d)", totalRequestedTopicsCount, maxTopicsPerBatchQueryRequest)
	// }

	// Naive implementation, perform all sub query requests sequentially
	responses := make([]*proto.QueryResponse, 0)
	for _, query := range req.Requests {
		// if len(query.ContentTopics) > maxTopicsPerQueryRequest {
		// 	return nil, status.Errorf(codes.InvalidArgument, "the number of content topics(%d) exceed the maximum topics per query request (%d)", len(query.ContentTopics), maxTopicsPerQueryRequest)
		// }
		// We execute the query using the existing Query API
		resp, err := s.Query(ctx, query)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		responses = append(responses, resp)
	}

	return &proto.BatchQueryResponse{
		Responses: responses,
	}, nil
}

func buildEnvelope(msg *wakupb.WakuMessage) *proto.Envelope {
	return &proto.Envelope{
		ContentTopic: msg.ContentTopic,
		TimestampNs:  fromWakuTimestamp(msg.Timestamp),
		Message:      msg.Payload,
	}
}

func fromWakuTimestamp(ts int64) uint64 {
	if ts < 0 {
		return 0
	}
	return uint64(ts)
}

func buildNatsSubject(topic string) string {
	hasher := fnv.New64a()
	hasher.Write([]byte(topic))
	return fmt.Sprintf("%x", hasher.Sum64())
}
