package api

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	wakupb "github.com/status-im/go-waku/waku/v2/protocol/pb"
	wakurelay "github.com/status-im/go-waku/waku/v2/protocol/relay"
	proto "github.com/xmtp/proto/v3/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	proto.UnimplementedMessageApiServer

	// Configured as constructor options.
	log  *zap.Logger
	waku *wakunode.WakuNode

	// Configured internally.
	ctx        context.Context
	ctxCancel  func()
	wg         sync.WaitGroup
	dispatcher *dispatcher
	relaySub   *wakurelay.Subscription
}

func NewService(node *wakunode.WakuNode, logger *zap.Logger) (s *Service, err error) {
	s = &Service{
		waku: node,
		log:  logger.Named("message/v1"),
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	s.dispatcher = newDispatcher()
	s.relaySub, err = s.waku.Relay().Subscribe(s.ctx)
	if err != nil {
		return nil, errors.Wrap(err, "subscribing to relay")
	}
	tracing.GoPanicWrap(s.ctx, &s.wg, "broadcast", func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case wakuEnv := <-s.relaySub.C:
				if wakuEnv == nil {
					continue
				}
				env := buildEnvelope(wakuEnv.Message())
				s.dispatcher.Submit(env.ContentTopic, env)
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
		log.Info("received message")

		wakuMsg := &wakupb.WakuMessage{
			ContentTopic: env.ContentTopic,
			Timestamp:    toWakuTimestamp(env.TimestampNs),
			Payload:      env.Message,
		}
		_, err := s.waku.Relay().Publish(ctx, wakuMsg)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		metrics.EmitPublishedEnvelope(ctx, env)
	}
	return &proto.PublishResponse{}, nil
}

func (s *Service) Subscribe(req *proto.SubscribeRequest, stream proto.MessageApi_SubscribeServer) error {
	log := s.log.Named("subscribe").With(zap.Strings("content_topics", req.ContentTopics))
	log.Info("started")
	defer log.Info("stopped")

	subC := s.dispatcher.Register(req.ContentTopics...)
	defer s.dispatcher.Unregister(subC, req.ContentTopics...)

	for {
		select {
		case <-stream.Context().Done():
			log.Info("stream closed")
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
	log.Info("started")
	defer log.Info("stopped")

	relaySub, err := s.waku.Relay().Subscribe(s.ctx)
	if err != nil {
		return err
	}

	defer relaySub.Unsubscribe()

	for {
		select {
		case <-stream.Context().Done():
			log.Info("stream closed")
			return nil
		case <-s.ctx.Done():
			log.Info("service closed")
			return nil
		case wakuEnv := <-relaySub.C:
			if wakuEnv == nil {
				continue
			}
			env := buildEnvelope(wakuEnv.Message())
			err := stream.Send(env)
			if err != nil {
				log.Error("sending envelope to subscriber", zap.Error(err))
			}
		}
	}
}

func (s *Service) Query(ctx context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
	log := s.log.Named("query").With(zap.Strings("content_topics", req.ContentTopics))
	log.Info("received request")

	store, ok := s.waku.Store().(*store.XmtpStore)
	if !ok {
		return nil, status.Errorf(codes.Internal, "waku store not xmtp store")
	}
	res, err := store.FindMessages(buildWakuQuery(req))
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	envs := make([]*proto.Envelope, 0, len(res.Messages))
	for _, msg := range res.Messages {
		envs = append(envs, buildEnvelope(msg))
	}

	return &proto.QueryResponse{
		Envelopes:  envs,
		PagingInfo: buildPagingInfo(res.PagingInfo),
	}, nil
}

func buildEnvelope(msg *wakupb.WakuMessage) *proto.Envelope {
	return &proto.Envelope{
		ContentTopic: msg.ContentTopic,
		TimestampNs:  fromWakuTimestamp(msg.Timestamp),
		Message:      msg.Payload,
	}
}

func buildWakuQuery(req *proto.QueryRequest) *wakupb.HistoryQuery {
	contentFilters := []*wakupb.ContentFilter{}
	for _, contentTopic := range req.ContentTopics {
		if contentTopic != "" {
			contentFilters = append(contentFilters, &wakupb.ContentFilter{
				ContentTopic: contentTopic,
			})
		}
	}

	return &wakupb.HistoryQuery{
		ContentFilters: contentFilters,
		StartTime:      toWakuTimestamp(req.StartTimeNs),
		EndTime:        toWakuTimestamp(req.EndTimeNs),
		PagingInfo:     buildWakuPagingInfo(req.PagingInfo),
	}
}

func buildPagingInfo(pi *wakupb.PagingInfo) *proto.PagingInfo {
	if pi == nil {
		return nil
	}
	var pagingInfo proto.PagingInfo
	pagingInfo.Limit = uint32(pi.PageSize)
	switch pi.Direction {
	case wakupb.PagingInfo_BACKWARD:
		pagingInfo.Direction = proto.SortDirection_SORT_DIRECTION_DESCENDING
	case wakupb.PagingInfo_FORWARD:
		pagingInfo.Direction = proto.SortDirection_SORT_DIRECTION_ASCENDING
	}
	if index := pi.Cursor; index != nil {
		pagingInfo.Cursor = &proto.Cursor{
			Cursor: &proto.Cursor_Index{
				Index: &proto.IndexCursor{
					Digest:       index.Digest,
					SenderTimeNs: uint64(index.SenderTime),
				}}}
	}
	return &pagingInfo
}

func buildWakuPagingInfo(pi *proto.PagingInfo) *wakupb.PagingInfo {
	if pi == nil {
		return nil
	}
	pagingInfo := &wakupb.PagingInfo{
		PageSize: uint64(pi.Limit),
	}
	switch pi.Direction {
	case proto.SortDirection_SORT_DIRECTION_ASCENDING:
		pagingInfo.Direction = wakupb.PagingInfo_FORWARD
	case proto.SortDirection_SORT_DIRECTION_DESCENDING:
		pagingInfo.Direction = wakupb.PagingInfo_BACKWARD
	}
	if ic := pi.Cursor.GetIndex(); ic != nil {
		pagingInfo.Cursor = &wakupb.Index{
			Digest:     ic.Digest,
			SenderTime: toWakuTimestamp(ic.SenderTimeNs),
		}
	}
	return pagingInfo
}

func fromWakuTimestamp(ts int64) uint64 {
	if ts < 0 {
		return 0
	}
	return uint64(ts)
}

func toWakuTimestamp(ts uint64) int64 {
	return int64(ts)
}
