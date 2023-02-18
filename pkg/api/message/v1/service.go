package api

import (
	"context"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	wakupb "github.com/status-im/go-waku/waku/v2/protocol/pb"
	proto "github.com/xmtp/proto/v3/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "google.golang.org/protobuf/proto"
)

const (
	validXMTPTopicPrefix = "/xmtp/0/"
	contentTopicAllXMTP  = validXMTPTopicPrefix + "*"

	MaxContentTopicNameSize = 300

	// 1048576 - 300 - 62 = 1048214
	MaxMessageSize = pubsub.DefaultMaxMessageSize - MaxContentTopicNameSize - 62
)

type Service struct {
	proto.UnimplementedMessageApiServer

	// Configured as constructor options.
	log   *zap.Logger
	nats  *nats.Conn
	store *store.Store

	// Configured internally.
	ctx       context.Context
	ctxCancel func()
	wg        sync.WaitGroup
}

func NewService(logger *zap.Logger, nats *nats.Conn, store *store.Store) (s *Service, err error) {
	s = &Service{
		log:   logger.Named("message/v1"),
		nats:  nats,
		store: store,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	return s, nil
}

func (s *Service) Close() {
	s.log.Info("closing")
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

		if len(env.ContentTopic) > MaxContentTopicNameSize {
			return nil, status.Errorf(codes.InvalidArgument, "topic length too big")
		}

		wakuMsg := &wakupb.WakuMessage{
			ContentTopic: env.ContentTopic,
			Timestamp:    toWakuTimestamp(env.TimestampNs),
			Payload:      env.Message,
		}

		if len(env.Message) > MaxMessageSize {
			return nil, status.Errorf(codes.InvalidArgument, "message too big")
		}

		_, err := s.store.InsertMessage(wakuMsg)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		envB, err := pb.Marshal(env)
		if err != nil {
			return nil, err
		}
		err = s.nats.Publish(env.ContentTopic, envB)
		if err != nil {
			return nil, err
		}

		// Publish to the "all" topic too.
		err = s.nats.Publish(contentTopicAllXMTP, envB)
		if err != nil {
			return nil, err
		}
	}
	return &proto.PublishResponse{}, nil
}

func (s *Service) Subscribe(req *proto.SubscribeRequest, stream proto.MessageApi_SubscribeServer) error {
	log := s.log.Named("subscribe").With(zap.Strings("content_topics", req.ContentTopics))
	log.Debug("started")
	defer log.Debug("stopped")

	for _, topic := range req.ContentTopics {
		sub, err := s.nats.Subscribe(topic, func(msg *nats.Msg) {
			var env proto.Envelope
			err := pb.Unmarshal(msg.Data, &env)
			if err != nil {
				log.Info("error unmarshaling envelope", zap.Error(err))
				return
			}
			err = stream.Send(&env)
			if err != nil {
				log.Error("error sending envelope to subscriber", zap.Error(err))
			}
		})
		if err != nil {
			return err
		}
		defer sub.Unsubscribe()
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

	if len(req.ContentTopics) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "content topics required")
	}

	res, err := s.store.FindMessages(buildWakuQuery(req))
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
