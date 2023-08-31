package server

import (
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/api"
	"github.com/xmtp/xmtp-node-go/pkg/store"
)

type RelayOptions struct {
	Disable                bool     `long:"no-relay" description:"Disable relay protocol"`
	Topics                 []string `long:"topics" description:"List of topics to listen"`
	MinRelayPeersToPublish int      `long:"min-relay-peers-to-publish" description:"Minimum number of peers to publish to Relay" default:"0"`
}

type FilterOptions struct {
	Enable  bool     `long:"filter" description:"Enable filter protocol"`
	Nodes   []string `long:"filter-node" description:"Multiaddr of a peer that supports filter protocol. Option may be repeated"`
	Timeout int      `long:"filter-timeout" description:"Timeout for filter node in seconds" default:"14400"`
}

// LightpushOptions are settings used to enable the lightpush protocol. This is
// a lightweight protocol used to avoid having to run the relay protocol which
// is more resource intensive. With this protocol a message is pushed to a peer
// that supports both the lightpush protocol and relay protocol. That peer will
// broadcast the message and return a confirmation that the message was
// broadcasted
type LightpushOptions struct {
	Enable bool     `long:"lightpush" description:"Enable lightpush protocol"`
	Nodes  []string `long:"lightpush-node" description:"Multiaddr of a peer that supports lightpush protocol. Option may be repeated"`
}

// StoreOptions are settings used for enabling the store protocol, used to
// retrieve message history from other nodes as well as acting as a store
// node and provide message history to nodes that ask for it.
type StoreOptions struct {
	Enable                   bool          `long:"store" description:"Enable store protocol"`
	DbConnectionString       string        `long:"message-db-connection-string" description:"A Postgres database connection string"`
	DbReaderConnectionString string        `long:"message-db-reader-connection-string" description:"A Postgres database reader connection string"`
	ReadTimeout              time.Duration `long:"message-db-read-timeout" description:"Timeout for reading from the database" default:"10s"`
	WriteTimeout             time.Duration `long:"message-db-write-timeout" description:"Timeout for writing to the database" default:"10s"`
	MaxOpenConns             int           `long:"max-open-conns" description:"Maximum number of open connections" default:"80"`
}

// MetricsOptions are settings used to start a prometheus server for obtaining
// useful node metrics to monitor the health of behavior of the node.
type MetricsOptions struct {
	Enable       bool          `long:"metrics" description:"Enable the metrics server"`
	Address      string        `long:"metrics-address" description:"Listening address of the metrics server" default:"127.0.0.1"`
	Port         int           `long:"metrics-port" description:"Listening HTTP port of the metrics server" default:"8008"`
	StatusPeriod time.Duration `long:"metrics-period" description:"Polling period for server status metrics" default:"30s"`
}

type NATSOptions struct {
	URL string `long:"url" description:"URL of NATS server" default:""`
}

// TracingOptions are settings controlling collection of DD APM traces and error tracking.
type TracingOptions struct {
	Enable bool `long:"tracing" description:"Enable DD APM trace collection"`
}

// ProfilingOptions enable DD APM Profiling.
type ProfilingOptions struct {
	Enable bool `long:"enable" description:"Enable CPU and Heap profiling"`
	// Following options have more overhead
	Block     bool `long:"block" description:"Enable block profiling"`
	Mutex     bool `long:"mutex" description:"Enable mutex profiling"`
	Goroutine bool `long:"goroutine" description:"Enable goroutine profiling"`
}

type AuthzOptions struct {
	DbConnectionString string        `long:"authz-db-connection-string" description:"Connection string for the authz DB"`
	ReadTimeout        time.Duration `long:"authz-read-timeout" description:"Timeout for reading from the database" default:"10s"`
	WriteTimeout       time.Duration `long:"authz-write-timeout" description:"Timeout for writing to the database" default:"10s"`
}

// Options contains all the available features and settings that can be
// configured.
type Options struct {
	LogLevel string `short:"l" long:"log-level" description:"Define the logging level, supported strings are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL, and their lower-case forms." default:"INFO"`
	// StaticCheck doesn't like duplicate params, but this is the only way to implement choice params
	//nolint:staticcheck
	LogEncoding            string        `long:"log-encoding" description:"Log encoding format. Either console or json" choice:"console" choice:"json" default:"console"`
	CreateMessageMigration string        `long:"create-message-migration" default:"" description:"Create a migration. Must provide a name"`
	CreateAuthzMigration   string        `long:"create-authz-migration" default:"" description:"Create a migration for the auth db. Must provide a name"`
	WaitForDB              time.Duration `long:"wait-for-db" description:"wait for DB on start, up to specified duration"`
	Version                bool          `long:"version" description:"Output binary version and exit"`
	GoProfiling            bool          `long:"go-profiling" description:"Enable Go profiling"`

	API       api.Options          `group:"API Options" namespace:"api"`
	NATS      NATSOptions          `group:"NATS Options" namespace:"nats"`
	Authz     AuthzOptions         `group:"Authz Options"`
	Store     StoreOptions         `group:"Store Options"`
	Metrics   MetricsOptions       `group:"Metrics Options"`
	Tracing   TracingOptions       `group:"DD APM Tracing Options"`
	Profiling ProfilingOptions     `group:"DD APM Profiling Options" namespace:"profiling"`
	Cleaner   store.CleanerOptions `group:"DB Cleaner Options" namespace:"cleaner"`
}
