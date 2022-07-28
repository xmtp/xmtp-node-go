package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	wakuprotocol "github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/relay"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"go.uber.org/zap"

	_ "net/http/pprof"
)

var (
	envShouldRunE2E                  = envVarBool("E2E")
	envShouldRunE2ETestsContinuously = envVarBool("E2E_CONTINUOUS")
	envNetworkEnv                    = envVar("XMTPD_E2E_ENV", "dev")
	envBootstrapAddrs                = envVarStrings("XMTPD_E2E_BOOTSTRAP_ADDRS")
	envNodesURL                      = envVar("XMTPD_E2E_NODES_URL", "https://nodes.xmtp.com")
	envDelayBetweenRunsSeconds       = envVarInt("XMTPD_E2E_DELAY", 5)
)

func TestE2E(t *testing.T) {
	ctx := context.Background()
	if envShouldRunE2E {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}
	withMetricsServer(t, ctx, func(t *testing.T) {
		for {
			runTest(t, ctx, "publish subscribe query", testPublishSubscribeQuery)

			if !envShouldRunE2ETestsContinuously {
				break
			}
			time.Sleep(time.Duration(envDelayBetweenRunsSeconds) * time.Second)
		}
	})
}

func runTest(t *testing.T, ctx context.Context, name string, fn func(t *testing.T)) {
	t.Run(testName(name), func(t *testing.T) {
		nameTag := newTag(testNameTagKey, name)
		started := time.Now().UTC()
		defer func() {
			ended := time.Now().UTC()
			err := recordRunDuration(ctx, ended.Sub(started), nameTag)
			require.NoError(t, err)
			if t.Failed() {
				err := recordFailedRun(ctx, nameTag)
				require.NoError(t, err)
			} else {
				err := recordSuccessfulRun(ctx, nameTag)
				require.NoError(t, err)
			}
		}()

		fn(t)
	})
}

func testPublishSubscribeQuery(t *testing.T) {
	// Fetch bootstrap node addresses.
	var bootstrapAddrs []string
	if len(envBootstrapAddrs) == 0 {
		var err error
		bootstrapAddrs, err = fetchBootstrapAddrs(envNetworkEnv)
		require.NoError(t, err)
		require.NotEmpty(t, bootstrapAddrs)
		require.Len(t, bootstrapAddrs, 3)
	} else {
		bootstrapAddrs = envBootstrapAddrs
	}

	// Create a client node for each bootstrap node, and connect to it.
	clients := make([]*wakunode.WakuNode, len(bootstrapAddrs))
	for i, addr := range bootstrapAddrs {
		c, cleanup := test.NewNode(t, nil)
		defer cleanup()
		test.ConnectWithAddr(t, c, addr)
		clients[i] = c
	}
	time.Sleep(500 * time.Millisecond)

	// Subscribe to a topic on each client, connected to each node.
	contentTopic := "test-" + test.RandomStringLower(5)
	envCs := make([]chan *wakuprotocol.Envelope, len(clients))
	for i, c := range clients {
		envCs[i] = test.SubscribeTo(t, c, []string{contentTopic})
	}
	time.Sleep(500 * time.Millisecond)

	// Send a message to every node.
	msgs := make([]*pb.WakuMessage, len(clients))
	for i := range clients {
		msgs[i] = test.NewMessage(contentTopic, int64(i+1), fmt.Sprintf("msg%d", i+1))
	}
	for i, sender := range clients {
		test.Publish(t, sender, msgs[i])
	}

	// Expect them to be relayed to all nodes.
	for _, envC := range envCs {
		test.SubscribeExpect(t, envC, msgs)
	}

	// Expect that they've all been stored on each node.
	for i, c := range clients {
		expectQueryMessagesEventually(t, c, bootstrapAddrs[i], []string{contentTopic}, msgs)
	}
}

func testName(s string) string {
	return fmt.Sprintf("%s/%d", s, time.Now().UTC().Unix())
}

func fetchBootstrapAddrs(env string) ([]string, error) {
	client := &http.Client{}
	r, err := client.Get(envNodesURL)
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

func queryMessages(t *testing.T, c *wakunode.WakuNode, peerAddr string, contentTopics []string) []*pb.WakuMessage {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	pi, err := peer.AddrInfoFromString(peerAddr)
	require.NoError(t, err)

	client, err := store.NewClient(
		store.WithClientLog(log),
		store.WithClientHost(c.Host()),
		store.WithClientPeer(pi.ID),
	)
	require.NoError(t, err)

	msgs := []*pb.WakuMessage{}
	ctx := context.Background()
	contentFilters := make([]*pb.ContentFilter, len(contentTopics))
	for i, contentTopic := range contentTopics {
		contentFilters[i] = &pb.ContentFilter{
			ContentTopic: contentTopic,
		}
	}
	msgCount, err := client.Query(ctx, &pb.HistoryQuery{
		PubsubTopic:    relay.DefaultWakuTopic,
		ContentFilters: contentFilters,
	}, func(res *pb.HistoryResponse) (int, bool) {
		msgs = append(msgs, res.Messages...)
		return len(res.Messages), true
	})
	require.NoError(t, err)
	require.Equal(t, msgCount, len(msgs))

	return msgs
}

func expectQueryMessagesEventually(t *testing.T, n *wakunode.WakuNode, peerAddr string, contentTopics []string, expectedMsgs []*pb.WakuMessage) []*pb.WakuMessage {
	var msgs []*pb.WakuMessage
	require.Eventually(t, func() bool {
		msgs = queryMessages(t, n, peerAddr, contentTopics)
		return len(msgs) == len(expectedMsgs)
	}, 3*time.Second, 500*time.Millisecond)
	require.ElementsMatch(t, expectedMsgs, msgs)
	return msgs
}
