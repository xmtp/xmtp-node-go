package server

import (
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/api"
	xtm "github.com/xmtp/xmtp-node-go/pkg/tendermint"
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
	Enable               bool     `long:"store" description:"Enable store protocol"`
	ShouldResume         bool     `long:"resume" description:"fix the gap in message history"`
	ResumeStartTime      int64    `long:"resume-start-time" description:"resume from this start time" default:"-1"`
	RetentionMaxDays     int      `long:"keep-history-days" description:"maximum number of days before a message is removed from the store" default:"30"`
	RetentionMaxMessages int      `long:"max-history-messages" description:"maximum number of messages to store" default:"50000"`
	Nodes                []string `long:"store-node" description:"Multiaddr of a peer that supports store protocol. Option may be repeated"`
	DbConnectionString   string   `long:"message-db-connection-string" description:"A Postgres database connection string"`
}

func (s *StoreOptions) RetentionMaxDaysDuration() time.Duration {
	return time.Duration(s.RetentionMaxDays) * time.Hour * 24
}

// MetricsOptions are settings used to start a prometheus server for obtaining
// useful node metrics to monitor the health of behavior of the go-waku node.
type MetricsOptions struct {
	Enable       bool          `long:"metrics" description:"Enable the metrics server"`
	Address      string        `long:"metrics-address" description:"Listening address of the metrics server" default:"127.0.0.1"`
	Port         int           `long:"metrics-port" description:"Listening HTTP port of the metrics server" default:"8008"`
	StatusPeriod time.Duration `long:"metrics-period" description:"Polling period for server status metrics" default:"30s"`
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
	DbConnectionString string `long:"authz-db-connection-string" description:"Connection string for the authz DB"`
}

// Options contains all the available features and settings that can be
// configured via flags when executing go-waku as a service.
type Options struct {
	Port        int      `short:"p" long:"port" description:"Libp2p TCP listening port (0 for random)" default:"60000"`
	Address     string   `long:"address" description:"Listening address" default:"0.0.0.0"`
	EnableWS    bool     `long:"ws" description:"Enable websockets support"`
	WSPort      int      `long:"ws-port" description:"Libp2p TCP listening port for websocket connection (0 for random)" default:"60001"`
	WSAddress   string   `long:"ws-address" description:"Listening address for websocket connections" default:"0.0.0.0"`
	GenerateKey bool     `long:"generate-key" description:"Generate private key file at path specified in --key-file"`
	NodeKey     string   `long:"nodekey" description:"P2P node private key as hex. Can also be set with GOWAKU-NODEKEY env variable (default random)"`
	KeyFile     string   `long:"key-file" description:"Path to a file containing the private key for the P2P node" default:"./nodekey"`
	Overwrite   bool     `long:"overwrite" description:"When generating a keyfile, overwrite the nodekey file if it already exists"`
	StaticNodes []string `long:"static-node" description:"Multiaddr of peer to directly connect with. Option may be repeated"`
	KeepAlive   int      `long:"keep-alive" default:"20" description:"Interval in seconds for pinging peers to keep the connection alive."`
	LogLevel    string   `short:"l" long:"log-level" description:"Define the logging level, supported strings are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL, and their lower-case forms." default:"INFO"`
	// StaticCheck doesn't like duplicate params, but this is the only way to implement choice params
	//nolint:staticcheck
	LogEncoding            string        `long:"log-encoding" description:"Log encoding format. Either console or json" choice:"console" choice:"json" default:"console"`
	CreateMessageMigration string        `long:"create-message-migration" default:"" description:"Create a migration. Must provide a name"`
	CreateAuthzMigration   string        `long:"create-authz-migration" default:"" description:"Create a migration for the auth db. Must provide a name"`
	WaitForDB              time.Duration `long:"wait-for-db" description:"wait for DB on start, up to specified duration"`
	Version                bool          `long:"version" description:"Output binary version and exit"`
	GoProfiling            bool          `long:"go-profiling" description:"Enable Go profiling"`

	API        api.Options      `group:"API Options" namespace:"api"`
	Tendermint xtm.Options      `group:"Tendermint Options" namespace:"tm"`
	Authz      AuthzOptions     `group:"Authz Options"`
	Relay      RelayOptions     `group:"Relay Options"`
	Store      StoreOptions     `group:"Store Options"`
	Filter     FilterOptions    `group:"Filter Options"`
	LightPush  LightpushOptions `group:"LightPush Options"`
	Metrics    MetricsOptions   `group:"Metrics Options"`
	Tracing    TracingOptions   `group:"DD APM Tracing Options"`
	Profiling  ProfilingOptions `group:"DD APM Profiling Options" namespace:"profiling"`
}
