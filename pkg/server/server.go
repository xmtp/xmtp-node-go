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
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/pkg/api"
	"github.com/xmtp/xmtp-node-go/pkg/authn"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	authzmigrations "github.com/xmtp/xmtp-node-go/pkg/migrations/authz"
	messagemigrations "github.com/xmtp/xmtp-node-go/pkg/migrations/messages"
	"go.uber.org/zap"
)

type Server struct {
	log           *zap.Logger
	hostAddr      *net.TCPAddr
	db            *sql.DB
	readerDB      *sql.DB
	metricsServer *metrics.Server
	nats          *nats.Conn
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

	s.ctx, s.cancel = context.WithCancel(logging.With(ctx, s.log))

	s.db, err = createDB(options.Store.DbConnectionString, options.WaitForDB)
	if err != nil {
		return nil, errors.Wrap(err, "creating db")
	}
	s.log.Info("created db")

	s.readerDB, err = createDB(options.Store.DbReaderConnectionString, options.WaitForDB)
	if err != nil {
		return nil, errors.Wrap(err, "creating db")
	}

	if options.Metrics.Enable {
		s.metricsServer = metrics.NewMetricsServer(options.Metrics.Address, options.Metrics.Port, s.log)
		metrics.RegisterViews(s.log)
		s.metricsServer.Start(s.ctx)
	}

	if options.Authz.DbConnectionString != "" {
		db, err := createBunDB(options.Authz.DbConnectionString, options.WaitForDB)
		if err != nil {
			return nil, errors.Wrap(err, "creating authz db")
		}
		s.allowLister = authz.NewDatabaseWalletAllowLister(db, s.log)
		err = s.allowLister.Start(s.ctx)
		if err != nil {
			return nil, errors.Wrap(err, "creating wallet authorizer")
		}
	}

	// if options.Store.Enable {
	// 	nodeOpts = append(nodeOpts, node.WithWakuStoreAndRetentionPolicy(options.Store.ShouldResume, options.Store.RetentionMaxDaysDuration(), options.Store.RetentionMaxMessages))
	// 	dbStore, err := xmtpstore.NewDBStore(s.log, xmtpstore.WithDBStoreDB(s.db))
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "creating db store")
	// 	}
	// 	nodeOpts = append(nodeOpts, node.WithMessageProvider(dbStore))
	// 	// Not actually using the store just yet, as I would like to release this in chunks rather than have a monstrous PR.

	// 	nodeOpts = append(nodeOpts, node.WithWakuStoreFactory(func(w *node.WakuNode) store.Store {
	// 		store, err := xmtpstore.NewXmtpStore(
	// 			xmtpstore.WithLog(s.log),
	// 			xmtpstore.WithHost(w.Host()),
	// 			xmtpstore.WithDB(s.db),
	// 			xmtpstore.WithReaderDB(s.readerDB),
	// 			xmtpstore.WithCleaner(options.Cleaner),
	// 			xmtpstore.WithMessageProvider(dbStore),
	// 			xmtpstore.WithStatsPeriod(options.Metrics.StatusPeriod),
	// 			xmtpstore.WithResumeStartTime(options.Store.ResumeStartTime),
	// 		)
	// 		if err != nil {
	// 			s.log.Fatal("initializing store", zap.Error(err))
	// 		}
	// 		return store
	// 	}))
	// }

	s.nats, err = nats.Connect(options.NATSURL)
	if err != nil {
		return nil, errors.Wrap(err, "initializing nats connection")
	}

	s.authenticator = authn.NewXmtpAuthentication(s.ctx, s.log)
	s.authenticator.Start()

	// Initialize gRPC server.
	s.grpc, err = api.New(
		&api.Config{
			Options:     options.API,
			Log:         s.log.Named("api"),
			NATS:        s.nats,
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

	// close the nats connection
	s.nats.Close()

	if s.allowLister != nil {
		s.allowLister.Stop()
	}

	// Close the DB.
	s.db.Close()

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
		if prvKey, err = hexToECDSA(options.NodeKey); err != nil {
			return nil, fmt.Errorf("error converting key into valid ecdsa key: %w", err)
		}
	} else {
		keyString := os.Getenv("GOWAKU-NODEKEY")
		if keyString != "" {
			if prvKey, err = hexToECDSA(keyString); err != nil {
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

func CreateMessageMigration(migrationName, dbConnectionString string, waitForDb time.Duration) error {
	db, err := createBunDB(dbConnectionString, waitForDb)
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

func CreateAuthzMigration(migrationName, dbConnectionString string, waitForDb time.Duration) error {
	db, err := createBunDB(dbConnectionString, waitForDb)
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

func createBunDB(dsn string, waitForDB time.Duration) (*bun.DB, error) {
	db, err := createDB(dsn, waitForDB)
	if err != nil {
		return nil, err
	}
	return bun.NewDB(db, pgdialect.New()), nil
}

func createDB(dsn string, waitForDB time.Duration) (*sql.DB, error) {
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
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

func hexToECDSA(key string) (*ecdsa.PrivateKey, error) {
	if len(key) == 60 {
		key = "0000" + key
	}
	if len(key) == 62 {
		key = "00" + key
	}
	return ethcrypto.HexToECDSA(key)
}
