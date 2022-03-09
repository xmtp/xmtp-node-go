package server

import "time"

type RelayOptions struct {
	Disable                bool     `long:"no-relay" description:"Disable relay protocol"`
	Topics                 []string `long:"topics" description:"List of topics to listen"`
	MinRelayPeersToPublish int      `long:"min-relay-peers-to-publish" description:"Minimum number of peers to publish to Relay" default:"1"`
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
	RetentionMaxDays     int      `long:"keep-history-days" description:"maximum number of days before a message is removed from the store" default:"30"`
	RetentionMaxMessages int      `long:"max-history-messages" description:"maximum number of messages to store" default:"50000"`
	Nodes                []string `long:"store-node" description:"Multiaddr of a peer that supports store protocol. Option may be repeated"`
}

func (s *StoreOptions) RetentionMaxDaysDuration() time.Duration {
	return time.Duration(s.RetentionMaxDays) * time.Hour * 24
}

// MetricsOptions are settings used to start a prometheus server for obtaining
// useful node metrics to monitor the health of behavior of the go-waku node.
type MetricsOptions struct {
	Enable  bool   `long:"metrics" description:"Enable the metrics server"`
	Address string `long:"metrics-address" description:"Listening address of the metrics server" default:"127.0.0.1"`
	Port    int    `long:"metrics-port" description:"Listening HTTP port of the metrics server" default:"8008"`
}

// Options contains all the available features and settings that can be
// configured via flags when executing go-waku as a service.
type Options struct {
	Port             int      `short:"p" long:"port" description:"Libp2p TCP listening port (0 for random)" default:"60000"`
	Address          string   `long:"address" description:"Listening address" default:"0.0.0.0"`
	EnableWS         bool     `long:"ws" description:"Enable websockets support"`
	WSPort           int      `long:"ws-port" description:"Libp2p TCP listening port for websocket connection (0 for random)" default:"60001"`
	WSAddress        string   `long:"ws-address" description:"Listening address for websocket connections" default:"0.0.0.0"`
	GenerateKey      bool     `long:"generate-key" description:"Generate private key file at path specified in --key-file"`
	NodeKey          string   `long:"nodekey" description:"P2P node private key as hex. Can also be set with GOWAKU-NODEKEY env variable (default random)"`
	KeyFile          string   `long:"key-file" description:"Path to a file containing the private key for the P2P node" default:"./nodekey"`
	Overwrite        bool     `long:"overwrite" description:"When generating a keyfile, overwrite the nodekey file if it already exists"`
	StaticNodes      []string `long:"static-node" description:"Multiaddr of peer to directly connect with. Option may be repeated"`
	KeepAlive        int      `long:"keep-alive" default:"20" description:"Interval in seconds for pinging peers to keep the connection alive."`
	DBPath           string   `long:"dbpath" default:"./store.db" description:"Path to DB file"`
	AdvertiseAddress string   `long:"advertise-address" default:"" description:"External address to advertise to other nodes (overrides --address and --ws-address flags)"`
	ShowAddresses    bool     `long:"show-addresses" description:"Display listening addresses according to current configuration"`
	LogLevel         string   `short:"l" long:"log-level" description:"Define the logging level, supported strings are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL, and their lower-case forms." default:"INFO"`
	LogEncoding      string   `long:"log-encoding" description:"Log encoding format. Either console or json" choice:"console" choice:"json" default:"console"`

	Relay     RelayOptions     `group:"Relay Options"`
	Store     StoreOptions     `group:"Store Options"`
	Filter    FilterOptions    `group:"Filter Options"`
	LightPush LightpushOptions `group:"LightPush Options"`
	Metrics   MetricsOptions   `group:"Metrics Options"`
}
