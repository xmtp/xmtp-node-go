package e2e

import (
	"bytes"
	"context"
	"fmt"
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

	client := messageclient.NewHTTPClient(ctx, s.config.APIURL)
	contentTopic := "test-" + randomStringLower(5)

	ctx, err := withAuth(ctx)
	if err != nil {
		return err
	}

	// Subscribe across nodes.
	stream, err := client.Subscribe(ctx, &messagev1.SubscribeRequest{
		ContentTopics: []string{
			contentTopic,
		},
	})
	if err != nil {
		return errors.Wrap(err, "subscribing")
	}
	defer stream.Close()

	// Publish messages.
	envs := []*messagev1.Envelope{
		{
			ContentTopic: contentTopic,
			TimestampNs:  1,
			Message:      []byte("msg1"),
		},
	}
	_, err = client.Publish(ctx, &messagev1.PublishRequest{
		Envelopes: envs,
	})
	if err != nil {
		return errors.Wrap(err, "publishing")
	}

	// Expect them to relayed to each subscription.
	envC := make(chan *messagev1.Envelope, 100)
	go func() {
		for {
			env, err := stream.Next()
			if err != nil {
				if isErrUseOfClosedConnection(err) {
					break
				}
				s.log.Error("getting next", zap.Error(err))
				break
			}
			envC <- env
		}
	}()
	err = subscribeExpect(envC, envs)
	if err != nil {
		return err
	}

	// Expect that they're stored.

	return nil
}

func subscribeExpect(envC chan *messagev1.Envelope, envs []*messagev1.Envelope) error {
	receivedEnvs := []*messagev1.Envelope{}
	var done bool
	for !done {
		select {
		case env := <-envC:
			receivedEnvs = append(receivedEnvs, env)
			if len(receivedEnvs) == len(envs) {
				done = true
			}
		case <-time.After(5 * time.Second):
			done = true
		}
	}
	diff := cmp.Diff(
		envs,
		receivedEnvs,
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
		fmt.Println(diff)
		return fmt.Errorf("expected equal, diff: %s", diff)
	}
	return nil
}

func isErrUseOfClosedConnection(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}
