package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	golog "github.com/ipfs/go-log"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/xmtp/xmtp-node-go/pkg/server"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

var Commit string

var options server.Options

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

	cleanup, err := initWakuLogging(options)
	if err != nil {
		log.Fatal("initializing waku logger", zap.Error(err))
	}
	defer cleanup()

	if options.Version {
		fmt.Printf("Version: %s", Commit)
		return
	}

	if options.GenerateKey {
		if err := server.WritePrivateKeyToFile(options.KeyFile, options.Overwrite); err != nil {
			log.Fatal("writing private key file", zap.Error(err))
		}
		return
	}

	if options.CreateMessageMigration != "" && options.Store.DbConnectionString != "" {
		if err := server.CreateMessageMigration(options.CreateMessageMigration, options.Store.DbConnectionString, options.WaitForDB); err != nil {
			log.Fatal("creating message db migration", zap.Error(err))
		}
		return
	}

	if options.CreateAuthzMigration != "" && options.Authz.DbConnectionString != "" {
		if err := server.CreateAuthzMigration(options.CreateAuthzMigration, options.Authz.DbConnectionString, options.WaitForDB); err != nil {
			log.Fatal("creating authz db migration", zap.Error(err))
		}
		return
	}

	if options.Tracing.Enable {
		log.Info("starting tracer")
		tracing.Start(Commit, utils.Logger())
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
		utils.Logger().Info("ending on signal", zap.String("signal", sig.String()))
	case <-doneC:
	}

	cancel()
	wg.Wait()
}

func fatal(msg string, args ...any) {
	log.Fatalf(msg, args...)
}

func initWakuLogging(options server.Options) (func(), error) {
	err := utils.SetLogLevel(options.LogLevel)
	if err != nil {
		return nil, errors.Wrap(err, "parsing log level")
	}
	utils.InitLogger(options.LogEncoding)

	lvl, err := golog.LevelFromString(options.LogLevel)
	if err != nil {
		return nil, errors.Wrap(err, "parsing log level")
	}
	golog.SetAllLoggers(lvl)

	// Note that libp2p reads the encoding from GOLOG_LOG_FMT env var.
	if options.LogEncoding == "json" && os.Getenv("GOLOG_LOG_FMT") == "" {
		utils.Logger().Warn("Set GOLOG_LOG_FMT=json to use json for libp2p logs")
	}

	cleanup := func() { _ = utils.Logger().Sync() }
	return cleanup, nil
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
