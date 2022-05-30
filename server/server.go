package server

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/hex"
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

	libp2p "github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/status-im/go-waku/waku/metrics"
	"github.com/status-im/go-waku/waku/persistence/sqlite"
	"github.com/status-im/go-waku/waku/v2/node"
	"github.com/status-im/go-waku/waku/v2/protocol/filter"
	"github.com/status-im/go-waku/waku/v2/protocol/lightpush"
	"github.com/status-im/go-waku/waku/v2/protocol/relay"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/authz"
	"github.com/xmtp/xmtp-node-go/logging"
	authzMigrations "github.com/xmtp/xmtp-node-go/migrations/authz"
	messageMigrations "github.com/xmtp/xmtp-node-go/migrations/messages"
	xmtpStore "github.com/xmtp/xmtp-node-go/store"
)

type Server struct {
	logger           *zap.Logger
	hostAddr         *net.TCPAddr
	db               *sql.DB
	metricsServer    *metrics.Server
	wakuNode         *node.WakuNode
	ctx              context.Context
	walletAuthorizer authz.WalletAuthorizer
}

// Create a new Server
func New(options Options) (server *Server) {
	server = new(Server)
	var err error

	server.logger = utils.InitLogger(options.LogEncoding)
	if options.LogEncoding == "json" && os.Getenv("GOLOG_LOG_FMT") == "" {
		server.logger.Warn("Set GOLOG_LOG_FMT=json to use json for libp2p logs")
	}

	server.hostAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", options.Address, options.Port))
	failOnErr(err, "invalid host address")

	prvKey, err := getPrivKey(options)
	failOnErr(err, "nodekey error")

	p2pPrvKey := libp2pcrypto.Secp256k1PrivateKey(*prvKey)
	id, err := peer.IDFromPublicKey((&p2pPrvKey).GetPublic())
	failOnErr(err, "deriving peer ID from private key")
	server.logger = server.logger.With(logging.HostID("node", id))

	server.db = createDbOrFail(options.Store.DbConnectionString, options.WaitForDB)
	server.ctx = context.Background()

	if options.Metrics.Enable {
		server.metricsServer = metrics.NewMetricsServer(options.Metrics.Address, options.Metrics.Port, server.logger)
		go server.metricsServer.Start()
	}

	if options.Authz.DbConnectionString != "" {
		db := createBunDbOrFail(options.Authz.DbConnectionString, options.WaitForDB)
		server.walletAuthorizer = authz.NewDatabaseWalletAuthorizer(db, server.logger)
		err = server.walletAuthorizer.Start(server.ctx)
		failOnErr(err, "wallet authorizer error")
	}

	nodeOpts := []node.WakuNodeOption{
		node.WithLogger(server.logger),
		node.WithPrivateKey(prvKey),
		node.WithHostAddress(server.hostAddr),
		node.WithKeepAlive(time.Duration(options.KeepAlive) * time.Second),
	}

	if options.EnableWS {
		nodeOpts = append(nodeOpts, node.WithWebsockets(options.WSAddress, options.WSPort))
	}

	libp2pOpts := node.DefaultLibP2POptions
	libp2pOpts = append(libp2pOpts, libp2p.NATPortMap()) // Attempt to open ports using uPNP for NATed hosts.)

	// Create persistent peerstore
	// This uses the "sqlite" package from go-waku, but with the Postgres powered store. The queries all work regardless of underlying datastore
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
		dbStore, err := xmtpStore.NewDBStore(server.logger, xmtpStore.WithDB(server.db))
		failOnErr(err, "DBStore")
		nodeOpts = append(nodeOpts, node.WithMessageProvider(dbStore))
		// Not actually using the store just yet, as I would like to release this in chunks rather than have a monstrous PR.

		nodeOpts = append(nodeOpts, node.WithWakuStoreFactory(func(w *node.WakuNode) store.Store {
			return xmtpStore.NewXmtpStore(w.Host(), server.db, dbStore, options.Store.RetentionMaxDaysDuration(), server.logger)
		}))
	}

	if options.LightPush.Enable {
		nodeOpts = append(nodeOpts, node.WithLightPush())
	}

	server.wakuNode, err = node.New(server.ctx, nodeOpts...)
	failOnErr(err, "Wakunode")

	addPeers(server.wakuNode, options.Store.Nodes, string(store.StoreID_v20beta4))
	addPeers(server.wakuNode, options.LightPush.Nodes, string(lightpush.LightPushID_v20beta1))
	addPeers(server.wakuNode, options.Filter.Nodes, string(filter.FilterID_v20beta1))

	if err = server.wakuNode.Start(); err != nil {
		server.logger.Fatal(fmt.Errorf("could not start waku node, %w", err).Error())
	}

	if len(options.Relay.Topics) == 0 {
		options.Relay.Topics = []string{string(relay.DefaultWakuTopic)}
	}

	if !options.Relay.Disable {
		for _, nodeTopic := range options.Relay.Topics {
			nodeTopic := nodeTopic
			sub, err := server.wakuNode.Relay().SubscribeToTopic(server.ctx, nodeTopic)
			failOnErr(err, "Error subscring to topic")
			// Unregister from broadcaster. Otherwise this channel will fill until it blocks publishing
			server.wakuNode.Broadcaster().Unregister(&nodeTopic, sub.C)
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

	maddrs, err := server.wakuNode.Host().Network().InterfaceListenAddresses()
	failOnErr(err, "getting listen addresses")
	server.logger.With(logging.MultiAddrs("listen", maddrs...)).Info("got server")
	return server
}

func (server *Server) WaitForShutdown() {
	// Wait for a SIGINT or SIGTERM signal
	termChannel := make(chan os.Signal, 1)
	signal.Notify(termChannel, syscall.SIGINT, syscall.SIGTERM)
	<-termChannel
	server.logger.Info("shutting down...")

	// shut the node down
	server.wakuNode.Stop()

	if server.metricsServer != nil {
		err := server.metricsServer.Stop(server.ctx)
		failOnErr(err, "metrics stop")
	}
}

func addPeers(wakuNode *node.WakuNode, addresses []string, protocols ...string) {
	for _, addrString := range addresses {
		if addrString == "" {
			continue
		}

		addr, err := multiaddr.NewMultiaddr(addrString)
		failOnErr(err, "invalid multiaddress")

		_, err = wakuNode.AddPeer(addr, protocols...)
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

func CreateMessageMigration(migrationName, dbConnectionString string, waitForDb time.Duration) error {
	db := createBunDbOrFail(dbConnectionString, waitForDb)
	migrator := migrate.NewMigrator(db, messageMigrations.Migrations)
	files, err := migrator.CreateSQLMigrations(context.Background(), migrationName)
	for _, mf := range files {
		fmt.Printf("created message migration %s (%s)\n", mf.Name, mf.Path)
	}

	return err
}

func CreateAuthzMigration(migrationName, dbConnectionString string, waitForDb time.Duration) error {
	db := createBunDbOrFail(dbConnectionString, waitForDb)
	migrator := migrate.NewMigrator(db, authzMigrations.Migrations)
	files, err := migrator.CreateSQLMigrations(context.Background(), migrationName)
	for _, mf := range files {
		fmt.Printf("created authz migration %s (%s)\n", mf.Name, mf.Path)
	}

	return err
}

func failOnErr(err error, msg string) {
	if err != nil {
		if msg != "" {
			msg = msg + ": "
		}
		utils.Logger().Fatal(msg, zap.Error(err))
	}
}

func createBunDbOrFail(dsn string, waitForDB time.Duration) *bun.DB {
	return bun.NewDB(createDbOrFail(dsn, waitForDB), pgdialect.New())
}

func createDbOrFail(dsn string, waitForDB time.Duration) *sql.DB {
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	waitUntil := time.Now().Add(waitForDB)
	err := db.Ping()
	for err != nil && time.Now().Before(waitUntil) {
		time.Sleep(3 * time.Second)
		err = db.Ping()
	}
	if err != nil {
		failOnErr(err, "timeout waiting for DB")
	}
	return db
}
