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
	"sync"
	"syscall"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	libp2p "github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/status-im/go-waku/waku/v2/node"
	"github.com/status-im/go-waku/waku/v2/protocol/filter"
	"github.com/status-im/go-waku/waku/v2/protocol/lightpush"
	"github.com/status-im/go-waku/waku/v2/protocol/relay"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/pkg/api"
	"github.com/xmtp/xmtp-node-go/pkg/authn"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"github.com/xmtp/xmtp-node-go/pkg/crypto"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	authzmigrations "github.com/xmtp/xmtp-node-go/pkg/migrations/authz"
	messagemigrations "github.com/xmtp/xmtp-node-go/pkg/migrations/messages"
	xmtpstore "github.com/xmtp/xmtp-node-go/pkg/store"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
)

type Server struct {
	log           *zap.Logger
	hostAddr      *net.TCPAddr
	store         *xmtpstore.Store
	db            *sql.DB
	readerDB      *sql.DB
	cleanerDB     *sql.DB
	metricsServer *metrics.Server
	wakuNode      *node.WakuNode
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	allowLister   authz.WalletAllowLister
	authenticator *authn.XmtpAuthentication
	grpc          *api.Server
}

// Create a new Server
func New(ctx context.Context, log *zap.Logger, options Options) (*Server, error) {
	s := &Server{
		log: log,
	}

	var err error
	s.hostAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", options.Address, options.Port))
	if err != nil {
		return nil, errors.Wrap(err, "resolving host address")
	}
	s.log.Info("resolved host addr", zap.Stringer("addr", s.hostAddr))

	prvKey, err := getPrivKey(options)
	if err != nil {
		return nil, errors.Wrap(err, "getting private key")
	}

	p2pPrvKey := utils.EcdsaPrivKeyToSecp256k1PrivKey(prvKey)
	id, err := peer.IDFromPublicKey(p2pPrvKey.GetPublic())
	if err != nil {
		return nil, errors.Wrap(err, "deriving peer id from private key")
	}
	s.log = s.log.With(logging.HostID("node", id))

	s.ctx, s.cancel = context.WithCancel(logging.With(ctx, s.log))

	if options.Metrics.Enable {
		s.metricsServer = metrics.NewMetricsServer(options.Metrics.Address, options.Metrics.Port, s.log)
		metrics.RegisterViews(s.log)
		s.metricsServer.Start(s.ctx)
	}

	nodeOpts := []node.WakuNodeOption{
		node.WithLogger(s.log.Named("gowaku")),
		node.WithPrivateKey(prvKey),
		node.WithHostAddress(s.hostAddr),
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
		directPeers := make([]peer.AddrInfo, 0, len(options.StaticNodes))
		for _, staticNode := range options.StaticNodes {
			ma, err := multiaddr.NewMultiaddr(staticNode)
			if err != nil {
				s.log.Error("building multiaddr from static node addr", zap.Error(err))
				continue
			}
			pi, err := peer.AddrInfoFromP2pAddr(ma)
			if err != nil {
				s.log.Error("getting peer addr info from static node addr", zap.Error(err))
				continue
			}
			if pi == nil {
				s.log.Error("static node peer addr is nil", zap.String("peer", staticNode))
				continue
			}
			directPeers = append(directPeers, *pi)
		}
		wakurelayopts = append(wakurelayopts, pubsub.WithPeerExchange(true), pubsub.WithDirectPeers(directPeers))
		nodeOpts = append(nodeOpts, node.WithWakuRelayAndMinPeers(options.Relay.MinRelayPeersToPublish, wakurelayopts...))
	}

	if options.Filter.Enable {
		nodeOpts = append(nodeOpts, node.WithWakuFilter(true, filter.WithTimeout(time.Duration(options.Filter.Timeout)*time.Second)))
	}

	if options.Authz.DbConnectionString != "" {
		db, err := createBunDB(options.Authz.DbConnectionString, options.WaitForDB, options.Authz.ReadTimeout, options.Authz.WriteTimeout, options.Store.MaxOpenConns)
		if err != nil {
			return nil, errors.Wrap(err, "creating authz db")
		}
		s.allowLister = authz.NewDatabaseWalletAllowLister(db, s.log)
		err = s.allowLister.Start(s.ctx)
		if err != nil {
			return nil, errors.Wrap(err, "creating wallet authorizer")
		}
	}

	if options.Store.Enable {
		s.db, err = createDB(options.Store.DbConnectionString, options.WaitForDB, options.Store.ReadTimeout, options.Store.WriteTimeout, options.Store.MaxOpenConns)
		if err != nil {
			return nil, errors.Wrap(err, "creating db")
		}
		s.log.Info("created db")

		s.readerDB, err = createDB(options.Store.DbReaderConnectionString, options.WaitForDB, options.Store.ReadTimeout, options.Store.WriteTimeout, options.Store.MaxOpenConns)
		if err != nil {
			return nil, errors.Wrap(err, "creating reader db")
		}

		s.cleanerDB, err = createDB(options.Store.DbConnectionString, options.WaitForDB, options.Store.Cleaner.ReadTimeout, options.Store.Cleaner.WriteTimeout, options.Store.MaxOpenConns)
		if err != nil {
			return nil, errors.Wrap(err, "creating cleaner db")
		}

		s.store, err = xmtpstore.New(&xmtpstore.Config{
			Options:   options.Store,
			Log:       log,
			DB:        s.db,
			ReaderDB:  s.readerDB,
			CleanerDB: s.cleanerDB,
		})
		if err != nil {
			return nil, errors.Wrap(err, "initializing store")
		}
	}

	if options.LightPush.Enable {
		nodeOpts = append(nodeOpts, node.WithLightPush())
	}

	s.wakuNode, err = node.New(s.ctx, nodeOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "initializing waku node")
	}

	if options.Metrics.Enable {
		tracing.GoPanicWrap(s.ctx, &s.wg, "status metrics", func(_ context.Context) { s.statusMetricsLoop(options) })
	}

	err = addPeers(s.wakuNode, options.LightPush.Nodes, string(lightpush.LightPushID_v20beta1))
	if err != nil {
		return nil, errors.Wrap(err, "adding peer")
	}
	err = addPeers(s.wakuNode, options.Filter.Nodes, string(filter.FilterID_v20beta1))
	if err != nil {
		return nil, errors.Wrap(err, "adding peer")
	}

	if err = s.wakuNode.Start(); err != nil {
		s.log.Fatal(fmt.Errorf("could not start waku node, %w", err).Error())
	}

	s.authenticator = authn.NewXmtpAuthentication(s.ctx, s.wakuNode.Host(), s.log)
	s.authenticator.Start()

	if len(options.Relay.Topics) == 0 {
		options.Relay.Topics = []string{string(relay.DefaultWakuTopic)}
	}

	if !options.Relay.Disable {
		for _, nodeTopic := range options.Relay.Topics {
			nodeTopic := nodeTopic
			sub, err := s.wakuNode.Relay().SubscribeToTopic(s.ctx, nodeTopic)
			if err != nil {
				return nil, errors.Wrap(err, "subscribing to pubsub topic")
			}
			// Unregister from broadcaster. Otherwise this channel will fill until it blocks publishing
			s.wakuNode.Broadcaster().Unregister(&nodeTopic, sub.C)
		}
	}

	tracing.GoPanicWrap(s.ctx, &s.wg, "static-nodes-connect-loop", func(_ context.Context) { s.staticNodesConnectLoop(options.StaticNodes) })

	maddrs, err := s.wakuNode.Host().Network().InterfaceListenAddresses()
	if err != nil {
		return nil, errors.Wrap(err, "getting listen addresses")
	}
	s.log.With(logging.MultiAddrs("listen", maddrs...)).Info("got server")

	// Initialize gRPC server.
	s.grpc, err = api.New(
		&api.Config{
			Options:     options.API,
			Log:         s.log.Named("api"),
			Waku:        s.wakuNode,
			Store:       s.store,
			AllowLister: s.allowLister,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing grpc server")
	}
	return s, nil
}

func (s *Server) WaitForShutdown() {
	// Wait for a SIGINT or SIGTERM signal
	termChannel := make(chan os.Signal, 1)
	signal.Notify(termChannel, syscall.SIGINT, syscall.SIGTERM)
	<-termChannel
	s.Shutdown()
}

func (s *Server) Shutdown() {
	s.log.Info("shutting down...")

	// Close waku node.
	s.wakuNode.Stop()

	// Close allow lister.
	if s.allowLister != nil {
		s.allowLister.Stop()
	}

	// Close the DBs and store.
	if s.db != nil {
		s.db.Close()
	}
	if s.readerDB != nil {
		s.readerDB.Close()
	}
	if s.cleanerDB != nil {
		s.cleanerDB.Close()
	}
	if s.store != nil {
		s.store.Close()
	}

	// Close metrics server.
	if s.metricsServer != nil {
		if err := s.metricsServer.Stop(s.ctx); err != nil {
			s.log.Error("stopping metrics", zap.Error(err))
		}
	}

	// Close the gRPC s.
	if s.grpc != nil {
		s.grpc.Close()
	}

	// Cancel outstanding goroutines
	s.cancel()
	s.wg.Wait()
	s.log.Info("shutdown complete")

}

func (s *Server) staticNodesConnectLoop(staticNodes []string) {
	dialPeer := func(peerAddr string) {
		err := s.wakuNode.DialPeer(s.ctx, peerAddr)
		if err != nil {
			s.log.Error("dialing static node", zap.Error(err), zap.String("peer_addr", peerAddr))
		}
	}

	for _, peerAddr := range staticNodes {
		dialPeer(peerAddr)
	}

	staticNodePeerIDs := make([]peer.ID, len(staticNodes))
	for i, peerAddr := range staticNodes {
		ma, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			s.log.Error("building multiaddr from static node addr", zap.Error(err))
		}
		pi, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			s.log.Error("getting peer addr info from static node addr", zap.Error(err))
		}
		staticNodePeerIDs[i] = pi.ID
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			peers := map[peer.ID]struct{}{}
			for _, peerID := range s.wakuNode.Host().Network().Peers() {
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

func (s *Server) statusMetricsLoop(options Options) {
	s.log.Info("starting status metrics loop", zap.Duration("period", options.MetricsPeriod))
	ticker := time.NewTicker(options.MetricsPeriod)
	bootstrapPeers := map[peer.ID]bool{}
	for _, addr := range options.StaticNodes {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			s.log.Error("parsing static node multiaddr", zap.String("addr", addr), zap.Error(err))
			continue
		}
		_, pid := peer.SplitAddr(maddr)
		bootstrapPeers[pid] = true
	}
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			metrics.EmitPeersByProtocol(s.ctx, s.wakuNode.Host())
			if len(bootstrapPeers) > 0 {
				metrics.EmitBootstrapPeersConnected(s.ctx, s.wakuNode.Host(), bootstrapPeers)
			}
		}
	}
}

func addPeers(wakuNode *node.WakuNode, addresses []string, protocols ...string) error {
	for _, addrString := range addresses {
		if addrString == "" {
			continue
		}

		addr, err := multiaddr.NewMultiaddr(addrString)
		if err != nil {
			return errors.Wrap(err, "invalid multiaddress")
		}

		_, err = wakuNode.AddPeer(addr, protocols...)
		if err != nil {
			return err
		}
	}
	return nil
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

	return ethcrypto.ToECDSA(pBytes)
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
	key, err := ethcrypto.GenerateKey()
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
					if prvKey, err = ethcrypto.GenerateKey(); err != nil {
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

func CreateMessageMigration(migrationName, dbConnectionString string, waitForDb, readTimeout, writeTimeout time.Duration, maxOpenConns int) error {
	db, err := createBunDB(dbConnectionString, waitForDb, readTimeout, writeTimeout, maxOpenConns)
	if err != nil {
		return err
	}
	migrator := migrate.NewMigrator(db, messagemigrations.Migrations)
	files, err := migrator.CreateSQLMigrations(context.Background(), migrationName)
	for _, mf := range files {
		fmt.Printf("created message migration %s (%s)\n", mf.Name, mf.Path)
	}

	return err
}

func CreateAuthzMigration(migrationName, dbConnectionString string, waitForDb, readTimeout, writeTimeout time.Duration, maxOpenConns int) error {
	db, err := createBunDB(dbConnectionString, waitForDb, readTimeout, writeTimeout, maxOpenConns)
	if err != nil {
		return err
	}
	migrator := migrate.NewMigrator(db, authzmigrations.Migrations)
	files, err := migrator.CreateSQLMigrations(context.Background(), migrationName)
	for _, mf := range files {
		fmt.Printf("created authz migration %s (%s)\n", mf.Name, mf.Path)
	}

	return err
}

func createBunDB(dsn string, waitForDB, readTimeout, writeTimeout time.Duration, maxOpenConns int) (*bun.DB, error) {
	db, err := createDB(dsn, waitForDB, readTimeout, writeTimeout, maxOpenConns)
	if err != nil {
		return nil, err
	}
	return bun.NewDB(db, pgdialect.New()), nil
}

func createDB(dsn string, waitForDB, readTimeout, writeTimeout time.Duration, maxOpenConns int) (*sql.DB, error) {
	db := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(dsn),
		pgdriver.WithReadTimeout(readTimeout),
		pgdriver.WithWriteTimeout(writeTimeout),
	))

	if maxOpenConns > 0 {
		db.SetMaxOpenConns(maxOpenConns)
	}

	waitUntil := time.Now().Add(waitForDB)
	err := db.Ping()
	for err != nil && time.Now().Before(waitUntil) {
		time.Sleep(3 * time.Second)
		err = db.Ping()
	}
	if err != nil {
		return nil, errors.New("timeout waiting for db")
	}
	return db, nil
}
