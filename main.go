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
		log.Fatal("Could not parse options", err)
	}

	addEnvVars()

	// for go-libp2p loggers
	lvl, err := logging.LevelFromString(options.LogLevel)
	if err != nil {
		log.Fatal("Invalid log level")
	}
	logging.SetAllLoggers(lvl)
	err = utils.SetLogLevel(options.LogLevel)
	if err != nil {
		log.Fatal("Failed to set log level")
	}

	if options.GenerateKey {
		if err := server.WritePrivateKeyToFile(options.KeyFile, options.Overwrite); err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	if options.CreateMessageMigration != "" && options.Store.DbConnectionString != "" {
		if err := server.CreateMessageMigration(options.CreateMessageMigration, options.Store.DbConnectionString); err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	if options.CreateAuthzMigration != "" && options.Authz.DbConnectionString != "" {
		if err := server.CreateAuthzMigration(options.CreateAuthzMigration, options.Authz.DbConnectionString); err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	s := server.New(options)
	s.WaitForShutdown()
}
