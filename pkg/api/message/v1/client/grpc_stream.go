package client

import (
	"context"
	"io"

	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcStream struct {
	cancel context.CancelFunc
	stream messagev1.MessageApi_SubscribeClient
}

func (s *grpcStream) Next() (*messagev1.Envelope, error) {
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
