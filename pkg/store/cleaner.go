package store

import (
	"time"

	"go.uber.org/zap"
)

type CleanerOptions struct {
	Enable        bool          `long:"enable" description:"Enable DB cleaner"`
	ActivePeriod  time.Duration `long:"active-period" description:"Time between successive runs of the cleaner when the last run was a full batch" default:"2s"`
	PassivePeriod time.Duration `long:"passive-period" description:"Time between successive runs of the cleaner when the last run was not a full batch" default:"5m"`
	RetentionDays int           `long:"retention-days" description:"Number of days in the past that messages must be before being deleted" default:"3"`
	BatchSize     int           `long:"batch-size" description:"Batch size of messages to be deleted in one iteration" default:"1000"`
}

func (s *XmtpStore) cleanerLoop() {
	log := s.log.Named("cleaner")

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			count, err := s.deleteNonXMTPMessagesBatch(log)
			if err != nil {
				log.Error("error deleting non-xmtp messages", zap.Error(err))
			}
			if count >= int64(s.cleaner.BatchSize-10) {
				time.Sleep(s.cleaner.ActivePeriod)
			} else {
				time.Sleep(s.cleaner.PassivePeriod)
			}
		}
	}
}

func (s *XmtpStore) deleteNonXMTPMessagesBatch(log *zap.Logger) (int64, error) {
	// We use a single atomic query here instead of breaking it up to hit the
	// reader for the non-indexed NOT LIKE query first, because ctid can change
	// during a full vacuum of the DB, so we want to avoid conflicting with
	// that scenario and deleting the wrong data.
	started := time.Now().UTC()
	stmt, err := s.db.Prepare(`
		WITH msg AS (
			SELECT ctid
			FROM message
			WHERE receivertimestamp < $1 AND contenttopic NOT LIKE '/xmtp/%'
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		DELETE FROM message WHERE ctid IN (TABLE msg);
	`)
	if err != nil {
		return 0, err
	}
	timestampThreshold := time.Now().UTC().Add(time.Duration(s.cleaner.RetentionDays) * -1 * 24 * time.Hour).UnixNano()
	res, err := stmt.Exec(timestampThreshold, s.cleaner.BatchSize)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	log.Info("non-xmtp messages cleaner", zap.Int64("deleted", count), zap.Duration("duration", time.Since(started)))

	return count, nil
}
