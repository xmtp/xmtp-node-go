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
)

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

func wakuConnectWithAddr(ctx context.Context, n *wakunode.WakuNode, addr string) error {
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

func wakuSubscribeTo(ctx context.Context, n *wakunode.WakuNode, contentTopics []string) (chan *wakuprotocol.Envelope, error) {
	_, f, err := n.Filter().Subscribe(ctx, wakufilter.ContentFilter{
		ContentTopics: contentTopics,
	})
	if err != nil {
		return nil, err
	}
	return f.Chan, nil
}

func newWakuMessage(contentTopic string, timestamp int64, content string) *wakupb.WakuMessage {
	return &wakupb.WakuMessage{
		Payload:      []byte(content),
		ContentTopic: contentTopic,
		Timestamp:    timestamp,
	}
}

func wakuPublish(ctx context.Context, n *wakunode.WakuNode, msg *wakupb.WakuMessage) error {
	_, err := n.Relay().Publish(ctx, msg)
	return err
}

func wakuSubscribeExpect(envC chan *wakuprotocol.Envelope, msgs []*wakupb.WakuMessage) error {
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

func wakuExpectQueryMessagesEventually(log *zap.Logger, n *wakunode.WakuNode, peerAddr string, contentTopics []string, expectedMsgs []*wakupb.WakuMessage) error {
	timeout := 3 * time.Second
	delay := 500 * time.Millisecond
	started := time.Now()
	for {
		msgs, err := wakuQueryMessages(log, n, peerAddr, contentTopics)
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

func wakuQueryMessages(log *zap.Logger, c *wakunode.WakuNode, peerAddr string, contentTopics []string) ([]*wakupb.WakuMessage, error) {
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
