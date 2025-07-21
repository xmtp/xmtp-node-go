package store

import (
	"time"
)

type StoreOptions struct {
	DbConnectionString       string        `long:"db-connection-string"        description:"Connection string for MLS DB"`
	DbReaderConnectionString string        `long:"reader-db-connection-string" description:"Connection string for MLS Reader DB"   env:"XMTPD_MLS_READER_DB_CONNECTION_STRING" default:""`
	ReadTimeout              time.Duration `long:"read-timeout"                description:"Timeout for reading from the database"                                             default:"10s"`
	WriteTimeout             time.Duration `long:"write-timeout"               description:"Timeout for writing to the database"                                               default:"10s"`
	MaxOpenConns             int           `long:"max-open-conns"              description:"Maximum number of open connections"                                                default:"80"`
}
