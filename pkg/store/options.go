package store

import (
	"database/sql"
	"time"

	wakustore "github.com/status-im/go-waku/waku/v2/protocol/store"
	"go.uber.org/zap"
)

type Option func(c *XmtpStore)

func WithLog(log *zap.Logger) Option {
	return func(s *XmtpStore) {
		s.log = log
	}
}

func WithDB(db *sql.DB) Option {
	return func(s *XmtpStore) {
		s.db = db
	}
}

func WithReaderDB(db *sql.DB) Option {
	return func(s *XmtpStore) {
		s.readerDB = db
	}
}

func WithCleanerDB(db *sql.DB) Option {
	return func(s *XmtpStore) {
		s.cleanerDB = db
	}
}

// TODO: replace MessageProvider with a local type and remove waku reference
func WithMessageProvider(p wakustore.MessageProvider) Option {
	return func(s *XmtpStore) {
		s.msgProvider = p
	}
}

func WithStatsPeriod(statsPeriod time.Duration) Option {
	return func(s *XmtpStore) {
		s.statsPeriod = statsPeriod
	}
}

func WithCleaner(opts CleanerOptions) Option {
	return func(s *XmtpStore) {
		s.cleaner = opts
	}
}
