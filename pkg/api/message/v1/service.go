package api

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	wakupb "github.com/status-im/go-waku/waku/v2/protocol/pb"
	wakurelay "github.com/status-im/go-waku/waku/v2/protocol/relay"
	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/crdt"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	validXMTPTopicPrefix = "/xmtp/0/"
	contentTopicAllXMTP  = validXMTPTopicPrefix + "*"
)

type Service struct {
	messagev1.UnimplementedMessageApiServer

	// Configured as constructor options.
	log            *zap.Logger
	waku           *wakunode.WakuNode
	crdt           *crdt.Node
	writeToCRDTDS  bool
	readFromCRDTDS bool

	// Configured internally.
	ctx        context.Context
	ctxCancel  func()
	wg         sync.WaitGroup
	dispatcher *dispatcher
	relaySub   *wakurelay.Subscription
}

func NewService(node *wakunode.WakuNode, logger *zap.Logger, crdt *crdt.Node, writeToCRDTDS, readFromCRDTDS bool) (s *Service, err error) {
	s = &Service{
		waku:           node,
		log:            logger.Named("message/v1"),
		crdt:           crdt,
		writeToCRDTDS:  writeToCRDTDS,
		readFromCRDTDS: readFromCRDTDS,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	s.dispatcher = newDispatcher()
	s.relaySub, err = s.waku.Relay().Subscribe(s.ctx)
	if err != nil {
		return nil, errors.Wrap(err, "subscribing to relay")
	}
	tracing.GoPanicWrap(s.ctx, &s.wg, "broadcast", func(ctx context.Context) {
		if s.crdt != nil && s.readFromCRDTDS {
			for {
				select {
				case <-ctx.Done():
					return
				case env := <-s.crdt.EnvC:
					s.dispatcher.Submit(env.ContentTopic, env)
				}
			}
		} else {
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

func (s *Service) Publish(ctx context.Context, req *messagev1.PublishRequest) (*messagev1.PublishResponse, error) {
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

		if s.crdt != nil && s.writeToCRDTDS {
			err := s.crdt.Publish(ctx, env)
			if err != nil {
				return nil, status.Errorf(codes.Internal, err.Error())
			}
		}

		metrics.EmitPublishedEnvelope(ctx, env)
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

	if s.crdt != nil && s.readFromCRDTDS {
		return s.Subscribe(&messagev1.SubscribeRequest{
			ContentTopics: []string{contentTopicAllXMTP},
		}, stream)
	} else {
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
				if env == nil || !isValidTopic(env.ContentTopic) {
					continue
				}
				err := stream.Send(env)
				if err != nil {
					log.Error("sending envelope to subscriber", zap.Error(err))
				}
			}
		}
	}
}

func (s *Service) Query(ctx context.Context, req *messagev1.QueryRequest) (*messagev1.QueryResponse, error) {
	log := s.log.Named("query").With(zap.Strings("content_topics", req.ContentTopics))
	log.Info("received request")

	envs := []*messagev1.Envelope{}
	var pagingInfo *messagev1.PagingInfo

	if s.crdt != nil && s.readFromCRDTDS {
		var err error
		envs, pagingInfo, err = s.crdt.Query(ctx, req)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
	} else {
		store, ok := s.waku.Store().(*store.XmtpStore)
		if !ok {
			return nil, status.Errorf(codes.Internal, "waku store not xmtp store")
		}
		res, err := store.FindMessages(buildWakuQuery(req))
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		for _, msg := range res.Messages {
			envs = append(envs, buildEnvelope(msg))
		}

		pagingInfo = buildPagingInfo(res.PagingInfo)
	}

	return &messagev1.QueryResponse{
		Envelopes:  envs,
		PagingInfo: pagingInfo,
	}, nil
}

func (s *Service) BatchQuery(ctx context.Context, req *messagev1.BatchQueryRequest) (*messagev1.BatchQueryResponse, error) {
	log := s.log.Named("batchQuery")
	log.Info("batch size", zap.Int("num_queries", len(req.Requests)))
	// NOTE: in our implementation, we implicitly limit batch size to 50 requests
	if len(req.Requests) > 50 {
		return nil, status.Errorf(codes.InvalidArgument, "cannot exceed 50 requests in single batch")
	}
	// Naive implementation, perform all sub query requests sequentially
	responses := make([]*messagev1.QueryResponse, 0)
	for _, query := range req.Requests {
		// We execute the query using the existing Query API
		resp, err := s.Query(ctx, query)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		responses = append(responses, resp)
	}

	return &messagev1.BatchQueryResponse{
		Responses: responses,
	}, nil
}

func buildEnvelope(msg *wakupb.WakuMessage) *messagev1.Envelope {
	return &messagev1.Envelope{
		ContentTopic: msg.ContentTopic,
		TimestampNs:  fromWakuTimestamp(msg.Timestamp),
		Message:      msg.Payload,
	}
}

func buildWakuQuery(req *messagev1.QueryRequest) *wakupb.HistoryQuery {
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

func buildPagingInfo(pi *wakupb.PagingInfo) *messagev1.PagingInfo {
	if pi == nil {
		return nil
	}
	var pagingInfo messagev1.PagingInfo
	pagingInfo.Limit = uint32(pi.PageSize)
	switch pi.Direction {
	case wakupb.PagingInfo_BACKWARD:
		pagingInfo.Direction = messagev1.SortDirection_SORT_DIRECTION_DESCENDING
	case wakupb.PagingInfo_FORWARD:
		pagingInfo.Direction = messagev1.SortDirection_SORT_DIRECTION_ASCENDING
	}
	if index := pi.Cursor; index != nil {
		pagingInfo.Cursor = &messagev1.Cursor{
			Cursor: &messagev1.Cursor_Index{
				Index: &messagev1.IndexCursor{
					Digest:       index.Digest,
					SenderTimeNs: uint64(index.SenderTime),
				}}}
	}
	return &pagingInfo
}

func buildWakuPagingInfo(pi *messagev1.PagingInfo) *wakupb.PagingInfo {
	if pi == nil {
		return nil
	}
	pagingInfo := &wakupb.PagingInfo{
		PageSize: uint64(pi.Limit),
	}
	switch pi.Direction {
	case messagev1.SortDirection_SORT_DIRECTION_ASCENDING:
		pagingInfo.Direction = wakupb.PagingInfo_FORWARD
	case messagev1.SortDirection_SORT_DIRECTION_DESCENDING:
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

func isValidTopic(topic string) bool {
	return strings.HasPrefix(topic, validXMTPTopicPrefix)
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
