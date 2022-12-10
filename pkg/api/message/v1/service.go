package api

import (
	"context"
	"strings"
	"sync"

	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/crdt"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	contentTopicAll = "[ALL]"
)

type Service struct {
	messagev1.UnimplementedMessageApiServer

	// Configured as constructor options.
	log  *zap.Logger
	crdt *crdt.Node

	// Configured internally.
	ctx        context.Context
	ctxCancel  func()
	wg         sync.WaitGroup
	dispatcher *dispatcher
}

func NewService(log *zap.Logger, crdt *crdt.Node) (s *Service, err error) {
	s = &Service{
		log:  log.Named("message/v1"),
		crdt: crdt,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	s.dispatcher = newDispatcher()
	tracing.GoPanicWrap(s.ctx, &s.wg, "broadcast", func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case env := <-s.crdt.EnvC:
				s.dispatcher.Submit(env.ContentTopic, env)
			}
		}
	})

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

func (s *Service) Publish(ctx context.Context, req *messagev1.PublishRequest) (*messagev1.PublishResponse, error) {
	for _, env := range req.Envelopes {
		log := s.log.Named("publish").With(zap.String("content_topic", env.ContentTopic))
		log.Info("received message")

		err := s.crdt.Publish(ctx, env)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		metrics.EmitPublishedEnvelope(ctx, log, env)
	}
	return &messagev1.PublishResponse{}, nil
}

func (s *Service) Subscribe(req *messagev1.SubscribeRequest, stream messagev1.MessageApi_SubscribeServer) error {
	log := s.log.Named("subscribe").With(zap.Strings("content_topics", req.ContentTopics))
	log.Info("started")
	defer log.Info("stopped")

	subC := s.dispatcher.Register(nil, req.ContentTopics...)
	defer s.dispatcher.Unregister(subC)

	for {
		select {
		case <-stream.Context().Done():
			log.Info("stream closed")
			return nil
		case <-s.ctx.Done():
			log.Info("service closed")
			return nil
		case obj := <-subC:
			env, ok := obj.(*messagev1.Envelope)
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

func (s *Service) SubscribeAll(req *messagev1.SubscribeAllRequest, stream messagev1.MessageApi_SubscribeAllServer) error {
	log := s.log.Named("subscribeAll")
	log.Info("started")
	defer log.Info("stopped")

	return s.Subscribe(&messagev1.SubscribeRequest{
		ContentTopics: []string{contentTopicAll},
	}, stream)
}

func (s *Service) Query(ctx context.Context, req *messagev1.QueryRequest) (*messagev1.QueryResponse, error) {
	log := s.log.Named("query").With(zap.Strings("content_topics", req.ContentTopics))
	log.Info("received request")

	envs, pagingInfo, err := s.crdt.Query(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &messagev1.QueryResponse{
		Envelopes:  envs,
		PagingInfo: pagingInfo,
	}, nil
}

func isValidTopic(topic string) bool {
	return strings.HasPrefix(topic, "/xmtp/0/")
}
