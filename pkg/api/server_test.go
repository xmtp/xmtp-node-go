package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	wakustore "github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	messageV1 "github.com/xmtp/proto/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestGRPCServer_HTTP(t *testing.T) {
	t.Parallel()

	server, cleanup := newTestServer(t)
	defer cleanup()

	// Root path responds with 404.
	var rootRes map[string]interface{}
	resp := httpPost(t, server, "", nil)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &rootRes)
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{
		"code":    float64(5),
		"message": "Not Found",
		"details": []interface{}{},
	}, rootRes)

	// Subscribe to topics.
	envC := httpSubscribe(t, server, &messageV1.SubscribeRequest{
		ContentTopics: []string{"test"},
	})
	time.Sleep(50 * time.Millisecond)

	// Publish 10 messages.
	var envs []*messageV1.Envelope
	for i := 0; i < 10; i++ {
		envs = append(envs, &messageV1.Envelope{
			ContentTopic: "test",
			Message:      []byte(fmt.Sprintf("msg %d", i)),
			TimestampNs:  uint64(i * 1000000000), // i seconds
		})
	}
	publishRes := httpPublish(t, server, &messageV1.PublishRequest{Envelopes: envs})
	expectProtoEqual(t, &messageV1.PublishResponse{}, publishRes)

	// Expect messages from subscribed topics.
	subscribeExpect(t, envC, envs)

	// Query for messages.
	var queryRes *messageV1.QueryResponse
	require.Eventually(t, func() bool {
		queryRes = httpQuery(t, server, &messageV1.QueryRequest{
			ContentTopics: []string{"test"},
			PagingInfo: &messageV1.PagingInfo{
				Limit:     uint32(len(envs)),
				Direction: messageV1.SortDirection_SORT_DIRECTION_ASCENDING},
		})
		return len(queryRes.Envelopes) == len(envs)
	}, 2*time.Second, 100*time.Millisecond)
	require.NotNil(t, queryRes)
	require.Len(t, queryRes.Envelopes, len(envs))
	for i, env := range queryRes.Envelopes {
		messageEqual(t, envs[i], env)
	}
}

func TestGRPCServer_GRPC_PublishSubscribeQuery(t *testing.T) {
	t.Parallel()

	server, cleanup := newTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Initialize the client.
	conn, err := server.dialGRPC(ctx)
	require.NoError(t, err)
	client := messageV1.NewMessageApiClient(conn)

	// Subscribe to topics.
	stream, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
		ContentTopics: []string{"topic"},
	})
	require.NoError(t, err)
	envC := make(chan *messageV1.Envelope, 100)
	go func() {
		for {
			env, err := stream.Recv()
			if isEOF(err) {
				return
			}
			require.NoError(t, err)
			envC <- env
		}
	}()
	time.Sleep(50 * time.Millisecond)

	// Publish 10 messages.
	var envs []*messageV1.Envelope
	for i := 0; i < 10; i++ {
		envs = append(envs, &messageV1.Envelope{
			ContentTopic: "topic",
			Message:      []byte(fmt.Sprintf("msg %d", i)),
			TimestampNs:  uint64(i * 1000000000), // i seconds
		})
	}
	publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs})
	require.NoError(t, err)
	require.NotNil(t, publishRes)
	// Expect messages from subscribed topics.
	subscribeExpect(t, envC, envs)

	// Query for messages.
	var queryRes *messageV1.QueryResponse
	require.Eventually(t, func() bool {
		queryRes, err = client.Query(ctx, &messageV1.QueryRequest{
			ContentTopics: []string{"topic"},
			PagingInfo: &messageV1.PagingInfo{
				Limit:     uint32(len(envs)),
				Direction: messageV1.SortDirection_SORT_DIRECTION_ASCENDING},
		})
		if err != nil {
			return false
		}
		return len(queryRes.Envelopes) == len(envs)
	}, 2*time.Second, 100*time.Millisecond)
	require.NoError(t, err)
	require.NotNil(t, queryRes)
	require.Len(t, queryRes.Envelopes, len(envs))
	for i, env := range queryRes.Envelopes {
		messageEqual(t, envs[i], env)
	}

	// Query for messages on a different topic.
	queryRes, err = client.Query(ctx, &messageV1.QueryRequest{
		ContentTopics: []string{"other"},
	})
	require.NoError(t, err)
	require.NotNil(t, queryRes)
	require.Len(t, queryRes.Envelopes, 0)
}

func newTestServer(t *testing.T) (*Server, func()) {
	waku, wakuCleanup := newTestNode(t, nil)
	s, err := New(&Config{
		Options: Options{
			GRPCAddress: "localhost",
			GRPCPort:    0,
			HTTPAddress: "localhost",
			HTTPPort:    0,
		},
		Waku: waku,
		Log:  test.NewLog(t),
	})
	require.NoError(t, err)
	return s, func() {
		s.Close()
		wakuCleanup()
	}
}

func newTestNode(t *testing.T, storeNodes []*wakunode.WakuNode, opts ...wakunode.WakuNodeOption) (*wakunode.WakuNode, func()) {
	var dbCleanup func()
	n, nodeCleanup := test.NewNode(t, storeNodes,
		append(
			opts,
			wakunode.WithWakuStore(true, false),
			wakunode.WithWakuStoreFactory(func(w *wakunode.WakuNode) wakustore.Store {
				// Note that the node calls store.Stop() during it's cleanup,
				// but it that doesn't clean up the given DB, so we make sure
				// to return that in the node cleanup returned here.
				// Note that the same host needs to be used here.
				var store *store.XmtpStore
				store, _, _, dbCleanup = newTestStore(t, w.Host())
				return store
			}),
		)...,
	)
	return n, func() {
		nodeCleanup()
		dbCleanup()
	}
}

func newTestStore(t *testing.T, host host.Host) (*store.XmtpStore, *store.DBStore, func(), func()) {
	db, _, dbCleanup := test.NewDB(t)
	dbStore, err := store.NewDBStore(utils.Logger(), store.WithDBStoreDB(db))
	require.NoError(t, err)

	if host == nil {
		host = test.NewPeer(t)
	}
	store, err := store.NewXmtpStore(
		store.WithLog(utils.Logger()),
		store.WithHost(host),
		store.WithDB(db),
		store.WithMessageProvider(dbStore))
	require.NoError(t, err)

	store.Start(context.Background())

	return store, dbStore, store.Stop, dbCleanup
}

func httpClient(t *testing.T) *http.Client {
	transport := &http.Transport{}
	return &http.Client{Transport: transport}
}

func httpURL(t *testing.T, server *Server, path string) string {
	prefix := "http://" + server.httpListener.Addr().String()
	return prefix + path
}

func httpPost(t *testing.T, server *Server, path string, req interface{}) *http.Response {
	client := httpClient(t)

	var reqJSON []byte
	var err error
	switch req := req.(type) {
	case proto.Message:
		reqJSON, err = protojson.Marshal(req)
		require.NoError(t, err)
	default:
		reqJSON, err = json.Marshal(req)
		require.NoError(t, err)
	}

	url := httpURL(t, server, path)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqJSON))
	require.NoError(t, err)

	return resp
}

func httpPublish(t *testing.T, server *Server, req *messageV1.PublishRequest) *messageV1.PublishResponse {
	var res messageV1.PublishResponse
	resp := httpPost(t, server, "/message/v1/publish", req)
	expectStatusOK(t, resp)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = protojson.Unmarshal(body, &res)
	require.NoError(t, err)
	return &res
}

func httpSubscribe(t *testing.T, server *Server, req *messageV1.SubscribeRequest) chan *messageV1.Envelope {
	envC := make(chan *messageV1.Envelope, 100)
	go func() {
		resp := httpPost(t, server, "/message/v1/subscribe", req)
		expectStatusOK(t, resp)
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			require.NoError(t, err)

			var wrapper struct {
				Result interface{}
			}
			err = json.Unmarshal(line, &wrapper)
			require.NoError(t, err)

			envJSON, err := json.Marshal(wrapper.Result)
			require.NoError(t, err)

			var env messageV1.Envelope
			err = protojson.Unmarshal(envJSON, &env)
			require.NoError(t, err)
			envC <- &env
		}
	}()
	return envC
}

func httpQuery(t *testing.T, server *Server, req *messageV1.QueryRequest) *messageV1.QueryResponse {
	var res messageV1.QueryResponse
	resp := httpPost(t, server, "/message/v1/query", req)
	expectStatusOK(t, resp)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = protojson.Unmarshal(body, &res)
	require.NoError(t, err)
	return &res
}

func expectStatusOK(t *testing.T, resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	}
}

func expectProtoEqual(t *testing.T, a, b proto.Message) {
	require.True(t, proto.Equal(a, b))
}

func subscribeExpect(t *testing.T, envC chan *messageV1.Envelope, expected []*messageV1.Envelope) {
	received := []*messageV1.Envelope{}
	var done bool
	timer := time.After(3 * time.Second)
	for !done {
		select {
		case env := <-envC:
			received = append(received, env)
			if len(received) == len(expected) {
				done = true
			}
		case <-timer:
			done = true
		}
	}
	require.Equal(t, len(expected), len(received))
	sortEnvelopes(received)
	for i, env := range received {
		messageEqual(t, expected[i], env, "mismatched message[%d]", i)
	}
}

func messageEqual(t *testing.T, expected, actual *messageV1.Envelope, msgAndArgs ...interface{}) {
	if expected.TimestampNs != 0 {
		require.Equal(t, expected.TimestampNs, actual.TimestampNs, msgAndArgs...)
	}
	require.Equal(t, expected.ContentTopic, actual.ContentTopic, msgAndArgs...)
	require.Equal(t, expected.Message, actual.Message, msgAndArgs...)
}

func sortEnvelopes(envelopes []*messageV1.Envelope) {
	sort.SliceStable(envelopes, func(i, j int) bool {
		a, b := envelopes[i], envelopes[j]
		return a.ContentTopic < b.ContentTopic ||
			a.ContentTopic == b.ContentTopic && a.TimestampNs < b.TimestampNs ||
			a.ContentTopic == b.ContentTopic && a.TimestampNs == b.TimestampNs && bytes.Compare(a.Message, b.Message) < 0
	})
}

func isEOF(err error) bool {
	return err != nil && (err.Error() == "EOF" || err.Error() == "unexpected EOF")
}
