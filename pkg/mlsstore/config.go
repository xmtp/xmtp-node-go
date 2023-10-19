package mlsstore

import (
	"time"

	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type MlsOptions struct {
	DbConnectionString string        `long:"db-connection-string" description:"Connection string for MLS DB"`
	ReadTimeout        time.Duration `long:"read-timeout" description:"Timeout for reading from the database" default:"10s"`
	WriteTimeout       time.Duration `long:"write-timeout" description:"Timeout for writing to the database" default:"10s"`
	MaxOpenConns       int           `long:"max-open-conns" description:"Maximum number of open connections" default:"80"`
}

type Config struct {
	Log *zap.Logger
	DB  *bun.DB
}
