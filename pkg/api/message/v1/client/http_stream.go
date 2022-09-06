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
	respC      chan *http.Response
	errC       chan error
	bodyReader *bufio.Reader
	body       io.ReadCloser
}

func newHTTPStream(respC chan *http.Response, errC chan error) (*httpStream, error) {
	return &httpStream{
		respC: respC,
		errC:  errC,
	}, nil
}

func (s *httpStream) reader(ctx context.Context) (*bufio.Reader, error) {
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
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *httpStream) Next(ctx context.Context) (*messagev1.Envelope, error) {
	reader, err := s.reader(ctx)
	if err != nil {
		return nil, err
	}
	if s.body == nil { // stream was closed
		return nil, io.EOF
	}
	lineC := make(chan []byte)
	errC := make(chan error)
	go func() {
		line, err := reader.ReadBytes('\n')
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
	err = json.Unmarshal(line, &wrapper)
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
	return err
}
