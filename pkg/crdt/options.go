package crdt

type Options struct {
	NodeKey           string   `long:"node-key" description:"Encoded node private key"`
	DataPath          string   `long:"data-path" description:"Path to directory that contains embedded data store files" default:"./data"`
	P2PPort           int      `long:"p2p-port" description:"Listen on this port for libp2p networking" default:"33123"`
	P2PBootstrapNodes []string `long:"p2p-bootstrap-node" description:"List of bootstrap nodes for the libp2p network"`
}
