package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messageV1 "github.com/xmtp/proto/v3/go/message_api/v1"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

func Test_MaxMessageSize(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		// start subscribe stream
		stream, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer stream.Close()
		time.Sleep(50 * time.Millisecond)

		// publish valid message
		envs := []*messageV1.Envelope{
			{
				ContentTopic: "topic",
				Message:      make([]byte, testMaxMsgSize-100), // subtract some bytes for the rest of the envelope
				TimestampNs:  1,
			},
		}
		publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs})
		require.NoError(t, err)
		require.NotNil(t, publishRes)
		subscribeExpect(t, stream, envs)
		requireEventuallyStored(t, ctx, client, envs)

		// publish invalid message
		envs = []*messageV1.Envelope{
			{
				ContentTopic: "topic",
				Message:      make([]byte, testMaxMsgSize+100),
				TimestampNs:  1,
			},
		}
		_, err = client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs})
		grpcErr, ok := status.FromError(err)
		if ok {
			require.Equal(t, codes.ResourceExhausted, grpcErr.Code())
			require.Regexp(t, `grpc: received message larger than max \(\d+ vs\. \d+\)`, grpcErr.Message())
		} else {
			require.Regexp(t, `429 Too Many Requests: {"code"\s?:8,\s?"message":\s?"grpc: received message larger than max \(\d+ vs\. \d+\)",\s?"details":\s?\[\]}`, err.Error())
		}
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

		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		_, err = stream.Next(ctx)
		require.Equal(t, io.EOF, err)
	})
}

func Test_SubscribeAllClientClose(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		// start subscribe stream
		stream, err := client.SubscribeAll(ctx)
		require.NoError(t, err)
		defer stream.Close()
		time.Sleep(50 * time.Millisecond)

		// publish 5 messages
		envs := makeEnvelopes(10)
		for i, env := range envs {
			envs[i].ContentTopic = "/xmtp/0/" + env.ContentTopic
		}
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

		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		_, err = stream.Next(ctx)
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

		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		_, err = stream.Next(ctx)
		require.Equal(t, io.EOF, err)
	})
}

func Test_SubscribeAllServerClose(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		// Subscribe to topics.
		stream, err := client.SubscribeAll(ctx)
		require.NoError(t, err)
		defer stream.Close()
		time.Sleep(50 * time.Millisecond)

		// Publish 5 messages.
		envs := makeEnvelopes(5)
		for i, env := range envs {
			envs[i].ContentTopic = "/xmtp/0/" + env.ContentTopic
		}
		publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs})
		require.NoError(t, err)
		require.NotNil(t, publishRes)

		// Receive 5
		subscribeExpect(t, stream, envs[:5])

		// stop Server
		server.Close()

		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		_, err = stream.Next(ctx)
		require.Equal(t, io.EOF, err)
	})
}

func Test_Subscribe_ContextTimeout(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		stream, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer stream.Close()
		time.Sleep(100 * time.Millisecond)

		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		_, err = stream.Next(ctx)
		require.EqualError(t, err, context.DeadlineExceeded.Error())
	})
}

func Test_Subscribe_ContextCancel(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		stream, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer stream.Close()
		time.Sleep(100 * time.Millisecond)

		ctx, cancel := context.WithCancel(ctx)
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()
		_, err = stream.Next(ctx)
		require.EqualError(t, err, context.Canceled.Error())
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

func Test_BatchQuery(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		// Store 10 envelopes with increasing SenderTimestamp
		envs := makeEnvelopes(10)
		publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs})
		require.NoError(t, err)
		require.NotNil(t, publishRes)
		time.Sleep(50 * time.Millisecond)
		requireEventuallyStored(t, ctx, client, envs)

		batchSize := 50
		// Spam a bunch of batch queries with the same topic
		repeatedQueries := make([]*messageV1.QueryRequest, 0)
		for i := 0; i < batchSize; i++ {
			// Alternate sort directions to test that individual paging info is respected
			direction := messageV1.SortDirection_SORT_DIRECTION_ASCENDING
			if i%2 == 1 {
				direction = messageV1.SortDirection_SORT_DIRECTION_DESCENDING
			}
			query := &messageV1.QueryRequest{
				ContentTopics: []string{"topic"},
				PagingInfo: &messageV1.PagingInfo{
					Direction: direction,
				},
			}
			repeatedQueries = append(repeatedQueries, query)
		}
		batchQueryRes, err := client.BatchQuery(ctx, &messageV1.BatchQueryRequest{
			Requests: repeatedQueries,
		})
		require.NoError(t, err)
		require.NotNil(t, batchQueryRes)

		// Descending envs
		descendingEnvs := make([]*messageV1.Envelope, len(envs))
		for i := 0; i < len(envs); i++ {
			descendingEnvs[len(envs)-i-1] = envs[i]
		}

		for i, response := range batchQueryRes.Responses {
			if i%2 == 1 {
				// Reverse the response.Envelopes
				requireEnvelopesEqual(t, descendingEnvs, response.Envelopes)
			} else {
				requireEnvelopesEqual(t, envs, response.Envelopes)
			}
		}
	})
}

func Test_BatchQueryOverLimitError(t *testing.T) {
	ctx := withAuth(t, context.Background())
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		// Store 10 envelopes with increasing SenderTimestamp
		envs := makeEnvelopes(10)
		publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{Envelopes: envs})
		require.NoError(t, err)
		require.NotNil(t, publishRes)
		time.Sleep(50 * time.Millisecond)
		requireEventuallyStored(t, ctx, client, envs)

		// Limit is 50 queries implicitly so 100 should result in an error
		batchSize := 100

		// Spam a bunch of batch queries with the same topic
		repeatedQueries := make([]*messageV1.QueryRequest, 0)
		for i := 0; i < batchSize; i++ {
			query := &messageV1.QueryRequest{
				ContentTopics: []string{"topic"},
				PagingInfo:    &messageV1.PagingInfo{},
			}
			repeatedQueries = append(repeatedQueries, query)
		}
		_, err = client.BatchQuery(ctx, &messageV1.BatchQueryRequest{
			Requests: repeatedQueries,
		})
		grpcErr, ok := status.FromError(err)
		if ok {
			require.Equal(t, codes.InvalidArgument, grpcErr.Code())
			require.Regexp(t, `cannot exceed \d+ requests in single batch`, grpcErr.Message())
		} else {
			require.Regexp(t, `cannot exceed \d+ requests in single batch`, err.Error())
		}
	})
}

func Test_Publish_DenyListed(t *testing.T) {
	token, data, err := GenerateToken(time.Now(), false)
	require.NoError(t, err)
	et, err := EncodeToken(token)
	require.NoError(t, err)
	ctx := metadata.AppendToOutgoingContext(context.Background(), authorizationMetadataKey, "Bearer "+et)

	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, s *Server) {
		err := s.AllowLister.Deny(ctx, data.WalletAddr)
		require.NoError(t, err)

		publishRes, err := client.Publish(ctx, &messageV1.PublishRequest{})
		expectWalletDenied(t, err)
		require.Nil(t, publishRes)
	})
}

func expectWalletDenied(t *testing.T, err error) {
	grpcErr, ok := status.FromError(err)
	if ok {
		require.Equal(t, codes.PermissionDenied, grpcErr.Code())
		require.Equal(t, "wallet is deny listed", grpcErr.Message())
	} else {
		parts := strings.SplitN(err.Error(), ": ", 2)
		reason, msgJSON := parts[0], parts[1]
		require.Equal(t, "403 Forbidden", reason)
		var msg map[string]interface{}
		err := json.Unmarshal([]byte(msgJSON), &msg)
		require.NoError(t, err)
		require.Equal(t, map[string]interface{}{
			"code":    float64(codes.PermissionDenied),
			"message": "wallet is deny listed",
			"details": []interface{}{},
		}, msg)
	}
}
