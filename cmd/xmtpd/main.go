package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/xmtp/xmtp-node-go/pkg/crdt"
	"github.com/xmtp/xmtp-node-go/pkg/server"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"

	_ "net/http/pprof"
)

var Commit string

var options server.Options

// Avoiding replacing the flag parser with something fancier like Viper or urfave/cli
// This hack will work up to a point.
func addEnvVars() {
	if connStr, hasConnstr := os.LookupEnv("MESSAGE_DB_DSN"); hasConnstr {
		options.MessageDBDSN = connStr
	}

	if connStr, hasConnstr := os.LookupEnv("AUTHZ_DB_DSN"); hasConnstr {
		options.Authz.DBDSN = connStr
	}
}

func main() {
	if _, err := flags.Parse(&options); err != nil {
		if err, ok := err.(*flags.Error); !ok || err.Type != flags.ErrHelp {
			fatal("Could not parse options: %s", err)
		}
		return
	}

	log, err := buildLogger(options)
	if err != nil {
		fatal("Could not build logger: %s", err)
	}

	addEnvVars()

	if options.GoProfiling {
		// Spin up a default HTTP server for https://pkg.go.dev/net/http/pprof
		go func() {
			err := http.ListenAndServe("0.0.0.0:6060", nil)
			if err != nil {
				log.Error("serving profiler", zap.Error(err))
			}
		}()
	}

	if options.Version {
		fmt.Printf("Version: %s", Commit)
		return
	}

	if options.GenerateKey {
		privKey, err := crdt.GenerateNodeKey()
		if err != nil {
			log.Fatal("generating node private key", zap.Error(err))
		}
		privKeyHex, err := crdt.HexEncodeNodeKey(privKey)
		if err != nil {
			log.Fatal("encoding node key", zap.Error(err))
		}

		nodeId, err := peer.IDFromPublicKey(privKey.GetPublic())
		if err != nil {
			log.Fatal("creating node id from key", zap.Error(err))
		}

		output := map[string]string{
			"private_node_key": privKeyHex,
			"public_node_id":   nodeId.Pretty(),
		}
		outputJSON, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			log.Fatal("marshaling json", zap.Error(err))
		}
		fmt.Println(string(outputJSON))
		return
	}

	if options.CreateMessageMigration != "" && options.MessageDBDSN != "" {
		if err := server.CreateMessageMigration(options.CreateMessageMigration, options.MessageDBDSN, options.WaitForDB); err != nil {
			log.Fatal("creating message db migration", zap.Error(err))
		}
		return
	}

	if options.CreateAuthzMigration != "" && options.Authz.DBDSN != "" {
		if err := server.CreateAuthzMigration(options.CreateAuthzMigration, options.Authz.DBDSN, options.WaitForDB); err != nil {
			log.Fatal("creating authz db migration", zap.Error(err))
		}
		return
	}

	if options.Tracing.Enable {
		log.Info("starting tracer")
		tracing.Start(Commit, log)
		defer func() {
			log.Info("stopping tracer")
			tracing.Stop()
		}()
	}

	if options.Profiling.Enable {
		env := os.Getenv("ENV")
		if env == "" {
			env = "test"
		}
		ptypes := []profiler.ProfileType{
			profiler.CPUProfile,
			profiler.HeapProfile,
		}
		if options.Profiling.Block {
			ptypes = append(ptypes, profiler.BlockProfile)
		}
		if options.Profiling.Mutex {
			ptypes = append(ptypes, profiler.MutexProfile)
		}
		if options.Profiling.Goroutine {
			ptypes = append(ptypes, profiler.GoroutineProfile)
		}
		if err := profiler.Start(
			profiler.WithService("xmtpd"),
			profiler.WithEnv(env),
			profiler.WithVersion(Commit),
			profiler.WithProfileTypes(ptypes...),
		); err != nil {
			log.Fatal("starting profiler", zap.Error(err))
		}
		defer profiler.Stop()
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	doneC := make(chan bool, 1)
	tracing.GoPanicWrap(ctx, &wg, "main", func(ctx context.Context) {
		s, err := server.New(ctx, log, options)
		if err != nil {
			log.Fatal("initializing server", zap.Error(err))
		}
		s.WaitForShutdown()
		doneC <- true
	})

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	select {
	case sig := <-sigC:
		log.Info("ending on signal", zap.String("signal", sig.String()))
	case <-doneC:
	}

	cancel()
	wg.Wait()
}

func fatal(msg string, args ...any) {
	log.Fatalf(msg, args...)
}

func buildLogger(options server.Options) (*zap.Logger, error) {
	atom := zap.NewAtomicLevel()
	level := zapcore.InfoLevel
	err := level.Set(options.LogLevel)
	if err != nil {
		return nil, err
	}
	atom.SetLevel(level)

	cfg := zap.Config{
		Encoding:         options.LogEncoding,
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
		return nil, err
	}

	log = log.Named("xmtpd")

	return log, nil
}
