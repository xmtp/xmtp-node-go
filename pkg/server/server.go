package server

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/pkg/api"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"github.com/xmtp/xmtp-node-go/pkg/crdt"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	authzmigrations "github.com/xmtp/xmtp-node-go/pkg/migrations/authz"
	messagemigrations "github.com/xmtp/xmtp-node-go/pkg/migrations/messages"
	"go.uber.org/zap"
)

type Server struct {
	log         *zap.Logger
	db          *sql.DB
	metrics     *metrics.Server
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	allowLister authz.WalletAllowLister
	grpc        *api.Server
	crdt        *crdt.Node
}

// Create a new Server
func New(ctx context.Context, log *zap.Logger, options Options) (*Server, error) {
	s := &Server{
		log: log,
	}
	s.ctx, s.cancel = context.WithCancel(ctx)

	if options.MessageDBDSN != "" {
		var err error
		s.db, err = createDB(options.MessageDBDSN, options.WaitForDB)
		if err != nil {
			return nil, errors.Wrap(err, "creating db")
		}
		s.log.Info("created db")
	}

	if options.Metrics.Enable {
		var err error
		s.metrics, err = metrics.NewMetricsServer(s.ctx, s.log, options.Metrics.Address, options.Metrics.Port)
		if err != nil {
			return nil, errors.Wrap(err, "initializing metrics server")
		}
	}

	if options.Authz.DBDSN != "" {
		db, err := createBunDB(options.Authz.DBDSN, options.WaitForDB)
		if err != nil {
			return nil, errors.Wrap(err, "creating authz db")
		}
		s.allowLister = authz.NewDatabaseWalletAllowLister(db, s.log)
		err = s.allowLister.Start(s.ctx)
		if err != nil {
			return nil, errors.Wrap(err, "creating wallet authorizer")
		}
	}

	if options.NodeKey == "" {
		nodeKey := os.Getenv("XMTP_NODE_KEY")
		if nodeKey != "" {
			options.NodeKey = nodeKey
		}
	}

	// Initialize CRDT ds node.
	crdt, err := crdt.NewNode(ctx, log, crdt.Options{
		NodeKey:            options.NodeKey,
		DataPath:           options.DataPath,
		P2PPort:            options.P2PPort,
		P2PPersistentPeers: options.P2PPersistentPeers,
	})
	if err != nil {
		return nil, errors.Wrap(err, "initializing crdt")
	}
	s.crdt = crdt

	// Initialize gRPC server.
	s.grpc, err = api.New(
		&api.Config{
			Options:     options.API,
			Log:         s.log.Named("api"),
			AllowLister: s.allowLister,
			CRDT:        s.crdt,
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

	if s.allowLister != nil {
		s.allowLister.Stop()
	}

	// Close the DB.
	s.db.Close()

	if s.metrics != nil {
		if err := s.metrics.Stop(s.ctx); err != nil {
			s.log.Error("stopping metrics", zap.Error(err))
		}
	}

	// Close the gRPC s.
	if s.grpc != nil {
		s.grpc.Close()
	}

	// Close crdt.
	if s.crdt != nil {
		err := s.crdt.Close()
		if err != nil {
			s.log.Error("closing crdt", zap.Error(err))
		}
	}

	// Cancel outstanding goroutines
	s.cancel()
	s.wg.Wait()
	s.log.Info("shutdown complete")

}

func CreateMessageMigration(migrationName, dbDSN string, waitForDb time.Duration) error {
	db, err := createBunDB(dbDSN, waitForDb)
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

func CreateAuthzMigration(migrationName, dbDSN string, waitForDb time.Duration) error {
	db, err := createBunDB(dbDSN, waitForDb)
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
