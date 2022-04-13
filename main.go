package main

import (
	"fmt"
	"log"
	"os"

	logging "github.com/ipfs/go-log"
	"github.com/jessevdk/go-flags"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/xmtp/xmtp-node-go/server"
)

var options server.Options

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatal("Could not parse options", err)
		os.Exit(1)
	}

	// for go-libp2p loggers
	lvl, err := logging.LevelFromString(options.LogLevel)
	if err != nil {
		os.Exit(1)
	}
	logging.SetAllLoggers(lvl)

	// go-waku logger
	fmt.Println(options.LogLevel)
	err = utils.SetLogLevel(options.LogLevel)
	if err != nil {
		os.Exit(1)
	}

	if options.GenerateKey {
		if err := server.WritePrivateKeyToFile(options.KeyFile, options.Overwrite); err != nil {
			log.Fatalf(err.Error())
		}
		return
	}

	if options.CreateMigration != "" && options.Store.DbConnectionString != "" {
		if err := server.CreateMigration(options.CreateMigration, options.Store.DbConnectionString); err != nil {
			log.Fatalf(err.Error())
		}
		return
	}

	s := server.New(options)
	log.Println("Got a server", s)
	s.WaitForShutdown()
}
