package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messageV1 "github.com/xmtp/proto/go/message_api/v1"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
)

func Test_HTTPRootPath(t *testing.T) {
	t.Parallel()

	server, cleanup := newTestServer(t)
	defer cleanup()

	// Root path responds with 404.
	var rootRes map[string]interface{}
	resp, err := http.Post(server.httpListenAddr(), "application/json", nil)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &rootRes)
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{
		"code":    float64(5),
		"message": "Not Found",
		"details": []interface{}{},
	}, rootRes)
}

func Test_SubscribePublishQuery(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		// start subscribe stream
		stream, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer stream.Close()
		time.Sleep(50 * time.Millisecond)

		// publish 10 messages
		envs := makeEnvelopes(10)
		publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs})
		require.NoError(t, err)
		require.NotNil(t, publishRes)

		// read subscription
		subscribeExpect(t, stream, envs)

		// query for messages
		requireEventuallyStored(t, ctx, client, envs)
	})
}

func Test_QueryNonExistentTopic(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		queryRes, err := client.Query(ctx, &messageV1.QueryRequest{
			ContentTopics: []string{"does-not-exist"},
		})
		require.NoError(t, err)
		require.NotNil(t, queryRes)
		require.Len(t, queryRes.Envelopes, 0)
	})
}

func Test_SubscribeClientClose(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		// start subscribe stream
		stream, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer stream.Close()
		time.Sleep(50 * time.Millisecond)

		// publish 5 messages
		envs := makeEnvelopes(10)
		publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs[:5]})
		require.NoError(t, err)
		require.NotNil(t, publishRes)

		// receive 5 and close the stream
		subscribeExpect(t, stream, envs[:5])
		err = stream.Close()
		require.NoError(t, err)

		// publish another 5
		publishRes, err = client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs[5:]})
		require.NoError(t, err)
		require.NotNil(t, publishRes)
		time.Sleep(50 * time.Millisecond)

		_, err = stream.Next()
		require.Equal(t, io.EOF, err)
	})
}

func Test_SubscribeServerClose(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		// Subscribe to topics.
		stream, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer stream.Close()
		time.Sleep(50 * time.Millisecond)

		// Publish 5 messages.
		envs := makeEnvelopes(5)
		publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs})
		require.NoError(t, err)
		require.NotNil(t, publishRes)

		// Receive 5
		subscribeExpect(t, stream, envs[:5])

		// stop Server
		server.Close()

		_, err = stream.Next()
		require.Equal(t, io.EOF, err)
	})
}

func Test_MultipleSubscriptions(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		// start 2 streams
		stream1, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer stream1.Close()
		stream2, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer stream2.Close()
		time.Sleep(50 * time.Millisecond)

		// publish 5 envelopes
		envs := makeEnvelopes(10)
		publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs[:5]})
		require.NoError(t, err)
		require.NotNil(t, publishRes)

		// receive 5 envelopes on both streams
		subscribeExpect(t, stream1, envs[:5])
		subscribeExpect(t, stream2, envs[:5])

		// close stream1, start stream3
		err = stream1.Close()
		require.NoError(t, err)
		stream3, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer stream3.Close()
		time.Sleep(50 * time.Millisecond)

		// publish another 5 envelopes
		publishRes, err = client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs[5:]})
		require.NoError(t, err)
		require.NotNil(t, publishRes)

		// receive 5 on stream 2 and 3
		subscribeExpect(t, stream2, envs[5:])
		subscribeExpect(t, stream3, envs[5:])
	})
}

func Test_QueryPaging(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		// Store 10 envelopes with increasing SenderTimestamp
		envs := makeEnvelopes(10)
		publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs})
		require.NoError(t, err)
		require.NotNil(t, publishRes)
		time.Sleep(50 * time.Millisecond)
		requireEventuallyStored(t, ctx, client, envs)

		// We want to page through envs[2]-envs[8] in pages of 3 in reverse order
		result := make([]*messageV1.Envelope, 7)
		for i := 0; i < len(result); i++ {
			result[i] = envs[8-i]
		}

		query := &messageV1.QueryRequest{
			ContentTopics: []string{"topic"},
			StartTimeNs:   envs[2].TimestampNs,
			EndTimeNs:     envs[8].TimestampNs,
			PagingInfo: &messageV1.PagingInfo{
				Limit:     3,
				Direction: messageV1.SortDirection_SORT_DIRECTION_DESCENDING,
			},
		}

		// 1st page
		queryRes, err := client.Query(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, queryRes)
		requireEnvelopesEqual(t, result[0:3], queryRes.Envelopes)

		// 2nd page
		query.PagingInfo.Cursor = queryRes.PagingInfo.Cursor
		queryRes, err = client.Query(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, queryRes)
		requireEnvelopesEqual(t, result[3:6], queryRes.Envelopes)

		// 3rd page (only 1 envelope left)
		query.PagingInfo.Cursor = queryRes.PagingInfo.Cursor
		queryRes, err = client.Query(ctx, query)
		require.NoError(t, err)
		require.NotNil(t, queryRes)
		requireEnvelopesEqual(t, result[6:], queryRes.Envelopes)
	})
}
