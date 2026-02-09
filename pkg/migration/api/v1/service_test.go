package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func TestFetchD14NCutover(t *testing.T) {
	ctx := context.Background()
	logger := test.NewLog(t)

	// Example: Jan 1, 2024 00:00:00 UTC in nanoseconds
	expectedTimestamp := uint64(1704067200000000000)

	svc := NewService(logger, expectedTimestamp)
	defer svc.Close()

	resp, err := svc.FetchD14NCutover(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, expectedTimestamp, resp.TimestampNs)
}

func TestFetchD14NCutover_ZeroTimestamp(t *testing.T) {
	ctx := context.Background()
	logger := test.NewLog(t)

	svc := NewService(logger, 0)
	defer svc.Close()

	resp, err := svc.FetchD14NCutover(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(0), resp.TimestampNs)
}
