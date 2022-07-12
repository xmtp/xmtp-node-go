package store

import (
	"database/sql"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
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

func WithMessageProvider(p store.MessageProvider) Option {
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
