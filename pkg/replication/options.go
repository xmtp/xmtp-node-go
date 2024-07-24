package replication

type ApiOptions struct {
	Port int `short:"p" long:"port" description:"Port to listen on" default:"5050"`
}

type DbOptions struct {
	ReaderConnectionString string `long:"reader-connection-string" description:"Reader connection string"`
	WriterConnectionString string `long:"writer-connection-string" description:"Writer connection string" required:"true"`
}

type Options struct {
	LogLevel string `short:"l" long:"log-level" description:"Define the logging level, supported strings are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL, and their lower-case forms." default:"INFO"`
	//nolint:staticcheck
	LogEncoding string `long:"log-encoding" description:"Log encoding format. Either console or json" choice:"console" choice:"json" default:"console"`

	API ApiOptions `group:"API Options" namespace:"api"`
	DB  DbOptions  `group:"Database Options" namespace:"db"`
}
