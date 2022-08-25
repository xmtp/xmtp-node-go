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
	ctx, err := withAuth(ctx)
	if err != nil {
		return err
	}
	client := messageclient.NewHTTPClient(ctx, s.config.APIURL)

	contentTopic := "test-" + randomStringLower(5)
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

	return nil
}
