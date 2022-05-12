package auth

import (
	"context"
	"github.com/xmtp/xmtp-node-go/logging"
	"math"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	libp2pProtocol "github.com/libp2p/go-libp2p-core/protocol"
	"github.com/xmtp/go-msgio/protoio"
	"github.com/xmtp/xmtp-node-go/protocol/pb"
	"go.uber.org/zap"
)

const TransportAuthID_v01beta1 = libp2pProtocol.ID("/xmtplabs/xmtpv1/clientauth/0.1.0-beta1")

type XmtpAuthentication struct {
	h   host.Host
	ctx context.Context
	log *zap.Logger
}

func NewXmtpAuthentication(ctx context.Context, h host.Host, log *zap.Logger) *XmtpAuthentication {
	xmtpAuth := new(XmtpAuthentication)
	xmtpAuth.ctx = ctx
	xmtpAuth.h = h
	xmtpAuth.log = log.Named("client-auth")

	return xmtpAuth
}

func (xmtpAuth *XmtpAuthentication) Start() error {
	xmtpAuth.h.SetStreamHandler(TransportAuthID_v01beta1, xmtpAuth.onRequest)
	xmtpAuth.log.Info("Auth protocol started")

	return nil
}

func (xmtpAuth *XmtpAuthentication) onRequest(stream network.Stream) {
	defer stream.Close()

	xmtpAuth.log.Debug("AuthStream established with", logging.HostID("peer", stream.Conn().RemotePeer()))

	authReqRPC := &pb.ClientAuthRequest{}
	err := xmtpAuth.ReadAuthRequest(authReqRPC, stream)
	if err != nil {
		xmtpAuth.log.Error("could not read request", zap.Error(err))
		return
	}

	// TODO: Verify Signature

	// TODO: Save PeerId to walletAddress map

	err = xmtpAuth.WriteAuthResponse(stream, true)
	if err != nil {
		xmtpAuth.log.Error("could not write request", zap.Error(err))
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
