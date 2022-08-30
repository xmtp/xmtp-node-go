package api

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messageV1 "github.com/xmtp/proto/go/message_api/v1"
	messageclient "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/client"
)

func makeEnvelopes(count int) (envs []*messageV1.Envelope) {
	for i := 0; i < count; i++ {
		envs = append(envs, &messageV1.Envelope{
			ContentTopic: "topic",
			Message:      []byte(fmt.Sprintf("msg %d", i)),
			TimestampNs:  uint64(i * 1000000000), // i seconds
		})
	}
	return envs
}

func subscribeExpect(t *testing.T, stream messageclient.Stream, expected []*messageV1.Envelope) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	received := []*messageV1.Envelope{}
	for i := 0; i < len(expected); i++ {
		env, err := stream.Next(ctx)
		require.NoError(t, err)
		t.Logf("got %d", i)
		received = append(received, env)
	}
	sortEnvelopes(received)
	requireEnvelopesEqual(t, expected, received)
}

func requireEventuallyStored(t *testing.T, ctx context.Context, client messageclient.Client, expected []*messageV1.Envelope) {
	var queryRes *messageV1.QueryResponse
	require.Eventually(t, func() bool {
		var err error
		queryRes, err = client.Query(ctx, &messageV1.QueryRequest{
			ContentTopics: []string{expected[0].ContentTopic},
			PagingInfo: &messageV1.PagingInfo{
				Direction: messageV1.SortDirection_SORT_DIRECTION_ASCENDING},
		})
		require.NoError(t, err)
		return len(queryRes.Envelopes) == len(expected)
	}, 2*time.Second, 100*time.Millisecond)
	require.NotNil(t, queryRes)
	requireEnvelopesEqual(t, expected, queryRes.Envelopes)
}

func requireEnvelopesEqual(t *testing.T, expected, received []*messageV1.Envelope) {
	require.Equal(t, len(expected), len(received), "length mismatch")
	for i, env := range received {
		requireEnvelopeEqual(t, expected[i], env, "mismatched message[%d]", i)
	}
}

func requireEnvelopeEqual(t *testing.T, expected, actual *messageV1.Envelope, msgAndArgs ...interface{}) {
	require.Equal(t, expected.ContentTopic, actual.ContentTopic, msgAndArgs...)
	require.Equal(t, expected.Message, actual.Message, msgAndArgs...)
	if expected.TimestampNs != 0 {
		require.Equal(t, expected.TimestampNs, actual.TimestampNs, msgAndArgs...)
	}
}

func sortEnvelopes(envelopes []*messageV1.Envelope) {
	sort.SliceStable(envelopes, func(i, j int) bool {
		a, b := envelopes[i], envelopes[j]
		return a.ContentTopic < b.ContentTopic ||
			a.ContentTopic == b.ContentTopic && a.TimestampNs < b.TimestampNs ||
			a.ContentTopic == b.ContentTopic && a.TimestampNs == b.TimestampNs && bytes.Compare(a.Message, b.Message) < 0
	})
}
