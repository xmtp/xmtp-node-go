package server

import (
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/api"
)

// MetricsOptions are settings used to start a prometheus server for obtaining
// useful node metrics to monitor the health of behavior of the node.
type MetricsOptions struct {
	Enable       bool          `long:"metrics" description:"Enable the metrics server"`
	Address      string        `long:"metrics-address" description:"Listening address of the metrics server" default:"127.0.0.1"`
	Port         int           `long:"metrics-port" description:"Listening HTTP port of the metrics server" default:"8009"`
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

// Options contains all the available features and settings that can be
// configured via flags when executing xmtpd.
type Options struct {
	LogLevel    string `short:"l" long:"log-level" description:"Define the logging level, supported strings are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL, and their lower-case forms." default:"INFO"`
	LogEncoding string `long:"log-encoding" description:"Log encoding format. Either console or json" default:"console"`
	Version     bool   `long:"version" description:"Output binary version and exit"`
	GoProfiling bool   `long:"go-profiling" description:"Enable Go profiling"`

	GenerateKey        bool     `long:"generate-key" description:"Generate private key file at path specified in --key-file"`
	NodeKey            string   `long:"node-key" description:"Encoded node private key"`
	DataPath           string   `long:"data-path" description:"Path to directory that contains embedded data store files" default:"./data"`
	P2PPort            int      `long:"p2p-port" description:"Listen on this port for libp2p networking" default:"9001"`
	P2PPersistentPeers []string `long:"p2p-persistent-peer" description:"List of persistent peers for the libp2p network"`

	API       api.Options      `group:"API Options" namespace:"api"`
	Metrics   MetricsOptions   `group:"Metrics Options"`
	Tracing   TracingOptions   `group:"DD APM Tracing Options"`
	Profiling ProfilingOptions `group:"DD APM Profiling Options" namespace:"profiling"`
}
