package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

func CreateClient(ctx context.Context, log *zap.SugaredLogger) (host.Host, error) {
	maxAttempts := 5
	hostStr := "localhost"
	port := 0

	for i := 0; i < maxAttempts; i++ {
		addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(hostStr, "0"))
		if err != nil {
			log.Debugf("unable to resolve tcp addr: %v", err)
			continue
		}
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			l.Close()
			log.Debugf("unable to listen on addr %q: %v", addr, err)
			continue
		}

		port = l.Addr().(*net.TCPAddr).Port
		l.Close()

	}

	host, err := tests.MakeHost(ctx, port, rand.Reader)
	if err != nil {
		return nil, err
	}

	return host, nil
}

func CreateNode(ctx context.Context, log *zap.SugaredLogger) (*XmtpAuthentication, error) {

	host, err := CreateClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return NewXmtpAuthentication(ctx, host, log), nil
}

func ClientAuth(ctx context.Context, log *zap.SugaredLogger, h host.Host, peerId peer.ID, dest multiaddr.Multiaddr, protoId protocol.ID) (bool, error) {
	h.Peerstore().AddAddr(peerId, dest, peerstore.PermanentAddrTTL)

	err := h.Connect(ctx, h.Peerstore().PeerInfo(peerId))
	if err != nil {
		log.Info(err)
		return false, err
	}

	stream, err := h.NewStream(ctx, peerId, protoId)
	if err != nil {
		log.Info(err)
		return false, err
	}

	// Generate Wallet Address for testing
	bytes := make([]byte, 40)
	rand.Read(bytes)
	walletAddr := hex.EncodeToString(bytes)

	// Generates a random signature
	signature := pb2.Signature_EcdsaCompact{EcdsaCompact: &pb2.Signature_ECDSACompact{
		Bytes:    bytes,
		Recovery: 0,
	}}

	ucomp := pb2.PublicKey_Secp256K1Uncompresed{Bytes: bytes}
	pk := pb2.PublicKey_Secp256K1Uncompressed{Secp256K1Uncompressed: &ucomp}
	s2 := pb2.Signature{Union: &signature}

	pk2 := pb2.PublicKey{
		Timestamp: 0,
		Signature: &s2,
		Union:     &pk,
	}

	authReqRPC := &pb2.ClientAuthRequest{
		IdentityKey: &pk2,
		AuthDigest:  fmt.Sprintf("%s|%s", h.ID(), walletAddr),
		AuthSig:     &s2,
	}

	writer := protoio.NewDelimitedWriter(stream)
	reader := protoio.NewDelimitedReader(stream, math.MaxInt32)

	err = writer.WriteMsg(authReqRPC)
	if err != nil {
		log.Error("could not write request", err)
		return false, err
	}

	authResponseRPC := &pb2.ClientAuthResponse{}
	err = reader.ReadMsg(authResponseRPC)
	if err != nil {
		log.Error("could not read response", err)
		return false, err
	}

	return authResponseRPC.AuthSuccessful, nil
}

// convert types take an int and return a string value.
type AuthReqGenerator func(int) string

func TestNoop(t *testing.T) {
	require.True(t, true)
}

// This test uses random signatures and will fail in the future when signatures are validated correctly
func TestRoundTrip(t *testing.T) {

	log := tests.Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	node, err := CreateNode(ctx, log)
	if err != nil {
		log.Error(err)
		return
	}
	client, err := CreateClient(ctx, log)
	if err != nil {
		log.Error(err)
		return
	}

	go func() {
		node.Start()

		dest := node.h.Addrs()[0]

		didSucceed, err := ClientAuth(ctx, log.Named("MockClient"), client, node.h.ID(), dest, TransportAuthID_v00beta1)
		require.NoError(t, err)
		require.True(t, didSucceed)
		ClientAuth(ctx, log.Named("MockClient"), client, node.h.ID(), dest, TransportAuthID_v00beta1)
		require.NoError(t, err)
		require.True(t, didSucceed)
		ClientAuth(ctx, log.Named("MockClient"), client, node.h.ID(), dest, TransportAuthID_v00beta1)
		require.NoError(t, err)
		require.True(t, didSucceed)
		cancel()
	}()
	<-ctx.Done()

}
