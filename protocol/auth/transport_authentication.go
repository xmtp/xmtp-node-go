package auth

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	libp2pProtocol "github.com/libp2p/go-libp2p-core/protocol"
	"github.com/xmtp/go-msgio/protoio"
	"github.com/xmtp/xmtp-node-go/crypto"
	"github.com/xmtp/xmtp-node-go/logging"
	"github.com/xmtp/xmtp-node-go/protocol/pb"
	"github.com/xmtp/xmtp-node-go/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"math"
)

const TransportAuthID_v01beta1 = libp2pProtocol.ID("/xmtplabs/xmtp-v1/clientauth/0.1.0-beta1")

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrNoHandler        = errors.New("no handler found for given request version")
	ErrWalletMismatch   = errors.New("wallet address mismatch")
	ErrWrongPeerId      = errors.New("wrong peerId supplied")
)

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

	log := xmtpAuth.log.With(logging.HostID("peer", stream.Conn().RemotePeer()))
	log.Info("stream established")
	defer func() {
		if err := stream.Close(); err != nil {
			log.Error("closing stream", zap.Error(err))
		}
	}()

	authenticatedPeerId, authenticatedWalletAddr, err := xmtpAuth.handleRequest(stream, log)
	isAuthenticated := (err == nil) && authenticatedPeerId != types.InvalidPeerId() && authenticatedWalletAddr != types.InvalidWalletAddr()

	// TODO: Save PeerId to walletAddress map

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	err = xmtpAuth.writeAuthResponse(stream, isAuthenticated, errStr)
	if err != nil {
		log.Error("writing response", zap.Error(err))
		return
	}

	log.Info("Auth Result", logging.WalletAddressLabelled("walletAddr", authenticatedWalletAddr))

}

func (xmtpAuth *XmtpAuthentication) handleRequest(stream network.Stream, log *zap.Logger) (types.PeerId, types.WalletAddr, error) {

	authRequest, err := xmtpAuth.readAuthRequest(stream)
	if err != nil {
		log.Error("reading request", zap.Error(err))
		return types.InvalidPeerId(), types.InvalidWalletAddr(), err
	}

	var suppliedPeerId types.PeerId
	var walletAddr types.WalletAddr

	connectingPeerId := types.PeerId(stream.Conn().RemotePeer())

	switch version := authRequest.Version.(type) {
	case *pb.ClientAuthRequest_V1:
		suppliedPeerId, walletAddr, err = validateRequest(authRequest.GetV1(), connectingPeerId, log)
	default:
		xmtpAuth.log.Error("No handler for request", zap.Any("version", version))
		return types.InvalidPeerId(), types.InvalidWalletAddr(), ErrNoHandler
	}

	if err != nil {
		xmtpAuth.log.Error("validating request", zap.Error(err))
		return types.InvalidPeerId(), types.InvalidWalletAddr(), err
	}

	return suppliedPeerId, walletAddr, nil
}

func validateRequest(request *pb.V1ClientAuthRequest, connectingPeerId types.PeerId, log *zap.Logger) (types.PeerId, types.WalletAddr, error) {

	// Validate WalletSignature
	signingWalletAddress, err := recoverWalletAddress(request.IdentityKeyBytes, request.WalletSignature.GetEcdsaCompact())
	if err != nil {
		log.Error("verifying wallet signature", zap.Error(err))
		return types.InvalidPeerId(), types.InvalidWalletAddr(), err
	}

	// Validate AuthSignature
	suppliedPeerId, suppliedWalletAddress, err := verifyAuthSignature(request.IdentityKeyBytes, request.AuthDataBytes, request.AuthSignature.GetEcdsaCompact(), log)
	if err != nil {
		log.Error("verifying auth signature", zap.Error(err))
		return types.InvalidPeerId(), types.InvalidWalletAddr(), err
	}

	// To protect against spoofing, ensure the walletAddresses match in both signatures
	if signingWalletAddress != suppliedWalletAddress {
		log.Error("wallet address mismatch", zap.Error(err), logging.WalletAddressLabelled("recovered", signingWalletAddress), logging.WalletAddressLabelled("supplied", suppliedWalletAddress))
		return types.InvalidPeerId(), types.InvalidWalletAddr(), ErrWalletMismatch
	}

	// To protect against spoofing, ensure the AuthRequest originated from the same peerID that was authenticated.
	if connectingPeerId != suppliedPeerId {
		log.Error("peerId Mismatch", zap.Error(err), logging.HostID("supplied", peer.ID(suppliedPeerId)))
		return types.InvalidPeerId(), types.InvalidWalletAddr(), ErrWrongPeerId
	}

	return suppliedPeerId, signingWalletAddress, nil
}

func CreateIdentitySignRequest(identityBytes []byte) crypto.Message {
	return []byte(fmt.Sprintf("XMTP : Create Identity\n%s\n\nFor more info: https://xmtp.org/signatures/", hex.EncodeToString(identityBytes)))
}

func unpackAuthData(authDataBytes []byte) (*pb.AuthData, error) {
	authData := &pb.AuthData{}
	err := proto.Unmarshal(authDataBytes, authData)

	return authData, err
}

func recoverWalletAddress(identityKeyBytes []byte, signature *pb.Signature_ECDSACompact) (types.WalletAddr, error) {

	isrBytes := CreateIdentitySignRequest(identityKeyBytes)

	sig, err := crypto.SignatureFromBytes(signature.GetBytes())
	if err != nil {
		return types.InvalidWalletAddr(), err
	}
	return crypto.RecoverWalletAddress(isrBytes, sig, uint8(signature.GetRecovery()))
}

func verifyAuthSignature(identityKeyBytes []byte, authDataBytes []byte, authSig *pb.Signature_ECDSACompact, log *zap.Logger) (types.PeerId, types.WalletAddr, error) {

	pubKey := &pb.PublicKey{}
	err := proto.Unmarshal(identityKeyBytes, pubKey)
	if err != nil {
		return types.InvalidPeerId(), types.InvalidWalletAddr(), err
	}

	pub, err := crypto.PublicKeyFromBytes(pubKey.GetSecp256K1Uncompressed().Bytes)
	if err != nil {
		return types.InvalidPeerId(), types.InvalidWalletAddr(), err
	}

	signature, err := crypto.SignatureFromBytes(authSig.GetBytes())
	if err != nil {
		log.Error("signature decoding", zap.Error(err))
		return types.InvalidPeerId(), types.InvalidWalletAddr(), err
	}

	isVerified, err := crypto.Verify(pub, authDataBytes, signature)
	if err != nil {
		log.Error("signature verification", zap.Error(err))
		return types.InvalidPeerId(), types.InvalidWalletAddr(), err
	}

	if !isVerified {
		return types.InvalidPeerId(), types.InvalidWalletAddr(), ErrInvalidSignature
	}

	authData, err := unpackAuthData(authDataBytes)
	if err != nil {
		log.Error("unpacking auth data", zap.Error(err))
		return types.InvalidPeerId(), types.InvalidWalletAddr(), err
	}

	return types.PeerId(authData.PeerId), types.WalletAddr(authData.WalletAddr), nil
}

func (xmtpAuth *XmtpAuthentication) writeAuthResponse(stream network.Stream, isAuthenticated bool, errorString string) error {
	writer := protoio.NewDelimitedWriter(stream)
	authRespRPC := &pb.ClientAuthResponse{
		Version: &pb.ClientAuthResponse_V1{
			V1: &pb.V1ClientAuthResponse{
				AuthSuccessful: isAuthenticated,
				ErrorStr:       errorString,
			},
		},
	}
	return writer.WriteMsg(authRespRPC)
}

func (xmtpAuth *XmtpAuthentication) readAuthRequest(stream network.Stream) (*pb.ClientAuthRequest, error) {
	reader := protoio.NewDelimitedReader(stream, math.MaxInt32)
	authReqRPC := &pb.ClientAuthRequest{}
	err := reader.ReadMsg(authReqRPC)

	return authReqRPC, err
}