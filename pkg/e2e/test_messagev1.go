package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (s *Suite) testMessageV1PublishSubscribeQuery(log *zap.Logger) error {
	ctx := context.Background()

	clientCount := 5
	msgsPerClientCount := 3
	clients := make([]messageclient.Client, clientCount)
	for i := 0; i < clientCount; i++ {
		clients[i] = messageclient.NewHTTPClient(ctx, s.config.APIURL)
		defer clients[i].Close()
	}

	contentTopic := "test-" + randomStringLower(7)

	ctx, err := withAuth(ctx)
	if err != nil {
		return err
	}

	// Subscribe across nodes.
	streams := make([]messageclient.Stream, clientCount)
	for i, client := range clients {
		stream, err := client.Subscribe(ctx, &messagev1.SubscribeRequest{
			ContentTopics: []string{
				contentTopic,
			},
		})
		if err != nil {
			return errors.Wrap(err, "subscribing")
		}
		streams[i] = stream
		defer stream.Close()
	}
	time.Sleep(500 * time.Millisecond)

	// Publish messages.
	envs := make([]*messagev1.Envelope, 0, clientCount*1)
	for i, client := range clients {
		clientEnvs := make([]*messagev1.Envelope, msgsPerClientCount)
		for j := 0; j < msgsPerClientCount; j++ {
			clientEnvs[j] = &messagev1.Envelope{
				ContentTopic: contentTopic,
				TimestampNs:  uint64(j + 1),
				Message:      []byte(fmt.Sprintf("msg%d-%d", i+1, j+1)),
			}
		}
		envs = append(envs, clientEnvs...)
		_, err = client.Publish(ctx, &messagev1.PublishRequest{
			Envelopes: clientEnvs,
		})
		if err != nil {
			return errors.Wrap(err, "publishing")
		}
	}

	// Expect them to be relayed to each subscription.
	for i := 0; i < clientCount; i++ {
		stream := streams[i]
		envC := make(chan *messagev1.Envelope, 100)
		go func() {
			for {
				env, err := stream.Next(ctx)
				if err != nil {
					if isErrClosedConnection(err) {
						break
					}
					s.log.Error("getting next", zap.Error(err))
					break
				}
				if env == nil {
					continue
				}
				envC <- env
			}
		}()
		err = subscribeExpect(envC, envs)
		if err != nil {
			return err
		}
	}

	// Expect that they're stored.
	for _, client := range clients {
		err := expectQueryMessagesEventually(ctx, client, []string{contentTopic}, envs)
		if err != nil {
			return err
		}
	}

	return nil
}

func subscribeExpect(envC chan *messagev1.Envelope, envs []*messagev1.Envelope) error {
	receivedEnvs := []*messagev1.Envelope{}
	waitC := time.After(5 * time.Second)
	var done bool
	for !done {
		select {
		case env := <-envC:
			receivedEnvs = append(receivedEnvs, env)
			if len(receivedEnvs) == len(envs) {
				done = true
			}
		case <-waitC:
			done = true
		}
	}
	return envsDiff(envs, receivedEnvs)
}

func isErrClosedConnection(err error) bool {
	return errors.Is(err, io.EOF) || strings.Contains(err.Error(), "use of closed network connection") || strings.Contains(err.Error(), "response body closed")
}

func expectQueryMessagesEventually(ctx context.Context, client messageclient.Client, contentTopics []string, expectedEnvs []*messagev1.Envelope) error {
	timeout := 3 * time.Second
	delay := 500 * time.Millisecond
	started := time.Now()
	for {
		res, err := client.Query(ctx, &messagev1.QueryRequest{
			ContentTopics: contentTopics,
		})
		if err != nil {
			return err
		}
		envs := res.Envelopes
		if len(envs) == len(expectedEnvs) {
			err := envsDiff(envs, expectedEnvs)
			if err != nil {
				return err
			}
			break
		}
		if time.Since(started) > timeout {
			err := envsDiff(envs, expectedEnvs)
			if err != nil {
				return err
			}
			return errors.Wrap(err, "timeout waiting for query expectation")
		}
		time.Sleep(delay)
	}
	return nil
}

func envsDiff(a, b []*messagev1.Envelope) error {
	diff := cmp.Diff(a, b,
		cmpopts.SortSlices(func(a, b *messagev1.Envelope) bool {
			if a.ContentTopic != b.ContentTopic {
				return a.ContentTopic < b.ContentTopic
			}
			if a.TimestampNs != b.TimestampNs {
				return a.TimestampNs < b.TimestampNs
			}
			return bytes.Compare(a.Message, b.Message) < 0
		}),
		cmp.Comparer(proto.Equal),
	)
	if diff != "" {
		return fmt.Errorf("expected equal, diff: %s", diff)
	}
	return nil
}
