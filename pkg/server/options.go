package server

import (
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/api"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"github.com/xmtp/xmtp-node-go/pkg/store"
)

type RelayOptions struct {
	Disable                bool     `long:"no-relay"                   description:"Disable relay protocol"`
	Topics                 []string `long:"topics"                     description:"List of topics to listen"`
	MinRelayPeersToPublish int      `long:"min-relay-peers-to-publish" description:"Minimum number of peers to publish to Relay" default:"0"`
}

// MetricsOptions are settings used to start a prometheus server for obtaining
// useful node metrics to monitor the health of behavior of the go-waku node.
type MetricsOptions struct {
	Enable  bool   `long:"metrics"         description:"Enable the metrics server"`
	Address string `long:"metrics-address" description:"Listening address of the metrics server"   default:"127.0.0.1"`
	Port    int    `long:"metrics-port"    description:"Listening HTTP port of the metrics server" default:"8008"`
}

// TracingOptions are settings controlling collection of DD APM traces and error tracking.
type TracingOptions struct {
	Enable bool `long:"tracing" description:"Enable DD APM trace collection"`
}

// ProfilingOptions enable DD APM Profiling.
type ProfilingOptions struct {
	Enable bool `long:"enable"    description:"Enable CPU and Heap profiling"`
	// Following options have more overhead
	Block     bool `long:"block"     description:"Enable block profiling"`
	Mutex     bool `long:"mutex"     description:"Enable mutex profiling"`
	Goroutine bool `long:"goroutine" description:"Enable goroutine profiling"`
}

type AuthzOptions struct {
	DbConnectionString string        `long:"authz-db-connection-string" description:"Connection string for the authz DB"`
	ReadTimeout        time.Duration `long:"authz-read-timeout"         description:"Timeout for reading from the database" default:"10s"`
	WriteTimeout       time.Duration `long:"authz-write-timeout"        description:"Timeout for writing to the database"   default:"10s"`
}

type PruneConfig struct {
	MaxCycles      int  `long:"max-prune-cycles" description:"Maximum pruning cycles"                   default:"10"`
	CountDeletable bool `long:"count-deletable"  description:"Attempt to count all deletable envelopes"`
	DryRun         bool `long:"dry-run"          description:"Dry run mode"`
}

type PruneOptions struct {
	Log         LogOptions            `group:"Logging Options"`
	PruneConfig PruneConfig           `group:"Prune Options"   namespace:"prune"`
	MLSStore    mlsstore.StoreOptions `group:"MLS Options"     namespace:"mls-store"`
	Version     bool                  `                                              long:"version" description:"Output binary version and exit"`
}

type LogOptions struct {
	LogLevel string `short:"l" long:"log-level"    description:"Define the logging level, supported strings are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL, and their lower-case forms." default:"INFO"`
	// StaticCheck doesn't like duplicate params, but this is the only way to implement choice params
	//nolint:staticcheck
	LogEncoding string `          long:"log-encoding" description:"Log encoding format. Either console or json"                                                                                  default:"console" choice:"console"`
}

// Options contains all the available features and settings that can be
// configured via flags when executing go-waku as a service.
type Options struct {
	Port                   int           `short:"p" long:"port"                     description:"Libp2p TCP listening port (0 for random)"                                                       default:"60000"`
	Address                string        `          long:"address"                  description:"Listening address"                                                                              default:"0.0.0.0"`
	EnableWS               bool          `          long:"ws"                       description:"Enable websockets support"`
	WSPort                 int           `          long:"ws-port"                  description:"Libp2p TCP listening port for websocket connection (0 for random)"                              default:"60001"`
	WSAddress              string        `          long:"ws-address"               description:"Listening address for websocket connections"                                                    default:"0.0.0.0"`
	GenerateKey            bool          `          long:"generate-key"             description:"Generate private key file at path specified in --key-file"`
	NodeKey                string        `          long:"nodekey"                  description:"P2P node private key as hex. Can also be set with GOWAKU-NODEKEY env variable (default random)"`
	KeyFile                string        `          long:"key-file"                 description:"Path to a file containing the private key for the P2P node"                                     default:"./nodekey"`
	Overwrite              bool          `          long:"overwrite"                description:"When generating a keyfile, overwrite the nodekey file if it already exists"`
	StaticNodes            []string      `          long:"static-node"              description:"Multiaddr of peer to directly connect with. Option may be repeated"`
	KeepAlive              int           `          long:"keep-alive"               description:"Interval in seconds for pinging peers to keep the connection alive."                            default:"20"`
	CreateMessageMigration string        `          long:"create-message-migration" description:"Create a migration. Must provide a name"                                                        default:""`
	CreateAuthzMigration   string        `          long:"create-authz-migration"   description:"Create a migration for the auth db. Must provide a name"                                        default:""`
	CreateMlsMigration     string        `          long:"create-mls-migration"     description:"Create a migration for the mls db. Must provide a name"                                         default:""`
	WaitForDB              time.Duration `          long:"wait-for-db"              description:"wait for DB on start, up to specified duration"                                                 default:"30s"`
	Version                bool          `          long:"version"                  description:"Output binary version and exit"`
	GoProfiling            bool          `          long:"go-profiling"             description:"Enable Go profiling"`
	MetricsPeriod          time.Duration `          long:"metrics-period"           description:"Polling period for server status metrics"                                                       default:"30s"`

	Log LogOptions `group:"Logging Options"`

	API           api.Options                      `group:"API Options"              namespace:"api"`
	Authz         AuthzOptions                     `group:"Authz Options"`
	Relay         RelayOptions                     `group:"Relay Options"`
	Store         store.Options                    `group:"Store Options"            namespace:"store"`
	Metrics       MetricsOptions                   `group:"Metrics Options"`
	Tracing       TracingOptions                   `group:"DD APM Tracing Options"`
	Profiling     ProfilingOptions                 `group:"DD APM Profiling Options" namespace:"profiling"`
	MLSStore      mlsstore.StoreOptions            `group:"MLS Options"              namespace:"mls-store"`
	MLSValidation mlsvalidate.MLSValidationOptions `group:"MLS Validation Options"   namespace:"mls-validation"`
}
