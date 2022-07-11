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

	"github.com/xmtp/xmtp-node-go/authn"
	"github.com/xmtp/xmtp-node-go/tracing"

	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/zap"

	libp2p "github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
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
	"github.com/xmtp/xmtp-node-go/metrics"
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
	cancel           context.CancelFunc
	walletAuthorizer authz.WalletAuthorizer
	authenticator    *authn.XmtpAuthentication
}

// Create a new Server
func New(ctx context.Context, options Options) (server *Server) {
	server = new(Server)
	var err error

	server.logger = utils.Logger()
	server.hostAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", options.Address, options.Port))
	failOnErr(err, "invalid host address")
	server.logger.Info("resolved host addr", zap.Stringer("addr", server.hostAddr))

	prvKey, err := getPrivKey(options)
	failOnErr(err, "nodekey error")

	p2pPrvKey := utils.EcdsaPrivKeyToSecp256k1PrivKey(prvKey)
	id, err := peer.IDFromPublicKey(p2pPrvKey.GetPublic())
	failOnErr(err, "deriving peer ID from private key")
	server.logger = server.logger.With(logging.HostID("node", id))
	server.ctx, server.cancel = context.WithCancel(logging.With(ctx, server.logger))

	server.db = createDbOrFail(options.Store.DbConnectionString, options.WaitForDB)
	server.logger.Info("created DB")

	if options.Metrics.Enable {
		server.metricsServer = metrics.NewMetricsServer(options.Metrics.Address, options.Metrics.Port, server.logger)
		metrics.RegisterViews(server.logger)
		go tracing.Do(server.ctx, "metrics", func(_ context.Context) { server.metricsServer.Start() })
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
	libp2pOpts = append(libp2pOpts, libp2p.NATPortMap()) // Attempt to open ports using uPNP for NATed hosts.

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
			return xmtpStore.NewXmtpStore(w.Host(), server.db, dbStore, options.Metrics.StatusPeriod, server.logger)
		}))
	}

	if options.LightPush.Enable {
		nodeOpts = append(nodeOpts, node.WithLightPush())
	}

	server.wakuNode, err = node.New(server.ctx, nodeOpts...)
	failOnErr(err, "Wakunode")

	if options.Metrics.Enable {
		go tracing.Do(server.ctx, "status metrics", func(_ context.Context) { server.statusMetricsLoop(options) })
	}

	addPeers(server.wakuNode, options.Store.Nodes, string(store.StoreID_v20beta4))
	addPeers(server.wakuNode, options.LightPush.Nodes, string(lightpush.LightPushID_v20beta1))
	addPeers(server.wakuNode, options.Filter.Nodes, string(filter.FilterID_v20beta1))

	if err = server.wakuNode.Start(); err != nil {
		server.logger.Fatal(fmt.Errorf("could not start waku node, %w", err).Error())
	}

	server.authenticator = authn.NewXmtpAuthentication(server.ctx, server.wakuNode.Host(), server.logger)
	server.authenticator.Start()

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

	go tracing.Do(server.ctx, "static-nodes-connect-loop", func(_ context.Context) { server.staticNodesConnectLoop(options.StaticNodes) })

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
	server.Shutdown()
}

func (server *Server) Shutdown() {
	server.logger.Info("shutting down...")
	server.cancel()

	// shut the node down
	server.wakuNode.Stop()

	// Close the DB.
	server.db.Close()

	if server.metricsServer != nil {
		err := server.metricsServer.Stop(server.ctx)
		failOnErr(err, "metrics stop")
	}
}

func (server *Server) staticNodesConnectLoop(staticNodes []string) {
	dialPeer := func(peerAddr string) {
		err := server.wakuNode.DialPeer(server.ctx, peerAddr)
		if err != nil {
			server.logger.Error("dialing static node", zap.Error(err), zap.String("peer_addr", peerAddr))
		}
	}

	for _, peerAddr := range staticNodes {
		dialPeer(peerAddr)
	}

	staticNodePeerIDs := make([]peer.ID, len(staticNodes))
	for i, peerAddr := range staticNodes {
		ma, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			server.logger.Error("building multiaddr from static node addr", zap.Error(err))
		}
		pi, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			server.logger.Error("getting peer addr info from static node addr", zap.Error(err))
		}
		staticNodePeerIDs[i] = pi.ID
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-server.ctx.Done():
			return
		case <-ticker.C:
			peers := map[peer.ID]struct{}{}
			for _, peerID := range server.wakuNode.Host().Network().Peers() {
				peers[peerID] = struct{}{}
			}
			for i, peerAddr := range staticNodes {
				peerID := staticNodePeerIDs[i]
				if _, exists := peers[peerID]; exists {
					continue
				}
				dialPeer(peerAddr)
			}
		}
	}
}

func (server *Server) statusMetricsLoop(options Options) {
	server.logger.Info("starting status metrics loop", zap.Duration("period", options.Metrics.StatusPeriod))
	ticker := time.NewTicker(options.Metrics.StatusPeriod)
	bootstrapPeers := map[peer.ID]bool{}
	for _, addr := range options.StaticNodes {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			server.logger.Error("parsing static node multiaddr", zap.String("addr", addr), zap.Error(err))
			continue
		}
		_, pid := peer.SplitAddr(maddr)
		bootstrapPeers[pid] = true
	}
	defer ticker.Stop()
	for {
		select {
		case <-server.ctx.Done():
			return
		case <-ticker.C:
			metrics.EmitPeersByProtocol(server.ctx, server.wakuNode.Host())
			if len(bootstrapPeers) > 0 {
				metrics.EmitBootstrapPeersConnected(server.ctx, server.wakuNode.Host(), bootstrapPeers)
			}
		}
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

	pBytes, err := p.Raw()
	if err != nil {
		return nil, err
	}

	return crypto.ToECDSA(pBytes)
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

	b := key.D.Bytes()

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
