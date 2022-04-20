package filter

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/status-im/go-waku/tests"
	"github.com/stretchr/testify/require"
)

func TestFilterOption(t *testing.T) {
	port, err := tests.FindFreePort(t, "", 5)
	require.NoError(t, err)

	host, err := tests.MakeHost(context.Background(), port, rand.Reader)
	require.NoError(t, err)

	options := []FilterSubscribeOption{
		WithPeer("QmWLxGxG65CZ7vRj5oNXCJvbY9WkF9d9FxuJg8cg8Y7q3"),
		WithAutomaticPeerSelection(),
		WithFastestPeerSelection(context.Background()),
	}

	params := new(FilterSubscribeParameters)
	params.host = host
	params.log = tests.Logger()

	for _, opt := range options {
		opt(params)
	}

	require.Equal(t, host, params.host)
	require.NotNil(t, params.selectedPeer)
}
