package auth

import (
	"context"
	"crypto/rand"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/status-im/go-waku/tests"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/go-msgio/protoio"
	pb2 "github.com/xmtp/xmtp-node-go/protocol/pb"
	"go.uber.org/zap"
	"math"
	"net"
	"testing"
	"time"
)

func CreateClient(ctx context.Context, log *zap.Logger) (host.Host, error) {
	maxAttempts := 5
	hostStr := "localhost"
	port := 0

	for i := 0; i < maxAttempts; i++ {
		addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(hostStr, "0"))
		if err != nil {
			log.Error("unable to resolve tcp addr: %v", zap.Error(err))
			continue
		}
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			log.Error("unable to listen on addr %q: %v", zap.String("ipaddr", addr.String()), zap.Error(err))
			err := l.Close()
			if err != nil {
				return nil, err
			}
			continue
		}

		port = l.Addr().(*net.TCPAddr).Port
		err = l.Close()
		if err != nil {
			return nil, err
		}

	}

	libP2pHost, err := tests.MakeHost(ctx, port, rand.Reader)
	if err != nil {
		return nil, err
	}

	return libP2pHost, nil
}

func CreateNode(ctx context.Context, log *zap.Logger) (*XmtpAuthentication, error) {

	libP2pHost, err := CreateClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return NewXmtpAuthentication(ctx, libP2pHost, log), nil
}

func ClientAuth(ctx context.Context, log *zap.Logger, h host.Host, peerId peer.ID, dest multiaddr.Multiaddr, protoId protocol.ID) (bool, error) {
	h.Peerstore().AddAddr(peerId, dest, peerstore.PermanentAddrTTL)

	err := h.Connect(ctx, h.Peerstore().PeerInfo(peerId))
	if err != nil {
		log.Error("host could not connect", zap.Error(err))
		return false, err
	}

	stream, err := h.NewStream(ctx, peerId, protoId)
	if err != nil {
		log.Info("", zap.Error(err))
		return false, err
	}

	// Generates a random signature
	signature := pb2.Signature_EcdsaCompact{EcdsaCompact: &pb2.Signature_ECDSACompact{
		Bytes:    []byte("This is an invalid signature"),
		Recovery: 0,
	}}

	s2 := pb2.Signature{Union: &signature}

	pk2 := pb2.PublicKey{
		Timestamp: 0,
		Signature: &s2,
		Union:     &pb2.PublicKey_Secp256K1Uncompressed{Secp256K1Uncompressed: &pb2.PublicKey_Secp256K1Uncompresed{Bytes: []byte("This is an invalid publicKey")}},
	}

	authReqRPC := &pb2.ClientAuthRequest{
		Version: &pb2.ClientAuthRequest_V1{
			V1: &pb2.V1ClientAuthRequest{
				IdentityKey: &pk2,
				PeerId:      h.ID().String(),
				WalletAddr:  "0x1234567890",
				AuthSig:     &s2,
			},
		},
	}

	writer := protoio.NewDelimitedWriter(stream)
	reader := protoio.NewDelimitedReader(stream, math.MaxInt32)

	err = writer.WriteMsg(authReqRPC)
	if err != nil {
		log.Error("could not write request", zap.Error(err))
		return false, err
	}

	authResponseRPC := &pb2.ClientAuthResponse{}
	err = reader.ReadMsg(authResponseRPC)
	if err != nil {
		log.Error("could not read response", zap.Error(err))
		return false, err
	}

	return authResponseRPC.AuthSuccessful, nil
}

// This test uses random signatures and will fail in the future when signatures are validated correctly
func TestRoundTrip(t *testing.T) {

	log, _ := zap.NewDevelopment()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	node, err := CreateNode(ctx, log)
	if err != nil {
		log.Error("Test node could not be created", zap.Error(err))
		cancel()
		return
	}
	client, err := CreateClient(ctx, log)
	if err != nil {
		log.Error("Test client could not be created", zap.Error(err))
		cancel()
		return
	}

	go func() {
		err := node.Start()
		require.NoError(t, err)
		dest := node.h.Addrs()[0]

		didSucceed, err := ClientAuth(ctx, log.Named("MockClient"), client, node.h.ID(), dest, TransportAuthID_v01beta1)
		require.NoError(t, err)
		require.True(t, didSucceed)
		cancel()
	}()
	<-ctx.Done()

}
