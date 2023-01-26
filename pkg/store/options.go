package store

import (
	"database/sql"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	wakustore "github.com/status-im/go-waku/waku/v2/protocol/store"
	"go.uber.org/zap"
)

type Option func(c *XmtpStore)

func WithLog(log *zap.Logger) Option {
	return func(s *XmtpStore) {
		s.log = log
	}
}

func WithHost(host host.Host) Option {
	return func(s *XmtpStore) {
		s.host = host
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

func WithResumeStartTime(resumeStartTime int64) Option {
	return func(s *XmtpStore) {
		s.resumeStartTime = resumeStartTime
	}
}

func WithCleaner(opts CleanerOptions) Option {
	return func(s *XmtpStore) {
		s.cleaner = opts
	}
}
