package main

import (
	"log"
	"os"

	logging "github.com/ipfs/go-log"
	"github.com/jessevdk/go-flags"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/xmtp/xmtp-node-go/server"
	"github.com/xmtp/xmtp-node-go/tracing"
)

var options server.Options

var parser = flags.NewParser(&options, flags.Default)

// Avoiding replacing the flag parser with something fancier like Viper or urfave/cli
// This hack will work up to a point.
func addEnvVars() {
	if connStr, hasConnstr := os.LookupEnv("MESSAGE_DB_CONNECTION_STRING"); hasConnstr {
		options.Store.DbConnectionString = connStr
	}

	if connStr, hasConnstr := os.LookupEnv("AUTHZ_DB_CONNECTION_STRING"); hasConnstr {
		options.Authz.DbConnectionString = connStr
	}
}

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("Could not parse options: %s", err)
	}

	addEnvVars()

	// for go-libp2p loggers
	lvl, err := logging.LevelFromString(options.LogLevel)
	if err != nil {
		log.Fatalf("Could not parse log level: %s", err)
	}
	logging.SetAllLoggers(lvl)
	err = utils.SetLogLevel(options.LogLevel)
	if err != nil {
		log.Fatalf("Could not set log level: %s", err)
	}

	// Set encoding for logs (console, json, ...)
	// Note that libp2p reads the encoding from GOLOG_LOG_FMT env var.
	utils.InitLogger(options.LogEncoding)
	if options.LogEncoding == "json" && os.Getenv("GOLOG_LOG_FMT") == "" {
		utils.Logger().Warn("Set GOLOG_LOG_FMT=json to use json for libp2p logs")
	}

	if options.GenerateKey {
		if err := server.WritePrivateKeyToFile(options.KeyFile, options.Overwrite); err != nil {
			log.Fatalf("Could not write private key file: %s", err)
		}
		return
	}

	if options.CreateMessageMigration != "" && options.Store.DbConnectionString != "" {
		if err := server.CreateMessageMigration(options.CreateMessageMigration, options.Store.DbConnectionString, options.WaitForDB); err != nil {
			log.Fatalf("Could not create message migration: %s", err)
		}
		return
	}

	if options.CreateAuthzMigration != "" && options.Authz.DbConnectionString != "" {
		if err := server.CreateAuthzMigration(options.CreateAuthzMigration, options.Authz.DbConnectionString, options.WaitForDB); err != nil {
			log.Fatalf("Could not create authz migration: %s", err)
		}
		return
	}

	if options.Tracing.Enable {
		tracing.Start()
		defer tracing.Stop()
	}
	tracing.Do("main", func() {
		server.New(options).WaitForShutdown()
	})
}
