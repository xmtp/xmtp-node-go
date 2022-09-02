package client

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"

	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type httpStream struct {
	ctx        context.Context
	bodyReader *bufio.Reader
	body       io.ReadCloser
}

func newHTTPStream(ctx context.Context, respC chan *http.Response, errC chan error) (*httpStream, error) {
	s := &httpStream{
		ctx: ctx,
	}

	go s.reader(ctx, respC, errC)

	return s, nil
}

func (s *httpStream) reader(ctx context.Context, respC chan *http.Response, errC chan error) error {
	select {
	case err := <-errC:
		return err
	case resp, ok := <-respC:
		if !ok {
			// Exit when channel is closed.
			return nil
		}
		s.body = resp.Body
		s.bodyReader = bufio.NewReader(s.body)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *httpStream) Next(ctx context.Context) (*messagev1.Envelope, error) {
	if s.body == nil || s.bodyReader == nil {
		return nil, io.EOF
	}
	lineC := make(chan []byte)
	errC := make(chan error)
	go func() {
		line, err := s.bodyReader.ReadBytes('\n')
		if ctx.Err() != nil {
			// If the context has already closed, then just return out of this.
			return
		}
		if err != nil {
			errC <- err
			return
		}
		lineC <- line
	}()
	var wrapper struct {
		Result interface{}
	}
	var line []byte
	select {
	case v := <-lineC:
		line = v
	case err := <-errC:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	err := json.Unmarshal(line, &wrapper)
	if err != nil {
		return nil, err
	}
	envJSON, err := json.Marshal(wrapper.Result)
	if err != nil {
		return nil, err
	}

	var env messagev1.Envelope
	err = protojson.Unmarshal(envJSON, &env)
	return &env, err
}

func (s *httpStream) Close() error {
	if s.body == nil {
		return nil
	}
	err := s.body.Close()
	s.body = nil
	s.bodyReader = nil
	return err
}
