package store

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	nonXMTPMessageRetentionDays = 3
)

func (s *XmtpStore) cleanerLoop() {
	log := s.log.Named("cleaner")

	for {
		select {
		case <-s.ctx.Done():
		default:
			err := s.deleteNonXMTPMessagesBatch(log)
			if err != nil {
				log.Error("error deleting non-xmtp messages", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (s *XmtpStore) deleteNonXMTPMessagesBatch(log *zap.Logger) error {
	stmt, err := s.readerDB.Prepare("SELECT id FROM message WHERE receivertimestamp < $1 AND contenttopic NOT LIKE '/xmtp/%' LIMIT 1000")
	if err != nil {
		return err
	}
	timestampThreshold := time.Now().UTC().Add(time.Duration(nonXMTPMessageRetentionDays) * -1 * 24 * time.Hour).UnixNano()
	rows, err := stmt.Query(timestampThreshold)
	if err != nil {
		return err
	}
	defer rows.Close()

	ids := []any{}
	for rows.Next() {
		var id []byte
		err = rows.Scan(&id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	for i := 0; i < len(ids); i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	stmt, err = s.db.Prepare("DELETE FROM message WHERE id IN (" + strings.Join(placeholders, ",") + ")")
	if err != nil {
		return err
	}
	res, err := stmt.Exec(ids...)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	log.Info("deleted non-xmtp messages", zap.Int64("deleted", count), zap.Int("ids", len(ids)))

	return nil
}
