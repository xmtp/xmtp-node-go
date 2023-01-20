package store

import (
	"time"

	"go.uber.org/zap"
)

type CleanerOptions struct {
	Enable        bool          `long:"enable" description:"Enable DB cleaner"`
	Period        time.Duration `long:"period" description:"Time between successive runs of the cleaner" default:"5s"`
	RetentionDays int           `long:"retention-days" description:"Number of days in the past that messages must be before being deleted" default:"3"`
}

func (s *XmtpStore) cleanerLoop() {
	log := s.log.Named("cleaner")

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			err := s.deleteNonXMTPMessagesBatch(log)
			if err != nil {
				log.Error("error deleting non-xmtp messages", zap.Error(err))
			}
			time.Sleep(s.cleaner.Period)
		}
	}
}

func (s *XmtpStore) deleteNonXMTPMessagesBatch(log *zap.Logger) error {
	stmt, err := s.db.Prepare(`
		WITH msg AS (
			SELECT ctid
			FROM message
			WHERE receivertimestamp < $1 AND contenttopic NOT LIKE '/xmtp/%'
			LIMIT 1000
			FOR UPDATE SKIP LOCKED
		)
		DELETE FROM message WHERE ctid IN (TABLE msg);
	`)
	if err != nil {
		return err
	}
	timestampThreshold := time.Now().UTC().Add(time.Duration(s.cleaner.RetentionDays) * -1 * 24 * time.Hour).UnixNano()
	res, err := stmt.Exec(timestampThreshold)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	log.Info("deleted non-xmtp messages", zap.Int64("deleted", count))

	return nil
}
