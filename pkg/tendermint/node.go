package tendermint

import (
	"context"
	"fmt"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	tmconfig "github.com/tendermint/tendermint/config"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	tmnode "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	"google.golang.org/protobuf/proto"
)

type Node struct {
	config *Config
	tm     *tmnode.Node
	db     *badger.DB
	http   *tmhttp.HTTP
}

func NewNode(config *Config) (*Node, error) {
	// Build config.
	cfg, err := config.TendermintConfig()
	if err != nil {
		return nil, err
	}

	// create logger
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger, err = tmflags.ParseLogLevel(cfg.LogLevel, logger, tmconfig.DefaultLogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	// read private validator
	pv := privval.LoadFilePV(
		cfg.PrivValidatorKeyFile(),
		cfg.PrivValidatorStateFile(),
	)

	// read node key
	nodeKey, err := p2p.LoadNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("failed to load node's key: %w", err)
	}

	// create app
	db, err := badger.Open(badger.DefaultOptions(config.DBPath))
	if err != nil {
		return nil, errors.Wrap(err, "opening badger db")
	}
	app := newApplication(config.Log, db)

	// create node
	node, err := tmnode.NewNode(
		cfg,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		tmnode.DefaultGenesisDocProviderFunc(cfg),
		tmnode.DefaultDBProvider,
		tmnode.DefaultMetricsProvider(cfg.Instrumentation),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create new Tendermint node: %w", err)
	}

	logger.Info(fmt.Sprintf("p2p listening on %s", node.Config().P2P.ListenAddress))
	logger.Info(fmt.Sprintf("rpc listening on %s", node.Config().RPC.ListenAddress))

	err = node.Start()
	if err != nil {
		db.Close()
		return nil, errors.Wrap(err, "starting tendermint node")
	}

	http, err := tmhttp.New(node.Config().RPC.ListenAddress, "/websocket")
	if err != nil {
		return nil, err
	}
	err = http.Start()
	if err != nil {
		return nil, err
	}

	return &Node{
		config: config,
		tm:     node,
		db:     db,
		http:   http,
	}, nil
}

func (n *Node) Close() {
	if n.http != nil {
		n.http.Stop()
	}
	if n.tm != nil {
		n.tm.Stop()
		n.tm.Wait()
	}
	if n.db != nil {
		n.db.Close()
	}
}

func (n *Node) HTTP() *tmhttp.HTTP {
	return n.http
}

func (n *Node) Query(ctx context.Context, topic string, limit uint32, cursor *messagev1.IndexCursor, reverse bool, startTime, endTime uint64) ([]*messagev1.Envelope, *messagev1.IndexCursor, error) {

	// TODO: validate/sanitize topics
	var envs []*messagev1.Envelope
	// var nextPageEnv *messagev1.Envelope
	err := n.db.View(func(txn *badger.Txn) error {
		prefix := buildKeyPrefix(topic)
		// if cursor != nil {
		// 	prefix += "," + fmt.Sprintf("%d", cursor.SenderTimeNs) + "," + string(cursor.Digest) + "]"
		// }
		it := txn.NewIterator(badger.IteratorOptions{
			PrefetchValues: true,
			PrefetchSize:   100,
			Reverse:        false,
			AllVersions:    false,
			Prefix:         []byte(prefix),
		})
		defer it.Close()
		var count uint32
		for it.Rewind(); it.ValidForPrefix([]byte(prefix)); it.Next() {
			// if limit > 0 && count >= limit+1 {
			// 	break
			// }
			err := it.Item().Value(func(v []byte) error {
				var env messagev1.Envelope
				err := proto.Unmarshal(v, &env)
				if err != nil {
					return err
				}
				if env.TimestampNs < startTime || (endTime > 0 && env.TimestampNs > endTime) {
					return nil
				}
				// if limit > 0 && count == limit {
				// 	count++
				// 	nextPageEnv = &env
				// 	return nil
				// }
				envs = append(envs, &env)
				count++
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	cursor = nil
	// if nextPageEnv != nil {
	// 	var msgCID string
	// 	if len(nextPageEnv.Message) > 0 {
	// 		id, err := BuildCID(nextPageEnv.Message)
	// 		if err != nil {
	// 			return nil, nil, err
	// 		}
	// 		msgCID = id.String()
	// 	}
	// 	cursor = &messagev1.IndexCursor{
	// 		SenderTimeNs: nextPageEnv.TimestampNs,
	// 		Digest:       []byte(msgCID),
	// 	}
	// }

	return envs, cursor, nil
}
