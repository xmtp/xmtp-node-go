package store

import (
	"database/sql"
	"time"

	"go.uber.org/zap"
)

type Option func(c *Store)

func WithLog(log *zap.Logger) Option {
	return func(s *Store) {
		s.log = log
	}
}

func WithDB(db *sql.DB) Option {
	return func(s *Store) {
		s.db = db
	}
}

func WithReaderDB(db *sql.DB) Option {
	return func(s *Store) {
		s.readerDB = db
	}
}

func WithCleanerDB(db *sql.DB) Option {
	return func(s *Store) {
		s.cleanerDB = db
	}
}

func WithStatsPeriod(statsPeriod time.Duration) Option {
	return func(s *Store) {
		s.statsPeriod = statsPeriod
	}
}

func WithCleaner(opts CleanerOptions) Option {
	return func(s *Store) {
		s.cleaner = opts
	}
}
