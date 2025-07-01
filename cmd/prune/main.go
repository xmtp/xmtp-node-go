package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"log"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/xmtp/xmtp-node-go/pkg/prune"
	"github.com/xmtp/xmtp-node-go/pkg/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "net/http/pprof"
)

var Commit string

var options server.PruneOptions

func main() {
	if _, err := flags.Parse(&options); err != nil {
		if err, ok := err.(*flags.Error); !ok || err.Type != flags.ErrHelp {
			fatal("Could not parse options: %s", err)
		}
		return
	}

	err := ValidatePruneOptions(options)
	if err != nil {
		fatal("Invalid prune options: %s", err)
	}

	ctx := context.Background()

	logger, _, err := buildLogger(options.LogEncoding)
	if err != nil {
		fatal("Could not build logger: %s", err)
	}

	dbInstance, err := newPGXDB(options.MLSStore.DbConnectionString, 30*time.Second, options.MLSStore.ReadTimeout)
	if err != nil {
		fatal("Could not open db: %s", err)
	}
	logger.Info("opened db")

	pruneExecutor := prune.NewPruneExecutor(ctx, logger, dbInstance, &options.PruneConfig)

	err = pruneExecutor.Run()
	if err != nil {
		fatal("Could not execute prune: %s", err)
	}
}

func fatal(msg string, args ...any) {
	log.Fatalf(msg, args...)
}

func buildLogger(encoding string) (*zap.Logger, *zap.Config, error) {
	atom := zap.NewAtomicLevel()
	level := zapcore.InfoLevel
	err := level.Set(options.LogLevel)
	if err != nil {
		return nil, nil, err
	}
	atom.SetLevel(level)

	cfg := zap.Config{
		Encoding:         encoding,
		Level:            atom,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			NameKey:      "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	log, err := cfg.Build()
	if err != nil {
		return nil, nil, err
	}

	log = log.Named("prune")

	return log, &cfg, nil
}

func newPGXDB(dsn string, waitForDB, statementTimeout time.Duration) (*sql.DB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprint(statementTimeout.Milliseconds())

	dbpool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDBFromPool(dbpool)

	waitUntil := time.Now().Add(waitForDB)

	err = db.Ping()
	for err != nil && time.Now().Before(waitUntil) {
		time.Sleep(3 * time.Second)
		err = db.Ping()
	}

	return db, nil
}

func ValidatePruneOptions(options server.PruneOptions) error {
	missingSet := make(map[string]struct{})

	if options.MLSStore.DbConnectionString == "" {
		missingSet["--mls-store.db-connection-string"] = struct{}{}
	}

	if len(missingSet) > 0 {
		var errorMessages []string
		for err := range missingSet {
			errorMessages = append(errorMessages, err)
		}

		return fmt.Errorf("missing required arguments: %s", strings.Join(errorMessages, ", "))
	}

	if options.PruneConfig.MaxCycles < 1 {
		return fmt.Errorf("max-cycles must be greater than 0")
	}

	return nil
}
