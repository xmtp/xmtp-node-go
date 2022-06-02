package authn

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
	"github.com/xmtp/xmtp-node-go/logging"
	pb2 "github.com/xmtp/xmtp-node-go/pb"
	"github.com/xmtp/xmtp-node-go/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"math"
	"net"
	"testing"
	"time"
)

var (
	sampleAuthReq001 = "0a9d020a4c088fd690bd91301a430a41047925bb341e63d10e3d16ccc33c7da6a6115d74e5429e8426761c1332961e48cb1627a62cdcebffe733d23c371a4b78b332d5caed7e3686c0534aa8388688a01212460a440a4076a5aafb9e8bf334629167bdc1ec3d765471f1ef22db661fbb8f24d05a34956b0e7225052ca955f73b97d38b30efb2e60ae0cd17bdda6e39b9b1cb2d0196809e10011a3f0a2a307865386162344535333335323338346234334333454533383942363041343734643137343145423336120a546573745065657249441894d990bd913022440a420a400d8423577fbf7f7ee0a878878a6804489104c19d6a4bdca280457fda7f0dc06408a040b3a79d23d403a610572453810932c5343bf945887f6148e9e9fad2fb29"
	sampleAuthReq002 = "0a9d020a4c08ddcbaebd91301a430a4104411941157aa8041fbf6a1972fe21316f897a3bd0029ce30db43fc76181924a0aa33a261c048016ae63949a9cf6f4a821a395d933a071b429dd321cb06041aaea12440a420a40947579ea9ecc0e5d20c21fff28356606df8ecd5479892828eca3689211a0c02b0b3c148f1012313978e86af29aecb9d6e9ceb3ab108428b7933ab9340b740d241a3f0a2a307836383339663743333537323941383064374136323130373832334337376330653932423930313438120a5465737450656572494418decdaebd913022460a440a4025383de3ff37416900195dd2c7b9a640762c785a22f474565fcb6e531855ca605d387f00ce1c1391fe3bf4bb03e724f244c2832c3303392dda2feefb5557c8f71001"
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

func CreateHost(ctx context.Context, log *zap.Logger) (host.Host, error) {
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

	libP2pHost, err := CreateHost(ctx, log)
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

	return authResponseRPC.GetV1().AuthSuccessful, nil
}

// This integration test checks that data can flow between a mock client and Auth protocol. As the authn request was
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
	client, err := CreateHost(ctx, log)
	if err != nil {
		log.Error("Test client could not be created", zap.Error(err))
		cancel()
		return
	}

	go func() {
		node.Start()
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
	logger, _ := zap.NewDevelopment()
	ctx := logging.With(context.Background(), logger)

	connectingPeerId := types.PeerId("TestPeerID")
	req, err := LoadSerializedAuthReq(sampleAuthReq001)
	require.NoError(t, err)

	authData, err := unpackAuthData(req.AuthDataBytes)
	require.NoError(t, err)

	peerId, walletAddr, err := validateRequest(ctx, req, connectingPeerId)
	require.NoError(t, err)
	require.Equal(t, authData.WalletAddr, string(walletAddr), "address Match")
	require.Equal(t, connectingPeerId, peerId, "address Match")
}

func TestV1_BadAuthSig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := logging.With(context.Background(), logger)

	connectingPeerId := types.PeerId("TestPeerID")
	req, err := LoadSerializedAuthReq(sampleAuthReq001)
	require.NoError(t, err)

	authData, err := unpackAuthData(req.AuthDataBytes)
	require.NoError(t, err)

	authData.WalletAddr = "0000000"
	req.AuthDataBytes, _ = proto.Marshal(authData)

	_, _, err = validateRequest(ctx, req, connectingPeerId)
	require.Error(t, err)
}

func TestV1_PeerIdSpoof(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := logging.With(context.Background(), logger)

	req, err := LoadSerializedAuthReq(sampleAuthReq001)
	require.NoError(t, err)
	connectingPeerId := types.PeerId("InvalidPeerID")

	_, _, err = validateRequest(ctx, req, connectingPeerId)
	require.Error(t, err)
}

func TestV1_SignatureMismatch(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := logging.With(context.Background(), logger)

	req1, err := LoadSerializedAuthReq(sampleAuthReq001)
	require.NoError(t, err)
	req2, err := LoadSerializedAuthReq(sampleAuthReq002)
	require.NoError(t, err)

	authData1, err := unpackAuthData(req1.AuthDataBytes)
	require.NoError(t, err)
	authData2, err := unpackAuthData(req2.AuthDataBytes)
	require.NoError(t, err)

	// Nominal Checks
	_, _, err = validateRequest(ctx, req1, types.PeerId(authData1.PeerId))
	require.NoError(t, err)
	_, _, err = validateRequest(ctx, req2, types.PeerId(authData2.PeerId))
	require.NoError(t, err)

	// Swap Signatures to check for valid but mismatched signatures
	req1.WalletSignature = req2.WalletSignature
	req2.AuthSignature = req1.AuthSignature

	// Expect Errors as the derived walletAddr will not match the one supplied in AuthData
	_, _, err = validateRequest(ctx, req1, types.PeerId(authData1.PeerId))
	require.Error(t, err)
	_, _, err = validateRequest(ctx, req2, types.PeerId(authData2.PeerId))
	require.Error(t, err)
}
