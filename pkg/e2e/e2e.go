package e2e

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/status-im/go-waku/tests"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	wakuprotocol "github.com/status-im/go-waku/waku/v2/protocol"
	wakufilter "github.com/status-im/go-waku/waku/v2/protocol/filter"
	wakupb "github.com/status-im/go-waku/waku/v2/protocol/pb"
	wakurelay "github.com/status-im/go-waku/waku/v2/protocol/relay"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"go.uber.org/zap"

	_ "net/http/pprof"
)

type E2E struct {
	ctx context.Context
	log *zap.Logger

	config *Config
}

type Config struct {
	Continuous              bool
	NetworkEnv              string
	BootstrapAddrs          []string
	NodesURL                string
	DelayBetweenRunsSeconds int
}

func New(ctx context.Context, log *zap.Logger, config *Config) *E2E {
	e := &E2E{
		ctx:    ctx,
		log:    log,
		config: config,
	}
	return e
}

func (e *E2E) Run() error {
	if e.config.Continuous {
		go func() {
			err := http.ListenAndServe("localhost:6060", nil)
			if err != nil {
				e.log.Error("serving profiler", zap.Error(err))
			}
		}()
	}

	err := e.withMetricsServer(func() {
		for {
			err := e.runTest("publish subscribe query", e.testPublishSubscribeQuery)
			if err != nil {
				e.log.Error("test failed", zap.Error(err))
			} else {
				e.log.Info("test passed")
			}

			if !e.config.Continuous {
				break
			}
			time.Sleep(time.Duration(e.config.DelayBetweenRunsSeconds) * time.Second)
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *E2E) runTest(name string, fn func() error) error {
	nameTag := newTag(testNameTagKey, name)
	started := time.Now().UTC()

	err := fn()
	if err != nil {
		recordErr := recordFailedRun(e.ctx, nameTag)
		if recordErr != nil {
			e.log.Error("recording failed run metric", zap.Error(err))
		}
		return err
	}
	ended := time.Now().UTC()

	err = recordSuccessfulRun(e.ctx, nameTag)
	if err != nil {
		return err
	}

	err = recordRunDuration(e.ctx, ended.Sub(started), nameTag)
	if err != nil {
		return err
	}

	return nil
}

func fetchBootstrapAddrs(nodesURL string, env string) ([]string, error) {
	client := &http.Client{}
	r, err := client.Get(nodesURL)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var manifest map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&manifest)
	if err != nil {
		return nil, err
	}

	envManifest := manifest[env].(map[string]interface{})
	addrs := make([]string, len(envManifest))
	i := 0
	for _, addr := range envManifest {
		addrs[i] = addr.(string)
		i++
	}

	return addrs, nil
}

func newNode(log *zap.Logger, opts ...wakunode.WakuNodeOption) (*wakunode.WakuNode, func(), error) {
	hostAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:0")

	prvKey, err := newPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	opts = append([]wakunode.WakuNodeOption{
		wakunode.WithLogger(log),
		wakunode.WithPrivateKey(prvKey),
		wakunode.WithHostAddress(hostAddr),
		wakunode.WithWakuRelay(),
		wakunode.WithWakuFilter(true),
		wakunode.WithWebsockets("0.0.0.0", 0),
	}, opts...)
	node, err := wakunode.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	err = node.Start()
	if err != nil {
		return nil, nil, err
	}

	return node, func() {
		node.Stop()
	}, nil
}

func newPrivateKey() (*ecdsa.PrivateKey, error) {
	key, err := tests.RandomHex(32)
	if err != nil {
		return nil, err
	}

	prvKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return nil, err
	}

	return prvKey, nil
}

func connectWithAddr(ctx context.Context, n *wakunode.WakuNode, addr string) error {
	err := n.DialPeer(ctx, addr)
	if err != nil {
		return err
	}

	// This delay is necessary, but it's unclear why at this point. We see
	// similar delays throughout the waku codebase as well for this reason.
	time.Sleep(100 * time.Millisecond)

	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func randomStringLower(n int) string {
	return strings.ToLower(randomString(n))
}

func subscribeTo(ctx context.Context, n *wakunode.WakuNode, contentTopics []string) (chan *wakuprotocol.Envelope, error) {
	_, f, err := n.Filter().Subscribe(ctx, wakufilter.ContentFilter{
		ContentTopics: contentTopics,
	})
	if err != nil {
		return nil, err
	}
	return f.Chan, nil
}

func newMessage(contentTopic string, timestamp int64, content string) *wakupb.WakuMessage {
	return &wakupb.WakuMessage{
		Payload:      []byte(content),
		ContentTopic: contentTopic,
		Timestamp:    timestamp,
	}
}

func publish(ctx context.Context, n *wakunode.WakuNode, msg *wakupb.WakuMessage) error {
	_, err := n.Relay().Publish(ctx, msg)
	return err
}

func subscribeExpect(envC chan *wakuprotocol.Envelope, msgs []*wakupb.WakuMessage) error {
	receivedMsgs := []*wakupb.WakuMessage{}
	var done bool
	for !done {
		select {
		case env := <-envC:
			receivedMsgs = append(receivedMsgs, env.Message())
			if len(receivedMsgs) == len(msgs) {
				done = true
			}
		case <-time.After(5 * time.Second):
			done = true
		}
	}
	diff := cmp.Diff(msgs, receivedMsgs, cmpopts.SortSlices(wakuMessagesLessFn))
	if diff != "" {
		fmt.Println(diff)
		return fmt.Errorf("expected equal, diff: %s", diff)
	}
	return nil
}

func expectQueryMessagesEventually(log *zap.Logger, n *wakunode.WakuNode, peerAddr string, contentTopics []string, expectedMsgs []*wakupb.WakuMessage) error {
	timeout := 3 * time.Second
	delay := 500 * time.Millisecond
	started := time.Now()
	for {
		msgs, err := queryMessages(log, n, peerAddr, contentTopics)
		if err != nil {
			return err
		}
		if len(msgs) == len(expectedMsgs) {
			diff := cmp.Diff(msgs, expectedMsgs, cmpopts.SortSlices(wakuMessagesLessFn))
			if diff != "" {
				fmt.Println(diff)
				return fmt.Errorf("expected equal, diff: %s", diff)
			}
			break
		}
		if time.Since(started) > timeout {
			diff := cmp.Diff(msgs, expectedMsgs, cmpopts.SortSlices(wakuMessagesLessFn))
			return fmt.Errorf("timeout waiting for query expectation, diff: %s", diff)
		}
		time.Sleep(delay)
	}
	return nil
}

func wakuMessagesLessFn(a, b *wakupb.WakuMessage) bool {
	if a.ContentTopic != b.ContentTopic {
		return a.ContentTopic < b.ContentTopic
	}
	if a.Timestamp != b.Timestamp {
		return a.Timestamp < b.Timestamp
	}
	return bytes.Compare(a.Payload, b.Payload) < 0
}

func queryMessages(log *zap.Logger, c *wakunode.WakuNode, peerAddr string, contentTopics []string) ([]*wakupb.WakuMessage, error) {
	pi, err := peer.AddrInfoFromString(peerAddr)
	if err != nil {
		return nil, err
	}

	client, err := store.NewClient(
		store.WithClientLog(log),
		store.WithClientHost(c.Host()),
		store.WithClientPeer(pi.ID),
	)
	if err != nil {
		return nil, err
	}

	msgs := []*wakupb.WakuMessage{}
	ctx := context.Background()
	contentFilters := make([]*wakupb.ContentFilter, len(contentTopics))
	for i, contentTopic := range contentTopics {
		contentFilters[i] = &wakupb.ContentFilter{
			ContentTopic: contentTopic,
		}
	}
	_, err = client.Query(ctx, &wakupb.HistoryQuery{
		PubsubTopic:    wakurelay.DefaultWakuTopic,
		ContentFilters: contentFilters,
	}, func(res *wakupb.HistoryResponse) (int, bool) {
		msgs = append(msgs, res.Messages...)
		return len(res.Messages), true
	})
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
