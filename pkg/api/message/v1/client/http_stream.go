package client

import (
	"bufio"
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

func (s *httpStream) Next() (*messagev1.Envelope, error) {
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
