package api

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	wakurelay "github.com/status-im/go-waku/waku/v2/protocol/relay"
	"github.com/tendermint/tendermint/types"
	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	xtm "github.com/xmtp/xmtp-node-go/pkg/tendermint"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	messagev1.UnimplementedMessageApiServer

	// Configured as constructor options.
	log *zap.Logger
	tm  *xtm.Node

	// Configured internally.
	ctx        context.Context
	ctxCancel  func()
	wg         sync.WaitGroup
	dispatcher *dispatcher
	relaySub   *wakurelay.Subscription
}

func NewService(log *zap.Logger, tm *xtm.Node) (s *Service, err error) {
	s = &Service{
		log: log.Named("message/v1"),
		tm:  tm,
	}
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	s.dispatcher = newDispatcher()
	if err != nil {
		return nil, errors.Wrap(err, "subscribing to relay")
	}
	tracing.GoPanicWrap(s.ctx, &s.wg, "broadcast", func(ctx context.Context) {
		resC, err := s.tm.HTTP().Subscribe(ctx, "1", "tm.event='Tx'", 100)
		if err != nil {
			s.log.Error("subscribing", zap.Error(err))
			return
		}
		for {
			select {
			case <-ctx.Done():
				return
			case res := <-resC:
				event, ok := res.Data.(types.EventDataTx)
				if !ok {
					s.log.Error("event is not a types.EventDataTx")
					continue
				}
				var env messagev1.Envelope
				err = proto.Unmarshal(event.Tx, &env)
				if err != nil {
					s.log.Error("parsing event", zap.Error(err))
					continue
				}
				s.dispatcher.Submit(env.ContentTopic, &env)
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

		envJSON, err := proto.Marshal(env)
		if err != nil {
			return nil, err
		}
		metrics.EmitPublishedEnvelope(ctx, env)
		_, err = s.tm.HTTP().BroadcastTxAsync(ctx, envJSON)
		if err != nil {
			return nil, err
		}
		// if res.CheckTx.IsErr() || res.DeliverTx.IsErr() {
		// 	log := res.CheckTx.Log
		// 	if res.DeliverTx.Log != "" {
		// 		log = res.DeliverTx.Log
		// 	}
		// 	return nil, errors.New(log)
		// }
	}
	return &messagev1.PublishResponse{}, nil
}

func (s *Service) Subscribe(req *messagev1.SubscribeRequest, stream messagev1.MessageApi_SubscribeServer) error {
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

// func (s *Service) SubscribeAll(req *messagev1.SubscribeAllRequest, stream messagev1.MessageApi_SubscribeAllServer) error {
// 	log := s.log.Named("subscribeAll")
// 	log.Info("started")
// 	defer log.Info("stopped")

// 	relaySub, err := s.waku.Relay().Subscribe(s.ctx)
// 	if err != nil {
// 		return err
// 	}

// 	defer relaySub.Unsubscribe()

// 	for {
// 		select {
// 		case <-stream.Context().Done():
// 			log.Info("stream closed")
// 			return nil
// 		case <-s.ctx.Done():
// 			log.Info("service closed")
// 			return nil
// 		case wakuEnv := <-relaySub.C:
// 			if wakuEnv == nil {
// 				continue
// 			}
// 			env := buildEnvelope(wakuEnv.Message())
// 			if env == nil || !isValidTopic(env.ContentTopic) {
// 				continue
// 			}
// 			err := stream.Send(env)
// 			if err != nil {
// 				log.Error("sending envelope to subscriber", zap.Error(err))
// 			}
// 		}
// 	}
// }

func (s *Service) Query(ctx context.Context, req *messagev1.QueryRequest) (*messagev1.QueryResponse, error) {
	log := s.log.Named("query").With(zap.Strings("content_topics", req.ContentTopics))
	log.Info("received request")

	// TODO: query multiple topics or update the request type to just have .Topic
	var limit uint32
	var cursor *messagev1.IndexCursor
	var reverse bool
	if req.PagingInfo != nil {
		limit = req.PagingInfo.Limit
		cursor = req.PagingInfo.Cursor.GetIndex()
		reverse = req.PagingInfo.Direction == messagev1.SortDirection_SORT_DIRECTION_DESCENDING
	}
	envs, _, err := s.tm.Query(ctx, req.ContentTopics[0], limit, cursor, reverse, req.StartTimeNs, req.EndTimeNs)
	if err != nil {
		return nil, err
	}
	// pageInfo := req.PagingInfo
	// if pageInfo != nil {
	// 	pageInfo.Cursor = &messagev1.Cursor{
	// 		Cursor: &messagev1.Cursor_Index{
	// 			Index: cursor,
	// 		},
	// 	}
	// }

	return &messagev1.QueryResponse{
		Envelopes: envs,
		// PagingInfo: pageInfo,
	}, nil
}

func isValidTopic(topic string) bool {
	return strings.HasPrefix(topic, "/xmtp/0/")
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
