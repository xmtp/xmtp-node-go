package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	queries "github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	message_apiv1 "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/subscriptions"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"github.com/xmtp/xmtp-node-go/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const DEFAULT_POLL_INTERVAL = 25 * time.Millisecond

type dbWorker struct {
	queries       *queries.Queries
	log           *zap.Logger
	subDispatcher *subscriptions.SubscriptionDispatcher
	ctx           context.Context
	wg            *sync.WaitGroup
	interval      time.Duration
}

func newDBWorker(
	ctx context.Context,
	log *zap.Logger,
	queries *queries.Queries,
	subDispatcher *subscriptions.SubscriptionDispatcher,
	pollInterval time.Duration,
) (*dbWorker, error) {
	if pollInterval == 0 {
		pollInterval = DEFAULT_POLL_INTERVAL
	}
	worker := dbWorker{
		ctx:           ctx,
		queries:       queries,
		log:           log,
		subDispatcher: subDispatcher,
		interval:      pollInterval,
		wg:            &sync.WaitGroup{},
	}

	if err := worker.start(); err != nil {
		return nil, err
	}

	return &worker, nil
}

func (w *dbWorker) getStartPoints() (int64, int64, error) {
	groupMessageID, err := w.queries.GetLatestGroupMessageID(w.ctx)
	if err != nil {
		return 0, 0, err
	}

	welcomeMessageID, err := w.queries.GetLatestWelcomeMessageID(w.ctx)
	if err != nil {
		return 0, 0, err
	}

	return groupMessageID, welcomeMessageID, nil
}

func (w *dbWorker) start() error {
	latestGroupMessageID, latestWelcomeMessageID, err := w.getStartPoints()
	if err != nil {
		w.log.Error("error getting start points", zap.Error(err))
		return err
	}

	tracing.GoPanicWrap(w.ctx, w.wg, "group-message-worker", func(ctx context.Context) {
		w.listenForGroupMessages(ctx, latestGroupMessageID)
	})

	tracing.GoPanicWrap(w.ctx, w.wg, "welcome-message-worker", func(ctx context.Context) {
		w.listenForWelcomeMessages(ctx, latestWelcomeMessageID)
	})

	return nil
}

func (w *dbWorker) listenForGroupMessages(ctx context.Context, startID int64) {
	currentID := startID
	ticker := time.NewTicker(w.interval)
	var (
		groupMessages []queries.GroupMessage
		err           error
	)

	for {
	SelectLoop:
		select {
		case <-ctx.Done():
			w.log.Info("context done, stopping group message worker")
			return
		case <-ticker.C:
			groupMessages, err = w.queries.GetAllGroupMessagesWithCursor(w.ctx, queries.GetAllGroupMessagesWithCursorParams{
				Cursor:  currentID,
				Numrows: 2000,
			})
			if err != nil {
				w.log.Error("error getting group messages", zap.Error(err))
				continue
			}

			for _, groupMessage := range groupMessages {
				groupMessageProto := mlsv1.GroupMessage{
					Version: &mlsv1.GroupMessage_V1_{
						V1: &mlsv1.GroupMessage_V1{
							Id:         uint64(groupMessage.ID),
							Data:       groupMessage.Data,
							CreatedNs:  uint64(groupMessage.CreatedAt.UnixNano()),
							SenderHmac: groupMessage.SenderHmac,
							ShouldPush: groupMessage.ShouldPush.Bool,
							GroupId:    groupMessage.GroupID,
						},
					},
				}

				// Log message details before dispatching
				w.log.Info("dispatching MLS group message to subscription system",
					zap.Int64("message_id", groupMessage.ID),
					zap.Uint64("created_ns", uint64(groupMessage.CreatedAt.UnixNano())),
					zap.String("group_id", fmt.Sprintf("%x", groupMessage.GroupID)),
					zap.Int("message_size_bytes", len(groupMessage.Data)),
					zap.Bool("should_push", groupMessage.ShouldPush.Bool),
				)

				data, err := proto.Marshal(&groupMessageProto)
				if err != nil {
					w.log.Error("error marshalling group message", zap.Error(err))
					// We cannot continue until this error is resolved
					break SelectLoop
				}

				w.subDispatcher.HandleEnvelope(&message_apiv1.Envelope{
					ContentTopic: topic.BuildMLSV1GroupTopic(groupMessage.GroupID),
					Message:      data,
					TimestampNs:  uint64(groupMessage.CreatedAt.UnixNano()),
				})
			}

			if len(groupMessages) > 0 {
				currentID = groupMessages[len(groupMessages)-1].ID
			}
		}
	}
}

func (w *dbWorker) listenForWelcomeMessages(ctx context.Context, startID int64) {
	currentID := startID
	ticker := time.NewTicker(w.interval)
	var (
		welcomeMessages []queries.WelcomeMessage
		err             error
	)
	for {
	SelectLoop:
		select {
		case <-ctx.Done():
			w.log.Info("context done, stopping welcome message worker")
			return
		case <-ticker.C:
			welcomeMessages, err = w.queries.GetAllWelcomeMessagesWithCursor(w.ctx, queries.GetAllWelcomeMessagesWithCursorParams{
				Cursor:  currentID,
				Numrows: 500,
			})
			if err != nil {
				w.log.Error("error getting welcome messages", zap.Error(err))
				continue
			}
			for _, welcomeMessage := range welcomeMessages {
				welcomeMessageProto := mlsv1.WelcomeMessage{
					Version: &mlsv1.WelcomeMessage_V1_{
						V1: &mlsv1.WelcomeMessage_V1{
							Id:               uint64(welcomeMessage.ID),
							Data:             welcomeMessage.Data,
							CreatedNs:        uint64(welcomeMessage.CreatedAt.UnixNano()),
							InstallationKey:  welcomeMessage.InstallationKey,
							HpkePublicKey:    welcomeMessage.HpkePublicKey,
							WrapperAlgorithm: types.WrapperAlgorithmToProto(types.WrapperAlgorithm(welcomeMessage.WrapperAlgorithm)),
							WelcomeMetadata:  welcomeMessage.WelcomeMetadata,
						},
					},
				}

				// Log message details before dispatching
				w.log.Info("dispatching MLS welcome message to subscription system",
					zap.Int64("message_id", welcomeMessage.ID),
					zap.Uint64("created_ns", uint64(welcomeMessage.CreatedAt.UnixNano())),
					zap.String("installation_key", fmt.Sprintf("%x", welcomeMessage.InstallationKey)),
					zap.Int("message_size_bytes", len(welcomeMessage.Data)),
				)

				data, err := proto.Marshal(&welcomeMessageProto)
				if err != nil {
					w.log.Error("error marshalling welcome message", zap.Error(err))
					break SelectLoop
				}
				w.subDispatcher.HandleEnvelope(&message_apiv1.Envelope{
					ContentTopic: topic.BuildMLSV1WelcomeTopic(welcomeMessage.InstallationKey),
					Message:      data,
					TimestampNs:  uint64(welcomeMessage.CreatedAt.UnixNano()),
				})
			}

			if len(welcomeMessages) > 0 {
				currentID = welcomeMessages[len(welcomeMessages)-1].ID
			}
		}
	}
}
