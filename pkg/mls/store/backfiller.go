package store

import (
	"context"
	"database/sql"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

// BackfillGroupMessage represents a row in the group_message table
type BackfillGroupMessage struct {
	ID       int64
	Data     []byte
	IsCommit bool
}

type Backfiller interface {
	Run()
	Close()
}

// IsCommitBackfiller encapsulates the backfill process
type IsCommitBackfiller struct {
	ctx               context.Context
	DB                *sql.DB
	log               *zap.Logger
	WG                sync.WaitGroup
	validationService mlsvalidate.MLSValidationService
	cancel            context.CancelFunc
}

var _ Backfiller = (*IsCommitBackfiller)(nil)

func NewIsCommitBackfiller(ctx context.Context, db *sql.DB,
	log *zap.Logger, validationService mlsvalidate.MLSValidationService) *IsCommitBackfiller {
	ctx, cancel := context.WithCancel(ctx)

	return &IsCommitBackfiller{
		ctx:               ctx,
		DB:                db,
		log:               log,
		validationService: validationService,
		cancel:            cancel,
	}
}

func (b *IsCommitBackfiller) RunInTx(
	ctx context.Context,
	opts *sql.TxOptions,
	fn func(ctx context.Context, txQueries *queries.Queries) error,
) error {
	tx, err := b.DB.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	var done bool

	defer func() {
		if !done {
			_ = tx.Rollback()
		}
	}()

	q := queries.New(tx)

	if err := fn(ctx, q); err != nil {
		return err
	}

	done = true
	return tx.Commit()
}

func (b *IsCommitBackfiller) classifyMessages(messages []BackfillGroupMessage) ([]BackfillGroupMessage, error) {
	payloads := make([][]byte, len(messages))
	for i, msg := range messages {
		payloads[i] = msg.Data
	}

	validationResults, err := b.validationService.ValidateGroupMessagePayloads(b.ctx, payloads)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate group message: %v", err)
	}

	for i, result := range validationResults {
		messages[i].IsCommit = result.IsCommit
	}

	return messages, nil
}

// Run orchestrates the pipeline
func (b *IsCommitBackfiller) Run() {
	b.WG.Add(1)
	go func() {
		defer b.WG.Done()
		for {
			select {
			case <-b.ctx.Done():
				return
			default:

				err := b.RunInTx(b.ctx, nil, func(ctx context.Context, querier *queries.Queries) error {
					ids, err := querier.SelectEnvelopesForIsCommitBackfill(ctx)
					if err != nil {
						b.log.Error("SelectEnvelopesForIsCommitBackfill error", zap.Error(err))
						return err
					}

					if len(ids) == 0 {
						b.log.Info("No messages to classify")
						time.Sleep(1 * time.Minute)
						return nil
					}

					convertedMsgs := make([]BackfillGroupMessage, len(ids))
					for i, msg := range ids {
						convertedMsgs[i] = BackfillGroupMessage{ID: msg.ID, Data: msg.Data}
					}

					classified, err := b.classifyMessages(convertedMsgs)
					if err != nil {
						b.log.Error("classifyMessages error", zap.Error(err))
						return err
					}

					for _, msg := range classified {
						b.log.Debug("Updating is_commit status", zap.Int64("id", msg.ID), zap.Bool("is_commit", msg.IsCommit))
						err = querier.UpdateIsCommitStatus(ctx, queries.UpdateIsCommitStatusParams{IsCommit: sql.NullBool{Bool: msg.IsCommit, Valid: true}, ID: msg.ID})
						if err != nil {
							b.log.Error("UpdateIsCommitStatus error", zap.Error(err))
							continue
						}
					}
					return nil
				})
				if err != nil {
					b.log.Error("RunInTx error", zap.Error(err))
					continue
				}

			}
		}
	}()
}

func (b *IsCommitBackfiller) Close() {
	b.cancel()
	b.WG.Wait()
}
