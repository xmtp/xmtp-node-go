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
	ReadCh            chan BackfillGroupMessage
	ClassifyCh        chan BackfillGroupMessage
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
		ReadCh:            make(chan BackfillGroupMessage, 100),
		ClassifyCh:        make(chan BackfillGroupMessage, 100),
		validationService: validationService,
		cancel:            cancel,
	}
}

// readFromDB pulls batches from the database
func (b *IsCommitBackfiller) readFromDB() {
	defer close(b.ReadCh)
	querier := queries.New(b.DB)

	for {
		select {
		case <-b.ctx.Done():
			return
		default:
			ids, err := querier.SelectEnvelopesForIsCommitBackfill(b.ctx)
			if err != nil {
				b.log.Error("SelectEnvelopesForIsCommitBackfill error", zap.Error(err))
				time.Sleep(60 * time.Second)
				continue
			}

			if len(ids) == 0 {
				b.log.Info("No messages to classify")
				time.Sleep(60 * time.Second)
				continue
			}

			for _, msg := range ids {
				msg := BackfillGroupMessage{ID: msg.ID, Data: msg.Data}
				b.ReadCh <- msg
			}

		}
	}
}

// classifyMessages uses external service to determine isCommit
func (b *IsCommitBackfiller) classifyMessages() {
	defer close(b.ClassifyCh)
	for {
		select {
		case <-b.ctx.Done():
			return
		case msg, ok := <-b.ReadCh:
			if !ok {
				return
			}
			classified, err := b.classifyMessage(&msg)
			if err != nil {
				b.log.Error("classifyMessage error", zap.Error(err))
				continue
			}
			b.ClassifyCh <- *classified
		}
	}
}

func (b *IsCommitBackfiller) classifyMessage(original *BackfillGroupMessage) (*BackfillGroupMessage, error) {
	validationResults, err := b.validationService.ValidateGroupMessage(b.ctx, original.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate group message: %v", err)
	}

	result := validationResults[0]

	original.IsCommit = result.IsCommit
	return original, nil
}

// writeBack updates the DB
func (b *IsCommitBackfiller) writeBack() {
	querier := queries.New(b.DB)
	for {
		select {
		case <-b.ctx.Done():
			return
		case msg, ok := <-b.ClassifyCh:
			if !ok {
				return
			}
			err := querier.UpdateIsCommitStatus(b.ctx, queries.UpdateIsCommitStatusParams{IsCommit: sql.NullBool{Bool: msg.IsCommit, Valid: true}, ID: msg.ID})
			if err != nil {
				b.log.Error("UpdateIsCommitStatus error", zap.Error(err))
				continue
			}
		}
	}
}

// Run orchestrates the pipeline
func (b *IsCommitBackfiller) Run() {
	b.WG.Add(3)
	go func() {
		defer b.WG.Done()
		b.readFromDB()
	}()
	go func() {
		defer b.WG.Done()
		b.classifyMessages()
	}()
	go func() {
		defer b.WG.Done()
		b.writeBack()
	}()
}

func (b *IsCommitBackfiller) Close() {
	b.cancel()
	b.WG.Wait()
}
