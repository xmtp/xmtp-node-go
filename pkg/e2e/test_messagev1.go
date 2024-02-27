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
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	messagev1 "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (s *Suite) testMessageV1PublishSubscribeQuery(log *zap.Logger) error {
	clientCount := 5
	msgsPerClientCount := 3
	clients := make([]messageclient.Client, clientCount)
	for i := 0; i < clientCount; i++ {
		appVersion := "xmtp-e2e/"
		if len(s.config.GitCommit) > 0 {
			appVersion += s.config.GitCommit[:7]
		}
		clients[i] = messageclient.NewHTTPClient(s.log, s.config.APIURL, s.config.GitCommit, appVersion)
		defer clients[i].Close()
	}

	contentTopic := "test-" + s.randomStringLower(12)

	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()
	ctx, err := s.withAuth(ctx)
	if err != nil {
		return errors.Wrap(err, "adding auth token")
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

	// Wait for subscriptions to be set up, for a sync env to be received
	// by all clients.
	syncEnvs := []*messagev1.Envelope{}
	prevSyncEnvs := map[string]bool{}
syncLoop:
	for {
		syncEnv := &messagev1.Envelope{
			ContentTopic: contentTopic,
			Message:      []byte("sync-" + s.randomStringLower(12)),
		}
		_, err = clients[0].Publish(ctx, &messagev1.PublishRequest{
			Envelopes: []*messagev1.Envelope{
				syncEnv,
			},
		})
		if err != nil {
			return errors.Wrap(err, "publishing sync envelope")
		}
		syncEnvs = append(syncEnvs, syncEnv)

		clientReady := make([]bool, len(clients))
		for i := range clients {
			var done bool
			for !done {
				ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				env, err := streams[i].Next(ctx)
				cancel()
				if err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						s.log.Info("waiting for subscription sync", zap.Int("client", i))
						prevSyncEnvs[string(syncEnv.Message)] = true
						continue syncLoop
					}
					if strings.Contains(err.Error(), "429 Too Many Requests") {
						s.log.Info("waiting for subscription sync", zap.Int("client", i), zap.Error(err))
						time.Sleep(1 * time.Minute)
					}
					return errors.Wrap(err, "reading sync envelope")
				}
				if prevSyncEnvs[string(env.Message)] {
					s.log.Info("skipping previous sync envelope", zap.String("value", string(env.Message)))
					continue
				}
				if !proto.Equal(env, syncEnv) {
					return fmt.Errorf("expected sync envelope, got: %s", env)
				}
				clientReady[i] = true
				done = true
			}
		}
		allReady := true
		for _, ready := range clientReady {
			allReady = allReady && ready
		}
		if allReady {
			prevSyncEnvs[string(syncEnv.Message)] = true
			break
		}
	}

	// start listeners
	streamsCh := make([]chan *messagev1.Envelope, len(clients))
	for i := 0; i < clientCount; i++ {
		envC := make(chan *messagev1.Envelope, 100)
		go func(stream messageclient.Stream, envC chan *messagev1.Envelope) {
			for {
				env, err := stream.Next(ctx)
				if err != nil {
					if isErrClosedConnection(err) || err.Error() == "context canceled" {
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
		}(streams[i], envC)
		streamsCh[i] = envC
	}

	// wait until all the listeners are up and ready.
	time.Sleep(100 * time.Millisecond)

	// Publish messages.
	envs := []*messagev1.Envelope{}
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
		err = subscribeExpect(streamsCh[i], envs)
		if err != nil {
			return err
		}
	}

	envs = append(syncEnvs, envs...)

	// Expect that they're stored.
	for _, client := range clients {
		err := expectQueryMessagesEventually(ctx, client, []string{contentTopic}, envs)
		if err != nil {
			return err
		}
	}

	// Expect that they're also available via batch query.
	for _, client := range clients {
		err := expectBatchQueryMessagesEventually(ctx, client, []string{contentTopic}, envs)
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
	err := envsDiff(envs, receivedEnvs)
	if err != nil {
		return errors.Wrap(err, "expected subscribe envelopes")
	}
	return nil
}

func isErrClosedConnection(err error) bool {
	return errors.Is(err, io.EOF) || strings.Contains(err.Error(), "use of closed network connection") || strings.Contains(err.Error(), "response body closed")
}

func expectQueryMessagesEventually(ctx context.Context, client messageclient.Client, contentTopics []string, expectedEnvs []*messagev1.Envelope) error {
	timeout := 10 * time.Second
	delay := 500 * time.Millisecond
	started := time.Now()
	for {
		envs, err := query(ctx, client, contentTopics)
		if err != nil {
			return errors.Wrap(err, "querying")
		}
		if len(envs) == len(expectedEnvs) {
			err := envsDiff(envs, expectedEnvs)
			if err != nil {
				return errors.Wrap(err, "expected query envelopes")
			}
			break
		}
		if time.Since(started) > timeout {
			err := envsDiff(envs, expectedEnvs)
			if err != nil {
				return errors.Wrap(err, "expected query envelopes")
			}
			return fmt.Errorf("timeout waiting for query expectation with no diff")
		}
		time.Sleep(delay)
	}
	return nil
}

func expectBatchQueryMessagesEventually(ctx context.Context, client messageclient.Client, contentTopics []string, expectedEnvs []*messagev1.Envelope) error {
	timeout := 10 * time.Second
	delay := 500 * time.Millisecond
	started := time.Now()
	for {
		batchEnvs, err := batchQuery(ctx, client, contentTopics)
		if err != nil {
			return errors.Wrap(err, "batch querying")
		}
		if len(batchEnvs) == len(expectedEnvs) {
			err := envsDiff(batchEnvs, expectedEnvs)
			if err != nil {
				return errors.Wrap(err, "expected query envelopes from batchquery")
			}
			break
		}
		if time.Since(started) > timeout {
			err := envsDiff(batchEnvs, expectedEnvs)
			if err != nil {
				return errors.Wrap(err, "expected query envelopes from batchquery")
			}
			return fmt.Errorf("timeout waiting for batchquery expectation with no diff")
		}
		time.Sleep(delay)
	}
	return nil
}

func query(ctx context.Context, client messageclient.Client, contentTopics []string) ([]*messagev1.Envelope, error) {
	var envs []*messagev1.Envelope
	var pagingInfo *messagev1.PagingInfo
	for {
		res, err := client.Query(ctx, &messagev1.QueryRequest{
			ContentTopics: contentTopics,
			PagingInfo:    pagingInfo,
		})
		if err != nil {
			return nil, err
		}
		envs = append(envs, res.Envelopes...)
		if len(res.Envelopes) == 0 || res.PagingInfo.Cursor == nil {
			break
		}
		pagingInfo = res.PagingInfo
	}
	return envs, nil
}

func batchQuery(ctx context.Context, client messageclient.Client, contentTopics []string) ([]*messagev1.Envelope, error) {
	var envs []*messagev1.Envelope
	var pagingInfo *messagev1.PagingInfo
	for {
		var queries = make([]*messagev1.QueryRequest, 0)
		for _, contentTopic := range contentTopics {
			queries = append(queries, &messagev1.QueryRequest{
				ContentTopics: []string{contentTopic},
				PagingInfo:    pagingInfo,
			})
		}
		res, err := client.BatchQuery(ctx, &messagev1.BatchQueryRequest{
			Requests: queries,
		})
		if err != nil {
			return nil, err
		}
		var resp *messagev1.QueryResponse
		for _, resp = range res.Responses {
			envs = append(envs, resp.Envelopes...)
			pagingInfo = resp.PagingInfo
		}
		if len(resp.Envelopes) == 0 || resp.PagingInfo.Cursor == nil {
			break
		}
	}
	return envs, nil
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
