package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	messageV1 "github.com/xmtp/proto/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/authn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// GRPCAndHTTPRun runs a test once with a GRPC client and once with an HTTP client.
// The client passed in supports the client interface defined below.
func GRPCAndHTTPRun(t *testing.T, f func(*testing.T, client, *Server)) {
	GRPCAndHTTPRunWithOptions(t, Options{
		GRPCPort: 0,
		HTTPPort: 0,
	}, f)
}

func GRPCAndHTTPRunWithOptions(t *testing.T, options Options, f func(*testing.T, client, *Server)) {
	t.Run("GRPC", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		server, cleanup := newTestServerWithOptions(t, options)
		defer cleanup()
		f(t, newGRPCClient(t, ctx, server), server)
	})
	t.Run("HTTP", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		server, cleanup := newTestServerWithOptions(t, options)
		defer cleanup()
		f(t, newHttpClient(t, ctx, server), server)
	})
}

// stream is an abstraction of the subscribe response stream
type stream interface {
	// Next returns io.EOF when the stream ends or is closed from either side.
	Next() (*messageV1.Envelope, error)
	// Closing the stream terminates the subscription.
	Close() error
}

// client is an abstraction of the API client
type client interface {
	UseToken(*messageV1.Token) error
	Subscribe(*testing.T, *messageV1.SubscribeRequest) stream
	Publish(*testing.T, *messageV1.PublishRequest) *messageV1.PublishResponse
	Query(*testing.T, *messageV1.QueryRequest) *messageV1.QueryResponse
	RawQuery(*messageV1.QueryRequest) (*messageV1.QueryResponse, error)
	RawPublish(*messageV1.PublishRequest) (*messageV1.PublishResponse, error)
}

// GRPC implementation of stream and client

type grpcStream struct {
	cancel context.CancelFunc
	stream messageV1.MessageApi_SubscribeClient
}

func (s *grpcStream) Next() (*messageV1.Envelope, error) {
	env, err := s.stream.Recv()
	if err == nil {
		return env, nil
	}
	if err.Error() == "EOF" ||
		err.Error() == "unexpected EOF" ||
		status.Code(err) == codes.Canceled {
		err = io.EOF
	}
	return env, err
}

func (s *grpcStream) Close() error {
	s.cancel()
	return nil
}

type grpcClient struct {
	ctx    context.Context
	client messageV1.MessageApiClient
}

func newGRPCClient(t *testing.T, ctx context.Context, server *Server) *grpcClient {
	conn, err := server.dialGRPC(ctx)
	require.NoError(t, err)
	client := messageV1.NewMessageApiClient(conn)
	return &grpcClient{ctx: ctx, client: client}
}

func (c *grpcClient) UseToken(token *messageV1.Token) error {
	et, err := authn.EncodeToken(token)
	if err != nil {
		return err
	}
	// NB: This means it's not possible to reset the token once it's set.
	c.ctx = metadata.AppendToOutgoingContext(c.ctx, "authorization", "Bearer "+et)
	return nil
}

func (c *grpcClient) Subscribe(t *testing.T, r *messageV1.SubscribeRequest) stream {
	ctx, cancel := context.WithCancel(c.ctx)
	stream, err := c.client.Subscribe(ctx, r)
	require.NoError(t, err)
	return &grpcStream{cancel: cancel, stream: stream}
}

func (c *grpcClient) Publish(t *testing.T, r *messageV1.PublishRequest) *messageV1.PublishResponse {
	resp, err := c.RawPublish(r)
	require.NoError(t, err)
	return resp
}

func (c *grpcClient) RawPublish(r *messageV1.PublishRequest) (*messageV1.PublishResponse, error) {
	return c.client.Publish(c.ctx, r)
}

func (c *grpcClient) Query(t *testing.T, q *messageV1.QueryRequest) *messageV1.QueryResponse {
	resp, err := c.RawQuery(q)
	require.NoError(t, err)
	return resp
}

func (c *grpcClient) RawQuery(q *messageV1.QueryRequest) (*messageV1.QueryResponse, error) {
	return c.client.Query(c.ctx, q)
}

// HTTP implementation of stream and client
//
// It seems the gateway causes very different behavior wrt to Subscribe.
// The POST has to happen when the Subscribe is called for the subscription to materialize,
// but the GW doesn't send the response headers until the first message is actually pushed through it.
// That's why the silly dance between Subscribe and httpStream, offloading the POST into a goroutine
// and creating the stream with response channels.

type httpStream struct {
	respC      chan *http.Response
	errC       chan error
	bodyReader *bufio.Reader
	body       io.ReadCloser
}

func (s *httpStream) reader() (*bufio.Reader, error) {
	if s.bodyReader != nil {
		return s.bodyReader, nil
	}
	select {
	case err := <-s.errC:
		return nil, err
	case resp := <-s.respC:
		s.body = resp.Body
		s.bodyReader = bufio.NewReader(s.body)
		return s.bodyReader, nil
	}
}

func (s *httpStream) Next() (*messageV1.Envelope, error) {
	reader, err := s.reader()
	if err != nil {
		return nil, err
	}
	if s.body == nil { // stream was closed
		return nil, io.EOF
	}
	line, err := reader.ReadBytes('\n')
	if err != nil {
		if err != io.EOF || len(line) == 0 {
			return nil, err
		}
	}
	var wrapper struct {
		Result interface{}
	}
	err = json.Unmarshal(line, &wrapper)
	if err != nil {
		return nil, err
	}
	envJSON, err := json.Marshal(wrapper.Result)
	if err != nil {
		return nil, err
	}

	var env messageV1.Envelope
	err = protojson.Unmarshal(envJSON, &env)
	return &env, err
}

func (s *httpStream) Close() error {
	if s.body == nil {
		return nil
	}
	err := s.body.Close()
	s.body = nil
	return err
}

func newHttpStream(respC chan *http.Response, errC chan error) *httpStream {
	return &httpStream{
		respC: respC,
		errC:  errC,
	}
}

type httpClient struct {
	ctx   context.Context
	url   string
	http  *http.Client
	token *messageV1.Token
}

func newHttpClient(t *testing.T, ctx context.Context, server *Server) *httpClient {
	transport := &http.Transport{}
	return &httpClient{
		ctx:  ctx,
		http: &http.Client{Transport: transport},
		url:  "http://" + server.httpListener.Addr().String(),
	}
}

func (c *httpClient) UseToken(token *messageV1.Token) error {
	c.token = token
	return nil
}

func (c *httpClient) Post(path string, req interface{}) (*http.Response, error) {
	var reqJSON []byte
	var err error
	switch req := req.(type) {
	case proto.Message:
		reqJSON, err = protojson.Marshal(req)
		if err != nil {
			return nil, err
		}
	default:
		reqJSON, err = json.Marshal(req)
		if err != nil {
			return nil, err
		}
	}

	url := c.url + path

	post, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, err
	}
	post.Header.Set("Content-Type", "application/json")
	if c.token != nil {
		et, err := authn.EncodeToken(c.token)
		if err != nil {
			return nil, err
		}
		post.Header.Set("Authorization", "Bearer "+et)
	}
	resp, err := c.http.Do(post)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (c *httpClient) Publish(t *testing.T, req *messageV1.PublishRequest) *messageV1.PublishResponse {
	res, err := c.RawPublish(req)
	require.NoError(t, err)
	return res
}

func (c *httpClient) RawPublish(req *messageV1.PublishRequest) (*messageV1.PublishResponse, error) {
	var res messageV1.PublishResponse
	resp, err := c.Post("/message/v1/publish", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("%s: %s", resp.Status, string(body))
	}
	err = protojson.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *httpClient) Subscribe(t *testing.T, req *messageV1.SubscribeRequest) stream {
	respC := make(chan *http.Response)
	errC := make(chan error)
	go func() {
		defer close(respC)
		defer close(errC)
		resp, err := c.Post("/message/v1/subscribe", req)
		if err != nil {
			errC <- err
			return
		}
		if resp.StatusCode != http.StatusOK {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				errC <- err
				return
			}
			errC <- errors.Errorf("%s: %s", resp.Status, string(body))
			return
		}
		respC <- resp
	}()
	return newHttpStream(respC, errC)
}

func (c *httpClient) RawQuery(req *messageV1.QueryRequest) (*messageV1.QueryResponse, error) {
	var res messageV1.QueryResponse
	resp, err := c.Post("/message/v1/query", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("%s: %s", resp.Status, string(body))
	}
	err = protojson.Unmarshal(body, &res)
	if err != nil {
		return nil, errors.Errorf("%s: %s", err, string(body))
	}
	return &res, nil
}

func (c *httpClient) Query(t *testing.T, req *messageV1.QueryRequest) *messageV1.QueryResponse {
	res, err := c.RawQuery(req)
	require.NoError(t, err)
	return res
}
