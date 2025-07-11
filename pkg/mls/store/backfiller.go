package store

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BackfillGroupMessage represents a single row from the group_message table
// that requires backfilling of its is_commit field. The Data field is used
// for validation, and IsCommit is the inferred result populated by the
// classification step.
type BackfillGroupMessage struct {
	ID       int64
	Data     []byte
	IsCommit bool
}

// Backfiller defines the interface for a background backfill process.
// It can be started with Run() and gracefully stopped with Close().
type Backfiller interface {
	Run()
	Close()
}

// IsCommitBackfiller is responsible for populating the `is_commit` column
// in the `group_message` table. This field is inferred from message data
// using a validation service.
//
// In high-availability (HA) deployments where multiple instances of this
// service may be running concurrently, the backfiller is designed to be
// safe and non-conflicting. It relies on transactional locking and SKIP LOCKED
// semantics (or a manual locking scheme) to ensure that no two workers
// process the same rows. This avoids duplicate work and race conditions
// without requiring external coordination.
//
// The backfill process runs in a loop and periodically:
//  1. Starts a transaction
//  2. Selects up to N unprocessed group messages (e.g., where is_commit IS NULL)
//  3. Uses the MLSValidationService to classify each message
//  4. Updates the corresponding rows with the is_commit result
//
// If no work is available, it sleeps briefly before retrying.
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
	log *zap.Logger, validationService mlsvalidate.MLSValidationService,
) *IsCommitBackfiller {
	ctx, cancel := context.WithCancel(ctx)

	return &IsCommitBackfiller{
		ctx:               ctx,
		DB:                db,
		log:               log,
		validationService: validationService,
		cancel:            cancel,
	}
}

func (b *IsCommitBackfiller) classifyMessages(
	messages []BackfillGroupMessage,
) ([]BackfillGroupMessage, error) {
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
				foundMessages := true
				err := RunInTx(
					b.ctx,
					b.DB,
					nil,
					func(ctx context.Context, querier *queries.Queries) error {
						ids, err := querier.SelectEnvelopesForIsCommitBackfill(ctx)
						if err != nil {
							b.log.Error(
								"could not select next batch for is_commit backfill processing",
								zap.Error(err),
							)
							return err
						}

						if len(ids) == 0 {
							b.log.Info("No messages to classify")
							foundMessages = false
							return nil
						}

						convertedMsgs := make([]BackfillGroupMessage, len(ids))
						for i, msg := range ids {
							convertedMsgs[i] = BackfillGroupMessage{ID: msg.ID, Data: msg.Data}
						}

						classified, err := b.classifyMessages(convertedMsgs)
						if err != nil {
							b.log.Error("could not determine is_commit for batch", zap.Error(err))
							return err
						}

						for _, msg := range classified {
							b.log.Debug(
								"Updating is_commit status",
								zap.Int64("id", msg.ID),
								zap.Bool("is_commit", msg.IsCommit),
							)
							err = querier.UpdateIsCommitStatus(
								ctx,
								queries.UpdateIsCommitStatusParams{
									IsCommit: sql.NullBool{Bool: msg.IsCommit, Valid: true},
									ID:       msg.ID,
								},
							)
							if err != nil {
								b.log.Error(
									"could not update is_commit for message",
									zap.Int64("id", msg.ID),
									zap.Error(err),
								)
								continue
							}
						}
						return nil
					},
				)
				if err != nil {
					b.log.Error("Failed to execute is_commit backfill cycle", zap.Error(err))
				} else if !foundMessages {
					time.Sleep(1 * time.Minute)
				}
			}
		}
	}()
}

func (b *IsCommitBackfiller) Close() {
	b.cancel()
	b.WG.Wait()
}
