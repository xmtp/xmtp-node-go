package store

import (
	"time"

	"go.uber.org/zap"
)

type CleanerOptions struct {
	Enable        bool          `long:"enable" description:"Enable DB cleaner"`
	ActivePeriod  time.Duration `long:"active-period" description:"Time between successive runs of the cleaner when the last run was a full batch" default:"2s"`
	PassivePeriod time.Duration `long:"passive-period" description:"Time between successive runs of the cleaner when the last run was not a full batch" default:"5m"`
	RetentionDays int           `long:"retention-days" description:"Number of days in the past that messages must be before being deleted" default:"1"`
	BatchSize     int           `long:"batch-size" description:"Batch size of messages to be deleted in one iteration" default:"50000"`
	ReadTimeout   time.Duration `long:"read-timeout" description:"Timeout for reading from the database" default:"60s"`
	WriteTimeout  time.Duration `long:"write-timeout" description:"Timeout for writing to the database" default:"60s"`
}

func (s *XmtpStore) cleanerLoop() {
	log := s.log.Named("cleaner")

	for {
		started := time.Now().UTC()
		select {
		case <-s.ctx.Done():
			return
		default:
			count, err := s.deleteNonXMTPMessagesBatch(log)
			if err != nil {
				log.Error("error deleting non-xmtp messages", zap.Error(err), zap.Duration("duration", time.Since(started)))
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
	started := time.Now().UTC()
	timestampThreshold := time.Now().UTC().Add(time.Duration(s.cleaner.RetentionDays) * -1 * 24 * time.Hour).UnixNano()
	whereClause := "receivertimestamp < $1 AND should_expire IS TRUE"

	stmt, err := s.cleanerDB.Prepare("SELECT COUNT(1) FROM message WHERE " + whereClause + ";")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(timestampThreshold)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, nil
	}
	var count int64
	err = rows.Scan(&count)
	if err != nil {
		return 0, err
	}
	if count < int64(s.cleaner.BatchSize) {
		return 0, nil
	}

	// We use a single atomic query here instead of breaking it up to hit the
	// reader for the non-indexed NOT LIKE query first, because ctid can change
	// during a full vacuum of the DB, so we want to avoid conflicting with
	// that scenario and deleting the wrong data.
	stmt, err = s.cleanerDB.Prepare(`
		WITH msg AS (
			SELECT ctid
			FROM message
			WHERE ` + whereClause + `
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		DELETE FROM message WHERE ctid IN (TABLE msg);
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(timestampThreshold, s.cleaner.BatchSize)
	if err != nil {
		return 0, err
	}
	count, err = res.RowsAffected()
	if err != nil {
		return 0, err
	}

	log.Info("non-xmtp messages cleaner", zap.Int64("deleted", count), zap.Duration("duration", time.Since(started)))

	return count, nil
}
