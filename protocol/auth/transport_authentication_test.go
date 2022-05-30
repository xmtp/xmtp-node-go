package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/status-im/go-waku/tests"
	"github.com/stretchr/testify/require"
	"github.com/xmtp/go-msgio/protoio"
	pb2 "github.com/xmtp/xmtp-node-go/protocol/pb"
	"github.com/xmtp/xmtp-node-go/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"math"
	"net"
	"testing"
	"time"
)

var (
	sampleAuthReq001 = "0a94020a4c08939c86ee8f301a430a41047ef767b343649e93d313eb3eabc28cac12012d20da79caae534780d96907b2c15f8e67c94550da17840a65c9178f3713fec50a006787857c7c4aaa2f7c1b1daf12440a420a405dda001c54248df91552fe6cd271b1292b0327d5a5774fc2aa2d4456c166f72d5110721377a23af3deb029edfdc5a6dd5abd8370becd09b8ee31586cf88ae9c61a0a54657374506565724944222a3078396130326441374538373933323066623832373846394532363432373938393235323039334137332a460a440a40a725ec9a6fb57e903afc13e3882f0f5d66f29cbe1f42d6d55520f932cd91585e63dad757d3409d168384baa35f2c38e46d907cba06211b782b1adbe253361bd71001"
	sampleAuthReq002 = "0a94020a4c08d6e2d5b091301a430a4104abe63f9f0b3e98977d5a1bd8cdce3f4d570bebb9586b78d0a8409cdd11bb25be30286d9f5f85d227f8380bb9a08b0695c669ae3d8dc6d314575ba87192d0d14b12460a440a40c6c38738043dcaf75bacf8aee3ec13616c8dc9fc30ef582c3e7b77532383f7d37eb9ff677d2d47f86a6057ad7364eead4e4d24d1ff5bab6f06972359813185d210011a0a54657374506565724944222a3078334643303938326642363735383741394166433441363330423735633730353466363861303633322a440a420a40fb4a4efe87a7a391c4817ea65055d2943ac9bec27ebf99d0aee8d378a3488c376faefe936db3eb774effe6767631c1f097968781ea5ce04b369b91b27e81aa66"
)

func LoadSerializedAuthReq(str string) (*pb2.V1ClientAuthRequest, error) {
	serializedAuthReq, _ := hex.DecodeString(str)

	authReq := &pb2.ClientAuthRequest{}
	err := proto.Unmarshal(serializedAuthReq, authReq)
	if err != nil {
		return nil, err
	}

	req := authReq.GetV1()

	return req, nil
}

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

func ClientAuth(ctx context.Context, log *zap.Logger, h host.Host, peerId types.PeerId, dest multiaddr.Multiaddr, protoId protocol.ID, serializedRequest string) (bool, error) {
	h.Peerstore().AddAddr(peerId.Raw(), dest, peerstore.PermanentAddrTTL)

	err := h.Connect(ctx, h.Peerstore().PeerInfo(peerId.Raw()))
	if err != nil {
		log.Error("host could not connect", zap.Error(err))
		return false, err
	}

	stream, err := h.NewStream(ctx, peerId.Raw(), protoId)
	if err != nil {
		log.Info("", zap.Error(err))
		return false, err
	}

	v1, _ := LoadSerializedAuthReq(serializedRequest)
	authReqRPC := &pb2.ClientAuthRequest{
		Version: &pb2.ClientAuthRequest_V1{
			V1: v1,
		},
	}

	log.Info("REQ", zap.Any("pack", authReqRPC))

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

// This integration test checks that data can flow between a mock client and Auth protocol. As the auth request was
// generated from an oracle the peerIDs between the saved request and the connecting stream will not match, resulting in
// a failed authentication.
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

		didSucceed, err := ClientAuth(ctx, log.Named("MockClient"), client, types.PeerId(node.h.ID()), dest, TransportAuthID_v01beta1, sampleAuthReq002)
		require.NoError(t, err)
		require.False(t, didSucceed)

		cancel()
	}()
	<-ctx.Done()

}

func TestV1_Nominal(t *testing.T) {
	log, _ := zap.NewDevelopment()

	connectingPeerId := types.PeerId("TestPeerID")
	req, err := LoadSerializedAuthReq(sampleAuthReq001)
	require.NoError(t, err)

	peerId, walletAddr, err := validateRequest(req, connectingPeerId, log)
	require.NoError(t, err)
	require.Equal(t, req.WalletAddr, string(walletAddr), "address Match")
	require.Equal(t, connectingPeerId, peerId, "address Match")
}

func TestV1_BadAuthSig(t *testing.T) {
	log, _ := zap.NewDevelopment()

	connectingPeerId := types.PeerId("TestPeerID")
	req, err := LoadSerializedAuthReq(sampleAuthReq001)
	require.NoError(t, err)

	req.WalletAddr = "0000000"

	_, _, err = validateRequest(req, connectingPeerId, log)
	require.Error(t, err)
}

func TestV1_PeerIdSpoof(t *testing.T) {
	log, _ := zap.NewDevelopment()

	req, err := LoadSerializedAuthReq(sampleAuthReq001)
	require.NoError(t, err)
	connectingPeerId := types.PeerId("InvalidPeerID")

	_, _, err = validateRequest(req, connectingPeerId, log)
	require.Error(t, err)
}

func TestV1_SignatureMismatch(t *testing.T) {
	log, _ := zap.NewDevelopment()

	req1, err := LoadSerializedAuthReq(sampleAuthReq001)
	require.NoError(t, err)
	req2, err := LoadSerializedAuthReq(sampleAuthReq002)
	require.NoError(t, err)

	_, _, err = validateRequest(req1, types.PeerId(req1.PeerId), log)
	require.NoError(t, err)

	_, _, err = validateRequest(req2, types.PeerId(req2.PeerId), log)
	require.NoError(t, err)

	req1.WalletSig = req2.WalletSig
	req2.AuthSig = req1.AuthSig

	_, _, err = validateRequest(req1, types.PeerId(req1.PeerId), log)
	require.Error(t, err)

	_, _, err = validateRequest(req2, types.PeerId(req2.PeerId), log)
	require.Error(t, err)
}
