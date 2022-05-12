package auth

import (
	"context"
	"math"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	libp2pProtocol "github.com/libp2p/go-libp2p-core/protocol"
	"github.com/xmtp/go-msgio/protoio"
	"github.com/xmtp/xmtp-node-go/protocol/pb"
	"go.uber.org/zap"
)

const TransportAuthID_v00beta1 = libp2pProtocol.ID("/xmtplabs/xmtpv1/clientauth/0.1.0-beta1")

type XmtpAuthentication struct {
	h       host.Host
	ctx     context.Context
	log     *zap.SugaredLogger
	started bool
}

func NewXmtpAuthentication(ctx context.Context, h host.Host, log *zap.SugaredLogger) *XmtpAuthentication {
	xmtpAuth := new(XmtpAuthentication)
	xmtpAuth.ctx = ctx
	xmtpAuth.h = h
	xmtpAuth.log = log.Named("client-auth")

	return xmtpAuth
}

func (xmtpAuth *XmtpAuthentication) Start() error {
	xmtpAuth.h.SetStreamHandler(TransportAuthID_v00beta1, xmtpAuth.onRequest)
	xmtpAuth.log.Info("Auth protocol started")
	xmtpAuth.started = true

	return nil
}

func (xmtpAuth *XmtpAuthentication) onRequest(stream network.Stream) {
	defer stream.Close()

	xmtpAuth.log.Debugf("AuthStream established with %s", stream.Conn().RemotePeer())

	authReqRPC := &pb.ClientAuthRequest{}
	err := xmtpAuth.ReadAuthRequest(authReqRPC, stream)
	if err != nil {
		xmtpAuth.log.Error("could not read request", err)
		return
	}

	// TODO: Verify Signature

	// TODO: Save PeerId to walletAddress map

	err = xmtpAuth.WriteAuthResponse(stream, true)
	if err != nil {
		xmtpAuth.log.Error("could not write request", err)
		return
	}
}

func (xmtpAuth *XmtpAuthentication) WriteAuthResponse(stream network.Stream, isAuthenticated bool) error {
	writer := protoio.NewDelimitedWriter(stream)
	authRespRPC := &pb.ClientAuthResponse{AuthSuccessful: isAuthenticated}
	return writer.WriteMsg(authRespRPC)
}

func (xmtpAuth *XmtpAuthentication) ReadAuthRequest(authReqRPC *pb.ClientAuthRequest, stream network.Stream) error {
	reader := protoio.NewDelimitedReader(stream, math.MaxInt32)
	return reader.ReadMsg(authReqRPC)
}
