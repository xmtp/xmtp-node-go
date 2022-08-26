package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type httpClient struct {
	ctx     context.Context
	url     string
	http    *http.Client
	version string
}

const (
	clientVersionHeaderKey = "x-client-version"
)

func NewHTTPClient(ctx context.Context, serverAddr string, gitCommit string) *httpClient {
	transport := &http.Transport{}
	version := "xmtp-go/"
	if len(gitCommit) > 0 {
		version += gitCommit[:7]
	}
	return &httpClient{
		ctx:     ctx,
		http:    &http.Client{Transport: transport},
		url:     serverAddr,
		version: version,
	}
}

func (c *httpClient) Close() error {
	c.http.CloseIdleConnections()
	return nil
}

func (c *httpClient) Publish(ctx context.Context, req *messagev1.PublishRequest) (*messagev1.PublishResponse, error) {
	res, err := c.rawPublish(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *httpClient) Subscribe(ctx context.Context, req *messagev1.SubscribeRequest) (Stream, error) {
	stream, err := newHTTPStream(func() (*http.Response, error) {
		return c.post(ctx, "/message/v1/subscribe", req)
	})
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (c *httpClient) Query(ctx context.Context, req *messagev1.QueryRequest) (*messagev1.QueryResponse, error) {
	res, err := c.rawQuery(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *httpClient) rawPublish(ctx context.Context, req *messagev1.PublishRequest) (*messagev1.PublishResponse, error) {
	var res messagev1.PublishResponse
	resp, err := c.post(ctx, "/message/v1/publish", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status, string(body))
	}
	err = protojson.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *httpClient) rawQuery(ctx context.Context, req *messagev1.QueryRequest) (*messagev1.QueryResponse, error) {
	var res messagev1.QueryResponse
	resp, err := c.post(ctx, "/message/v1/query", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status, string(body))
	}
	err = protojson.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, string(body))
	}
	return &res, nil
}

func (c *httpClient) post(ctx context.Context, path string, req interface{}) (*http.Response, error) {
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
	md, _ := metadata.FromOutgoingContext(ctx)
	for key, vals := range md {
		if len(vals) > 0 {
			post.Header.Set(key, vals[0])
		}
	}
	clientVersion := post.Header.Get(clientVersionHeaderKey)
	if clientVersion == "" {
		post.Header.Set(clientVersionHeaderKey, c.version)
	}
	resp, err := c.http.Do(post)
	if err != nil {
		return nil, err
	}

	return resp, err
}
