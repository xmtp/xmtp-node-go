package api

import (
	"context"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	proto "github.com/xmtp/proto/v3/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	validXMTPTopicPrefix = "/xmtp/0/"
	contentTopicAllXMTP  = validXMTPTopicPrefix + "*"
)

type Service struct {
	proto.UnimplementedMessageApiServer

	// Configured as constructor options.
	log  *zap.Logger
	nats *nats.Conn

	// Configured internally.
	ctx        context.Context
	ctxCancel  func()
	wg         sync.WaitGroup
	dispatcher *dispatcher
}

func NewService(nats *nats.Conn, logger *zap.Logger) (s *Service, err error) {
	s = &Service{
		nats: nats,
		log:  logger.Named("message/v1"),
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	s.dispatcher = newDispatcher()
	// s.relaySub, err = s.nats.Relay().Subscribe(s.ctx)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "subscribing to relay")
	// }
	// tracing.GoPanicWrap(s.ctx, &s.wg, "broadcast", func(ctx context.Context) {
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		case wakuEnv := <-s.relaySub.C:
	// 			if wakuEnv == nil {
	// 				continue
	// 			}
	// 			env := buildEnvelope(wakuEnv.Message())
	// 			s.dispatcher.Submit(env.ContentTopic, env)
	// 		}
	// 	}
	// })

	return s, nil
}

func (s *Service) Close() {
	s.log.Info("closing")

	if s.dispatcher != nil {
		s.dispatcher.Close()
	}

	if s.ctxCancel != nil {
		s.ctxCancel()
	}
	s.wg.Wait()
	s.log.Info("closed")
}

func (s *Service) Publish(ctx context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error) {
	for _, env := range req.Envelopes {
		log := s.log.Named("publish").With(zap.String("content_topic", env.ContentTopic))
		log.Debug("received message")

		// wakuMsg := &wakupb.WakuMessage{
		// 	ContentTopic: env.ContentTopic,
		// 	Timestamp:    toWakuTimestamp(env.TimestampNs),
		// 	Payload:      env.Message,
		// }
		// _, err := s.waku.Relay().Publish(ctx, wakuMsg)
		// if err != nil {
		// 	return nil, status.Errorf(codes.Internal, err.Error())
		// }
		metrics.EmitPublishedEnvelope(ctx, env)
	}
	return &proto.PublishResponse{}, nil
}

func (s *Service) Subscribe(req *proto.SubscribeRequest, stream proto.MessageApi_SubscribeServer) error {
	log := s.log.Named("subscribe").With(zap.Strings("content_topics", req.ContentTopics))
	log.Debug("started")
	defer log.Debug("stopped")

	subC := s.dispatcher.Register(nil, req.ContentTopics...)
	defer s.dispatcher.Unregister(subC)

	for {
		select {
		case <-stream.Context().Done():
			log.Debug("stream closed")
			return nil
		case <-s.ctx.Done():
			log.Info("service closed")
			return nil
		case obj := <-subC:
			env, ok := obj.(*proto.Envelope)
			if !ok {
				log.Warn("non-envelope received on subscription channel", zap.Any("object", obj))
				continue
			}
			err := stream.Send(env)
			if err != nil {
				log.Error("sending envelope to subscriber", zap.Error(err))
			}
		}
	}
}

func (s *Service) SubscribeAll(req *proto.SubscribeAllRequest, stream proto.MessageApi_SubscribeAllServer) error {
	log := s.log.Named("subscribeAll")
	log.Debug("started")
	defer log.Debug("stopped")

	return s.Subscribe(&proto.SubscribeRequest{
		ContentTopics: []string{contentTopicAllXMTP},
	}, stream)
}

func (s *Service) Query(ctx context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
	log := s.log.Named("query").With(zap.Strings("content_topics", req.ContentTopics))
	log.Debug("received request")

	// store, ok := s.waku.Store().(*store.XmtpStore)
	// if !ok {
	// 	return nil, status.Errorf(codes.Internal, "waku store not xmtp store")
	// }
	// res, err := store.FindMessages(buildWakuQuery(req))
	// if err != nil {
	// 	return nil, status.Errorf(codes.Internal, err.Error())
	// }

	// envs := make([]*proto.Envelope, 0, len(res.Messages))
	// for _, msg := range res.Messages {
	// 	envs = append(envs, buildEnvelope(msg))
	// }

	return &proto.QueryResponse{
		// Envelopes:  envs,
		// PagingInfo: buildPagingInfo(res.PagingInfo),
	}, nil
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

func isValidTopic(topic string) bool {
	return strings.HasPrefix(topic, validXMTPTopicPrefix)
}
