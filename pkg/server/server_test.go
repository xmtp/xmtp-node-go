package server

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/xmtp-node-go/pkg/api"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestServer_NewShutdown(t *testing.T) {
	t.Parallel()

	_, cleanup := newTestServer(t, nil)
	defer cleanup()
}

func TestServer_hexToECDSA(t *testing.T) {
	tcs := []struct {
		name        string
		key         string
		expectedErr string
	}{
		{
			name: "length 64 valid",
			key:  "1df2b07b69f67be885148ef85ced1576048c7c33fb861f7cd8e9b62e13013330",
		},
		{
			name: "length 62 valid",
			key:  "f2b07b69f67be885148ef85ced1576048c7c33fb861f7cd8e9b62e13013330",
		},
		{
			name: "length 60 valid",
			key:  "b07b69f67be885148ef85ced1576048c7c33fb861f7cd8e9b62e13013330",
		},
		{
			name:        "length 58 valid",
			key:         "7b69f67be885148ef85ced1576048c7c33fb861f7cd8e9b62e13013330",
			expectedErr: "invalid length, need 256 bits",
		},
		{
			name:        "length 66 valid",
			key:         "c7c33fb07b69f67be885148ef85ced1576048c7c33fb861f7cd8e9b62e13013330",
			expectedErr: "invalid length, need 256 bits",
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key, err := hexToECDSA(tc.key)
			if tc.expectedErr != "" {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedErr)
			} else {
				keyHex := hex.EncodeToString(key.D.Bytes())
				require.Equal(t, tc.key, keyHex)
			}
		})
	}
}

func newTestServer(t *testing.T, staticNodes []string) (*Server, func()) {
	_, dbDSN, _ := test.NewDB(t)
	ns, err := server.NewServer(&server.Options{
		Port: server.RANDOM_PORT,
	})
	require.NoError(t, err)
	go ns.Start()
	require.True(t, ns.ReadyForConnections(4*time.Second), "nats server not ready")
	s, err := New(context.Background(), test.NewLog(t), Options{
		Address: "localhost",
		Store: StoreOptions{
			Enable:                   true,
			DbConnectionString:       dbDSN,
			DbReaderConnectionString: dbDSN,
		},
		StaticNodes: staticNodes,
		WSAddress:   "0.0.0.0",
		WSPort:      0,
		EnableWS:    true,
		API: api.Options{
			HTTPPort: 0,
			GRPCPort: 0,
		},
		Metrics: MetricsOptions{
			Enable:       true,
			StatusPeriod: 5 * time.Second,
		},
		NATS: NATSOptions{
			URL: ns.ClientURL(),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, s)
	return s, func() {
		s.Shutdown()
		ns.Shutdown()
	}
}
