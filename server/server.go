package server

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	dssql "github.com/ipfs/go-ds-sql"
	"go.uber.org/zap"

	"github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"

	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/status-im/go-waku/waku/metrics"
	"github.com/status-im/go-waku/waku/persistence"
	"github.com/status-im/go-waku/waku/persistence/sqlite"
	"github.com/status-im/go-waku/waku/v2/node"
	"github.com/status-im/go-waku/waku/v2/protocol/filter"
	"github.com/status-im/go-waku/waku/v2/protocol/lightpush"
	"github.com/status-im/go-waku/waku/v2/protocol/relay"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
)

type Server struct {
	logger        *zap.Logger
	hostAddr      *net.TCPAddr
	db            *sql.DB
	metricsServer *metrics.Server
	wakuNode      *node.WakuNode
	ctx           context.Context
}

// Create a new Server
func New(options Options) (server *Server) {
	server = new(Server)
	var err error

	server.logger = utils.InitLogger(options.LogEncoding)
	server.hostAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", options.Address, options.Port))
	failOnErr(err, "invalid host address")

	prvKey, err := getPrivKey(options)
	failOnErr(err, "nodekey error")

	if options.DBPath == "" {
		failOnErr(errors.New("dbpath can't be null"), "")
	}

	server.db, err = sqlite.NewDB(options.DBPath)
	failOnErr(err, "Could not connect to DB")

	server.ctx = context.Background()

	if options.Metrics.Enable {
		server.metricsServer = metrics.NewMetricsServer(options.Metrics.Address, options.Metrics.Port, server.logger.Sugar())
		go server.metricsServer.Start()
	}

	nodeOpts := []node.WakuNodeOption{
		node.WithLogger(server.logger),
		node.WithPrivateKey(prvKey),
		node.WithHostAddress(server.hostAddr),
		node.WithKeepAlive(time.Duration(options.KeepAlive) * time.Second),
	}

	if options.EnableWS {
		wsMa, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d/ws", options.WSAddress, options.WSPort))
		nodeOpts = append(nodeOpts, node.WithMultiaddress([]multiaddr.Multiaddr{wsMa}))
	}

	libp2pOpts := node.DefaultLibP2POptions
	libp2pOpts = append(libp2pOpts, libp2p.NATPortMap()) // Attempt to open ports using uPNP for NATed hosts.)

	// Create persistent peerstore
	queries, err := sqlite.NewQueries("peerstore", server.db)
	failOnErr(err, "Peerstore")

	datastore := dssql.NewDatastore(server.db, queries)
	opts := pstoreds.DefaultOpts()
	peerStore, err := pstoreds.NewPeerstore(server.ctx, datastore, opts)
	failOnErr(err, "Peerstore")

	libp2pOpts = append(libp2pOpts, libp2p.Peerstore(peerStore))

	nodeOpts = append(nodeOpts, node.WithLibP2POptions(libp2pOpts...))

	if !options.Relay.Disable {
		var wakurelayopts []pubsub.Option
		wakurelayopts = append(wakurelayopts, pubsub.WithPeerExchange(true))
		nodeOpts = append(nodeOpts, node.WithWakuRelayAndMinPeers(options.Relay.MinRelayPeersToPublish, wakurelayopts...))
	}

	if options.Filter.Enable {
		nodeOpts = append(nodeOpts, node.WithWakuFilter(true, filter.WithTimeout(time.Duration(options.Filter.Timeout)*time.Second)))
	}

	if options.Store.Enable {
		nodeOpts = append(nodeOpts, node.WithWakuStoreAndRetentionPolicy(options.Store.ShouldResume, options.Store.RetentionMaxDaysDuration(), options.Store.RetentionMaxMessages))
		dbStore, err := persistence.NewDBStore(server.logger.Sugar(), persistence.WithDB(server.db), persistence.WithRetentionPolicy(options.Store.RetentionMaxMessages, options.Store.RetentionMaxDaysDuration()))
		failOnErr(err, "DBStore")
		nodeOpts = append(nodeOpts, node.WithMessageProvider(dbStore))
	}

	if options.LightPush.Enable {
		nodeOpts = append(nodeOpts, node.WithLightPush())
	}

	server.wakuNode, err = node.New(server.ctx, nodeOpts...)

	failOnErr(err, "Wakunode")

	addPeers(server.wakuNode, options.Store.Nodes, store.StoreID_v20beta3)
	addPeers(server.wakuNode, options.LightPush.Nodes, lightpush.LightPushID_v20beta1)
	addPeers(server.wakuNode, options.Filter.Nodes, filter.FilterID_v20beta1)

	if err = server.wakuNode.Start(); err != nil {
		server.logger.Fatal(fmt.Errorf("could not start waku node, %w", err).Error())
	}

	if len(options.Relay.Topics) == 0 {
		options.Relay.Topics = []string{string(relay.DefaultWakuTopic)}
	}

	if !options.Relay.Disable {
		for _, nodeTopic := range options.Relay.Topics {
			sub, err := server.wakuNode.Relay().SubscribeToTopic(server.ctx, nodeTopic)
			failOnErr(err, "Error subscring to topic")
			// Unregister from broadcaster. Otherwise this channel will fill until it blocks publishing
			server.wakuNode.Broadcaster().Unregister(sub.C)
		}
	}

	for _, n := range options.StaticNodes {
		go func(node string) {
			err = server.wakuNode.DialPeer(server.ctx, node)
			if err != nil {
				server.logger.Error("error dialing peer ", zap.Error(err))
			}
		}(n)
	}

	return server
}

func (server *Server) WaitForShutdown() {
	// Wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	server.logger.Info("Received signal, shutting down...")

	// shut the node down
	server.wakuNode.Stop()

	if server.metricsServer != nil {
		err := server.metricsServer.Stop(server.ctx)
		failOnErr(err, "MetricsClose")
	}
}

func addPeers(wakuNode *node.WakuNode, addresses []string, protocol protocol.ID) {
	for _, addrString := range addresses {
		if addrString == "" {
			continue
		}

		addr, err := multiaddr.NewMultiaddr(addrString)
		failOnErr(err, "invalid multiaddress")

		_, err = wakuNode.AddPeer(addr, protocol)
		failOnErr(err, "error adding peer")
	}
}

func loadPrivateKeyFromFile(path string) (*ecdsa.PrivateKey, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dst := make([]byte, hex.DecodedLen(len(src)))
	_, err = hex.Decode(dst, src)
	if err != nil {
		return nil, err
	}

	p, err := libp2pcrypto.UnmarshalSecp256k1PrivateKey(dst)
	if err != nil {
		return nil, err
	}

	privKey := (*ecdsa.PrivateKey)(p.(*libp2pcrypto.Secp256k1PrivateKey))
	privKey.Curve = crypto.S256()

	return privKey, nil
}

func checkForPrivateKeyFile(path string, overwrite bool) error {
	_, err := os.Stat(path)

	if err == nil && !overwrite {
		return fmt.Errorf("%s already exists. Use --overwrite to overwrite the file", path)
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func generatePrivateKey() ([]byte, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	privKey := libp2pcrypto.PrivKey((*libp2pcrypto.Secp256k1PrivateKey)(key))

	b, err := privKey.Raw()
	if err != nil {
		return nil, err
	}

	output := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(output, b)

	return output, nil
}

func WritePrivateKeyToFile(path string, overwrite bool) error {
	if err := checkForPrivateKeyFile(path, overwrite); err != nil {
		return err
	}

	output, err := generatePrivateKey()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, output, 0600)
}

func getPrivKey(options Options) (*ecdsa.PrivateKey, error) {
	var prvKey *ecdsa.PrivateKey
	var err error
	if options.NodeKey != "" {
		if prvKey, err = crypto.HexToECDSA(options.NodeKey); err != nil {
			return nil, fmt.Errorf("error converting key into valid ecdsa key: %w", err)
		}
	} else {
		keyString := os.Getenv("GOWAKU-NODEKEY")
		if keyString != "" {
			if prvKey, err = crypto.HexToECDSA(keyString); err != nil {
				return nil, fmt.Errorf("error converting key into valid ecdsa key: %w", err)
			}
		} else {
			if _, err := os.Stat(options.KeyFile); err == nil {
				if prvKey, err = loadPrivateKeyFromFile(options.KeyFile); err != nil {
					return nil, fmt.Errorf("could not read keyfile: %w", err)
				}
			} else {
				if os.IsNotExist(err) {
					if prvKey, err = crypto.GenerateKey(); err != nil {
						return nil, fmt.Errorf("error generating key: %w", err)
					}
				} else {
					return nil, fmt.Errorf("could not read keyfile: %w", err)
				}
			}
		}
	}
	return prvKey, nil
}

func failOnErr(err error, msg string) {
	if err != nil {
		if msg != "" {
			msg = msg + ": "
		}
		utils.Logger().Fatal(msg, zap.Error(err))
	}
}

func freePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	port := l.Addr().(*net.TCPAddr).Port
	err = l.Close()
	if err != nil {
		return 0, err
	}

	return port, nil
}
