package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	messageV1 "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_HTTPNotFound(t *testing.T) {
	t.Parallel()

	server, cleanup := newTestServer(t)
	defer cleanup()

	// Root path responds with 404.
	var rootRes map[string]interface{}
	client := http.Client{Timeout: time.Second * 2}
	resp, err := client.Post(server.httpListenAddr()+"/not-found", "application/json", nil)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
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

func Test_HTTPRootPath(t *testing.T) {
	t.Parallel()

	server, cleanup := newTestServer(t)
	defer cleanup()

	// Root path responds with 404.
	client := http.Client{Timeout: time.Second * 2}
	resp, err := client.Post(server.httpListenAddr(), "", nil)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)
}

func Test_QueryNoTopics(t *testing.T) {
	ctx := context.Background()
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		queryRes, err := client.Query(ctx, &messageV1.QueryRequest{
			ContentTopics: []string{},
		})
		grpcErr, ok := status.FromError(err)
		if ok {
			require.Equal(t, codes.InvalidArgument, grpcErr.Code())
			require.EqualError(
				t,
				err,
				`rpc error: code = InvalidArgument desc = content topics required`,
			)
		} else {
			require.Regexp(t, `400 Bad Request: {"code\":3,\s?"message":"content topics required",\s?"details":\[\]}`, err.Error())
		}
		require.Nil(t, queryRes)
	})
}

func Test_QueryTooManyRows(t *testing.T) {
	ctx := context.Background()
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		queryRes, err := client.Query(ctx, &messageV1.QueryRequest{
			ContentTopics: []string{"foo"},
			PagingInfo: &messageV1.PagingInfo{
				Limit: 200,
			},
		})
		grpcErr, ok := status.FromError(err)
		if ok {
			require.Equal(t, codes.InvalidArgument, grpcErr.Code())
			require.EqualError(
				t,
				err,
				`rpc error: code = InvalidArgument desc = cannot exceed 100 rows per query`,
			)
		} else {
			require.Regexp(t, `400 Bad Request: {"code\":3,\s?"message":"cannot exceed 100 rows per query",\s?"details":\[\]}`, err.Error())
		}
		require.Nil(t, queryRes)
	})
}

func Test_QueryNonExistentTopic(t *testing.T) {
	ctx := context.Background()
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, _ *Server) {
		queryRes, err := client.Query(ctx, &messageV1.QueryRequest{
			ContentTopics: []string{"does-not-exist"},
		})
		require.NoError(t, err)
		require.NotNil(t, queryRes)
		require.Len(t, queryRes.Envelopes, 0)
	})
}

func Test_Subscribe_ContextTimeout(t *testing.T) {
	ctx := context.Background()
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		stream, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer func() {
			_ = stream.Close()
		}()
		time.Sleep(100 * time.Millisecond)

		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		_, err = stream.Next(ctx)
		require.EqualError(t, err, context.DeadlineExceeded.Error())
	})
}

func Test_Subscribe_ContextCancel(t *testing.T) {
	ctx := context.Background()
	testGRPCAndHTTP(t, ctx, func(t *testing.T, client messageclient.Client, server *Server) {
		stream, err := client.Subscribe(ctx, &messageV1.SubscribeRequest{
			ContentTopics: []string{"topic"},
		})
		require.NoError(t, err)
		defer func() {
			_ = stream.Close()
		}()
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
