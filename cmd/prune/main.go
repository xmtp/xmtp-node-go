package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/xmtp/xmtp-node-go/pkg/prune"
	"github.com/xmtp/xmtp-node-go/pkg/server"
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

	addEnvVars()

	if options.Version {
		fmt.Printf("Version: %s", Commit)
		return
	}

	err := ValidatePruneOptions(options)
	if err != nil {
		fatal("Invalid prune options: %s", err)
	}

	ctx := context.Background()

	logger, _, err := server.BuildLogger(options.Log, "prune")
	if err != nil {
		fatal("Could not build logger: %s", err)
	}

	dbInstance, err := server.NewPGXDB(
		options.MLSStore.DbConnectionString,
		30*time.Second,
		options.MLSStore.ReadTimeout,
	)
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

func addEnvVars() {
	if connStr, hasConnstr := os.LookupEnv("MLS_DB_CONNECTION_STRING"); hasConnstr {
		options.MLSStore.DbConnectionString = connStr
	}
}

func fatal(msg string, args ...any) {
	log.Fatalf(msg, args...)
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
