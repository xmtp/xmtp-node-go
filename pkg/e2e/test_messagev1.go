package e2e

import (
	"context"

	"github.com/pkg/errors"
	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
	"go.uber.org/zap"
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
	sub, err := client.Subscribe(ctx, &messagev1.SubscribeRequest{
		ContentTopics: []string{
			contentTopic,
		},
	})
	if err != nil {
		return errors.Wrap(err, "subscribing")
	}
	envC := make(chan *messagev1.Envelope, 100)
	go func() {
		// TODO: teardown on suite close
		for {
			env, err := sub.Next()
			if err != nil {
				s.log.Error("next", zap.Error(err))
				break
			}
			envC <- env
		}
	}()

	// Publish messages.
	_, err = client.Publish(ctx, &messagev1.PublishRequest{
		Envelopes: []*messagev1.Envelope{
			{
				ContentTopic: contentTopic,
				TimestampNs:  1,
				Message:      []byte("msg1"),
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "publishing")
	}

	// Expect them to relayed to each subscription.
	// for _, envC := range envCs {
	// 	err := subscribeExpect(envC, msgs)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// Expect that they're stored.

	return nil
}
