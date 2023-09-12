package store

import (
	"database/sql"
	"errors"
	"time"

	"go.uber.org/zap"
)

var (
	ErrMissingLogOption       = errors.New("missing log option")
	ErrMissingDBOption        = errors.New("missing db option")
	ErrMissingReaderDBOption  = errors.New("missing reader db option")
	ErrMissingCleanerDBOption = errors.New("missing cleaner db option")
)

type Config struct {
	Options
	Log       *zap.Logger
	DB        *sql.DB
	ReaderDB  *sql.DB
	CleanerDB *sql.DB
}

type Options struct {
	Enable                   bool          `long:"enable" description:"Enable store"`
	DbConnectionString       string        `long:"db-connection-string" description:"A Postgres database connection string"`
	DbReaderConnectionString string        `long:"reader-db-connection-string" description:"A Postgres database reader connection string"`
	ReadTimeout              time.Duration `long:"db-read-timeout" description:"Timeout for reading from the database" default:"10s"`
	WriteTimeout             time.Duration `long:"db-write-timeout" description:"Timeout for writing to the database" default:"10s"`
	MaxOpenConns             int           `long:"max-open-conns" description:"Maximum number of open connections" default:"80"`
	MetricsPeriod            time.Duration `long:"metrics-period" description:"Polling period for store metrics" default:"30s"`

	Cleaner CleanerOptions `group:"DB Cleaner Options" namespace:"cleaner"`
}

func (c *Config) validate() error {
	if c.Log == nil {
		return ErrMissingLogOption
	}
	if c.DB == nil {
		return ErrMissingDBOption
	}
	if c.ReaderDB == nil {
		return ErrMissingReaderDBOption
	}
	if c.CleanerDB == nil {
		return ErrMissingCleanerDBOption
	}
	return nil
}
