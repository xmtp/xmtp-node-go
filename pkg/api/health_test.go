package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

func Test_Health(t *testing.T) {
	ctx := context.Background()
	server, cleanup := newTestServer(t)
	conn, err := server.dialGRPC(ctx)
	assert.NoError(t, err)
	healthClient := healthgrpc.NewHealthClient(conn)

	res, err := healthClient.Check(ctx, &healthgrpc.HealthCheckRequest{})
	assert.NoError(t, err)
	assert.Equal(t, res.Status, healthgrpc.HealthCheckResponse_SERVING)
	cleanup()
}
