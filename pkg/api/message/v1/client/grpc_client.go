package client

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type grpcClient struct {
	grpc messagev1.MessageApiClient
}

func NewGRPCClient(ctx context.Context, dialFn func(context.Context) (*grpc.ClientConn, error)) (*grpcClient, error) {
	conn, err := dialFn(ctx)
	if err != nil {
		return nil, err
	}
	return &grpcClient{
		grpc: messagev1.NewMessageApiClient(conn),
	}, nil
}

func (c *grpcClient) Close() error {
	return nil
}

func (c *grpcClient) Subscribe(ctx context.Context, r *messagev1.SubscribeRequest) (Stream, error) {
	ctx, cancel := context.WithCancel(ctx)
	sub, err := c.grpc.Subscribe(ctx, r)
	if err != nil {
		cancel()
		return nil, err
	}
	stream := &grpcStream{
		cancel: cancel,
		stream: sub,
	}

	// Wait for confirmation of the subscription.
	ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
	env, err := stream.Next(ctx)
	cancel()
	if err != nil {
		return nil, errors.Wrap(err, "waiting for subscribe confirmation")
	}
	if !proto.Equal(env, &messagev1.Envelope{}) {
		return nil, fmt.Errorf("expected empty env as subscribe confirmation, got: %s", env)
	}

	return stream, nil
}

func (c *grpcClient) Publish(ctx context.Context, r *messagev1.PublishRequest) (*messagev1.PublishResponse, error) {
	return c.grpc.Publish(ctx, r)
}

func (c *grpcClient) Query(ctx context.Context, q *messagev1.QueryRequest) (*messagev1.QueryResponse, error) {
	return c.grpc.Query(ctx, q)
}
