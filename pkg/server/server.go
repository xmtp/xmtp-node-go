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

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/xmtp-node-go/pkg/api"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	authzmigrations "github.com/xmtp/xmtp-node-go/pkg/migrations/authz"
	messagemigrations "github.com/xmtp/xmtp-node-go/pkg/migrations/messages"
	xmtpstore "github.com/xmtp/xmtp-node-go/pkg/store"
	"go.uber.org/zap"
)

type Server struct {
	log           *zap.Logger
	nats          *nats.Conn
	store         *xmtpstore.Store
	db            *sql.DB
	readerDB      *sql.DB
	cleanerDB     *sql.DB
	metricsServer *metrics.Server
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	allowLister   authz.WalletAllowLister
	grpc          *api.Server
}

// Create a new Server
func New(ctx context.Context, log *zap.Logger, options Options) (*Server, error) {
	s := &Server{
		log: log,
	}

	var err error

	natsURL := options.NATS.URL
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	s.nats, err = nats.Connect(natsURL)
	if err != nil {
		return nil, errors.Wrap(err, "initializing nats")
	}

	s.ctx, s.cancel = context.WithCancel(logging.With(ctx, s.log))

	s.db, err = createDB(options.Store.DbConnectionString, options.WaitForDB, options.Store.ReadTimeout, options.Store.WriteTimeout, options.Store.MaxOpenConns)
	if err != nil {
		return nil, errors.Wrap(err, "creating db")
	}
	s.log.Info("created db")

	s.readerDB, err = createDB(options.Store.DbReaderConnectionString, options.WaitForDB, options.Store.ReadTimeout, options.Store.WriteTimeout, options.Store.MaxOpenConns)
	if err != nil {
		return nil, errors.Wrap(err, "creating reader db")
	}

	s.cleanerDB, err = createDB(options.Store.DbConnectionString, options.WaitForDB, options.Cleaner.ReadTimeout, options.Cleaner.WriteTimeout, options.Store.MaxOpenConns)
	if err != nil {
		return nil, errors.Wrap(err, "creating cleaner db")
	}

	if options.Metrics.Enable {
		s.metricsServer = metrics.NewMetricsServer()
		metrics.RegisterViews(s.log)
		s.metricsServer.Start(s.ctx)
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

	// Initialize store.
	s.store, err = xmtpstore.New(
		xmtpstore.WithLog(s.log),
		xmtpstore.WithDB(s.db),
		xmtpstore.WithReaderDB(s.readerDB),
		xmtpstore.WithCleanerDB(s.cleanerDB),
		xmtpstore.WithCleaner(options.Cleaner),
		xmtpstore.WithStatsPeriod(options.Metrics.StatusPeriod),
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing store")
	}

	// Initialize gRPC server.
	s.grpc, err = api.New(
		&api.Config{
			Options:     options.API,
			Log:         s.log.Named("api"),
			NATS:        s.nats,
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

	if s.nats != nil {
		s.nats.Close()
	}
	if s.store != nil {
		s.store.Close()
	}

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
