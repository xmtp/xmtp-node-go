package main

import (
	"log"
	"os"

	logging "github.com/ipfs/go-log"
	"github.com/jessevdk/go-flags"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/xmtp/xmtp-node-go/server"
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

	if options.GenerateKey {
		if err := server.WritePrivateKeyToFile(options.KeyFile, options.Overwrite); err != nil {
			log.Fatalf("Could not write private key file: %s", err)
		}
		return
	}

	if options.CreateMessageMigration != "" && options.Store.DbConnectionString != "" {
		if err := server.CreateMessageMigration(options.CreateMessageMigration, options.Store.DbConnectionString); err != nil {
			log.Fatalf("Could not create message migration: %s", err)
		}
		return
	}

	if options.CreateAuthzMigration != "" && options.Authz.DbConnectionString != "" {
		if err := server.CreateAuthzMigration(options.CreateAuthzMigration, options.Authz.DbConnectionString); err != nil {
			log.Fatalf("Could not create authz migration: %s", err)
		}
		return
	}

	server.New(options).WaitForShutdown()
}
