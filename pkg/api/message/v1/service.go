package api

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"strings"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	wakunode "github.com/waku-org/go-waku/waku/v2/node"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	wakurelay "github.com/waku-org/go-waku/waku/v2/protocol/relay"
	proto "github.com/xmtp/proto/v3/go/message_api/v1"
	apicontext "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/context"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"github.com/xmtp/xmtp-node-go/pkg/mlsstore"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
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
)

type Service struct {
	proto.UnimplementedMessageApiServer

	// Configured as constructor options.
	log      *zap.Logger
	waku     *wakunode.WakuNode
	store    *store.Store
	mlsStore mlsstore.MlsStore

	// Configured internally.
	ctx       context.Context
	ctxCancel func()
	wg        sync.WaitGroup
	relaySub  *wakurelay.Subscription

	ns *server.Server
	nc *nats.Conn
}

func NewService(node *wakunode.WakuNode, logger *zap.Logger, store *store.Store, mlsStore mlsstore.MlsStore) (s *Service, err error) {
	s = &Service{
		waku:     node,
		log:      logger.Named("message/v1"),
		store:    store,
		mlsStore: mlsStore,
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

	// Initialize waku relay subscription.
	s.relaySub, err = s.waku.Relay().Subscribe(s.ctx)
	if err != nil {
		return nil, errors.Wrap(err, "subscribing to relay")
	}
	tracing.GoPanicWrap(s.ctx, &s.wg, "broadcast", func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case wakuEnv := <-s.relaySub.Ch:
				if wakuEnv == nil {
					continue
				}
				env := buildEnvelope(wakuEnv.Message())

				envB, err := pb.Marshal(env)
				if err != nil {
					s.log.Error("marshalling envelope", zap.Error(err))
					continue
				}
				err = s.nc.Publish(buildNatsSubject(env.ContentTopic), envB)
				if err != nil {
					s.log.Error("publishing envelope to local nats", zap.Error(err))
					continue
				}
			}
		}
	})

	return s, nil
}

func (s *Service) Close() {
	s.log.Info("closing")
	if s.relaySub != nil {
		s.relaySub.Unsubscribe()
	}

	if s.ctxCancel != nil {
		s.ctxCancel()
	}

	if s.nc != nil {
		s.nc.Close()
	}
	if s.ns != nil {
		s.ns.Shutdown()
	}

	s.wg.Wait()
	s.log.Info("closed")
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

		_, err := s.waku.Relay().Publish(ctx, &wakupb.WakuMessage{
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

	metrics.EmitSubscribeTopicsLength(stream.Context(), log, len(req.ContentTopics))

	var streamLock sync.Mutex
	for _, topic := range req.ContentTopics {
		subject := topic
		if subject != natsWildcardTopic {
			subject = buildNatsSubject(topic)
		}
		sub, err := s.nc.Subscribe(subject, func(msg *nats.Msg) {
			var env proto.Envelope
			err := pb.Unmarshal(msg.Data, &env)
			if err != nil {
				log.Error("parsing envelope from bytes", zap.Error(err))
				return
			}
			if topic == natsWildcardTopic && !isValidSubscribeAllTopic(env.ContentTopic) {
				return
			}
			func() {
				streamLock.Lock()
				defer streamLock.Unlock()
				err := stream.Send(&env)
				if err != nil {
					log.Error("sending envelope to subscribe", zap.Error(err))
				}
			}()
		})
		if err != nil {
			log.Error("error subscribing", zap.Error(err))
			return err
		}
		defer func() {
			_ = sub.Unsubscribe()
		}()
	}

	select {
	case <-stream.Context().Done():
		log.Debug("stream closed")
		return nil
	case <-s.ctx.Done():
		log.Info("service closed")
		return nil
	}
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

	subs := map[string]*nats.Subscription{}
	defer func() {
		for _, sub := range subs {
			_ = sub.Unsubscribe()
		}
	}()
	var streamLock sync.Mutex
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
			log.Info("updating subscription", zap.Int("num_content_topics", len(req.ContentTopics)))

			metrics.EmitSubscribeTopicsLength(stream.Context(), log, len(req.ContentTopics))

			topics := map[string]bool{}
			for _, topic := range req.ContentTopics {
				topics[topic] = true

				// If topic not in existing subscriptions, then subscribe.
				if _, ok := subs[topic]; !ok {
					sub, err := s.nc.Subscribe(buildNatsSubject(topic), func(msg *nats.Msg) {
						var env proto.Envelope
						err := pb.Unmarshal(msg.Data, &env)
						if err != nil {
							log.Info("unmarshaling envelope", zap.Error(err))
							return
						}
						func() {
							streamLock.Lock()
							defer streamLock.Unlock()

							err = stream.Send(&env)
							if err != nil {
								log.Error("sending envelope to subscriber", zap.Error(err))
							}
						}()
					})
					if err != nil {
						return err
					}
					subs[topic] = sub
				}
			}

			// If subscription not in topic, then unsubscribe.
			for topic, sub := range subs {
				if topics[topic] {
					continue
				}
				_ = sub.Unsubscribe()
				delete(subs, topic)
			}
		}
	}
}

func (s *Service) SubscribeAll(req *proto.SubscribeAllRequest, stream proto.MessageApi_SubscribeAllServer) error {
	log := s.log.Named("subscribeAll")
	log.Debug("started")
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
		ri := apicontext.NewRequesterInfo(ctx)
		log.Info("query with multiple topics", ri.ZapFields()...)
	} else {
		log = log.With(zap.String("topic_type", topic.Category(req.ContentTopics[0])))
	}
	log = log.With(logging.QueryParameters(req))
	if req.StartTimeNs != 0 || req.EndTimeNs != 0 {
		ri := apicontext.NewRequesterInfo(ctx)
		log.Info("query with time filters", append(
			ri.ZapFields(),
			zap.Uint64("start_time", req.StartTimeNs),
			zap.Uint64("end_time", req.EndTimeNs),
		)...)
	}

	if req.PagingInfo != nil && req.PagingInfo.Cursor != nil {
		cursor := req.PagingInfo.Cursor.GetIndex()
		if cursor != nil && cursor.SenderTimeNs == 0 && cursor.Digest == nil {
			log.Info("query with partial cursor", zap.Int("cursor_timestamp", int(cursor.SenderTimeNs)), zap.Any("cursor_digest", cursor.Digest))
		}
	}

	var hasV3Topic, hasNonV3Topic bool
	for _, contentTopic := range req.ContentTopics {
		if topic.IsV3(contentTopic) {
			hasV3Topic = true
		} else {
			hasNonV3Topic = true
		}
	}
	if hasV3Topic && hasNonV3Topic {
		return nil, errors.New("cannot query v3 topics with non-v3 topics")
	}
	if hasV3Topic {
		if s.mlsStore == nil {
			return nil, errors.New("mls store not configured")
		}
		return s.mlsStore.QueryMessages(ctx, req)
	}

	return s.store.Query(req)
}

func (s *Service) BatchQuery(ctx context.Context, req *proto.BatchQueryRequest) (*proto.BatchQueryResponse, error) {
	log := s.log.Named("batchQuery")
	logFunc := log.Debug
	if len(req.Requests) > 10 {
		logFunc = log.Info
	}
	logFunc("large batch query", zap.Int("num_queries", len(req.Requests)))
	// NOTE: in our implementation, we implicitly limit batch size to 50 requests
	if len(req.Requests) > 50 {
		return nil, status.Errorf(codes.InvalidArgument, "cannot exceed 50 requests in single batch")
	}
	// Naive implementation, perform all sub query requests sequentially
	responses := make([]*proto.QueryResponse, 0)
	for _, query := range req.Requests {
		// We execute the query using the existing Query API

		var hasV3Topic, hasNonV3Topic bool
		for _, contentTopic := range query.ContentTopics {
			if topic.IsV3(contentTopic) {
				hasV3Topic = true
			} else {
				hasNonV3Topic = true
			}
		}
		if hasV3Topic && hasNonV3Topic {
			return nil, errors.New("cannot query v3 topics with non-v3 topics")
		}
		if hasV3Topic {
			if s.mlsStore == nil {
				return nil, errors.New("mls store not configured")
			}
			resp, err := s.mlsStore.QueryMessages(ctx, query)
			if err != nil {
				return nil, status.Errorf(codes.Internal, err.Error())
			}
			responses = append(responses, resp)
		}

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

func isValidSubscribeAllTopic(topic string) bool {
	return strings.HasPrefix(topic, validXMTPTopicPrefix)
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
