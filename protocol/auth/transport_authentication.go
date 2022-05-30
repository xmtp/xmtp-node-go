package auth

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
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
	"strings"
)

const TransportAuthID_v01beta1 = libp2pProtocol.ID("/xmtplabs/xmtpv1/clientauth/0.1.0-beta1")

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrNoHandler        = errors.New("no handler found for given request version")
	ErrWalletMismatch   = errors.New("wallet address mismatch")
	ErrWrongPeerId      = errors.New("wrong peerId supplied")
)

type Message []byte

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
	defer func() {
		if err := stream.Close(); err != nil {
			xmtpAuth.log.Error("closing stream", zap.Error(err))
		}
	}()

	log := xmtpAuth.log.With(logging.HostID("peer", stream.Conn().RemotePeer()))
	log.Debug("stream established")

	authenticatedPeerId, authenticatedWalletAddr, err := xmtpAuth.handleRequest(stream, log)

	isAuthenticated := (err == nil) && authenticatedPeerId != "" && authenticatedWalletAddr != ""

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	err = xmtpAuth.WriteAuthResponse(stream, isAuthenticated, errStr)
	if err != nil {
		log.Error("writing response", zap.Error(err))
		return
	}

	log.Info("Auth Result", zap.Any("peerId", authenticatedPeerId), zap.Any("walletAddr", authenticatedWalletAddr))
	// TODO: Save PeerId to walletAddress map

}

func (xmtpAuth *XmtpAuthentication) handleRequest(stream network.Stream, log *zap.Logger) (peer.ID, types.WalletAddr, error) {

	authRequest, err := xmtpAuth.ReadAuthRequest(stream)
	if err != nil {
		log.Error("reading request", zap.Error(err))
		return "", "", err
	}

	var suppliedPeerId peer.ID
	var walletAddr types.WalletAddr

	connectingPeerId := stream.Conn().RemotePeer()

	switch version := authRequest.Version.(type) {
	case *pb.ClientAuthRequest_V1:
		suppliedPeerId, walletAddr, err = validateRequest(authRequest.GetV1(), connectingPeerId, log)
	default:
		xmtpAuth.log.Error("No handler for request", zap.Any("version", version))
		return "", "", ErrNoHandler
	}

	if err != nil {
		xmtpAuth.log.Error("validating request", zap.Error(err))
		return "", "", err
	}

	return suppliedPeerId, walletAddr, nil
}

func validateRequest(request *pb.V1ClientAuthRequest, connectingPeerId peer.ID, log *zap.Logger) (peer.ID, types.WalletAddr, error) {

	// Validate WalletSignature
	signingWalletAddress, err := recoverWalletAddress(request.IdentityKeyBytes, request.WalletSig.GetEcdsaCompact())
	if err != nil {
		log.Error("verifying wallet signature", zap.Error(err), zap.Any("AuthRequest", request))
		return "", "", err
	}

	// Validate AuthSignature
	suppliedPeerId, suppliedWalletAddress, err := VerifyAuthSignature(request.IdentityKeyBytes, peer.ID(request.PeerId), types.WalletAddr(request.WalletAddr), request.AuthSig.GetEcdsaCompact(), log)
	if err != nil {
		log.Error("verifying auth signature", zap.Error(err), zap.Any("AuthRequest", request))
		return "", "", err
	}

	// To protect against spoofing, ensure the walletAddresses match in both signatures
	if signingWalletAddress != suppliedWalletAddress {
		log.Error("wallet address mismatch", zap.Error(err), zap.Any("AuthRequest", request))
		return "", "", ErrWalletMismatch
	}

	// To protect against spoofing, ensure the AuthRequest originated from the same peerID that was authenticated.
	if connectingPeerId != suppliedPeerId {
		log.Error("peerId Mismatch", zap.Error(err), zap.String("expected", connectingPeerId.String()), zap.String("supplied", suppliedPeerId.String()))
		return "", "", ErrWrongPeerId
	}

	return suppliedPeerId, signingWalletAddress, nil
}

func CreateIdentitySignRequest(identityBytes []byte) Message {
	return []byte(fmt.Sprintf("XMTP : Create Identity\n%s\n\nFor more info: https://xmtp.org/signatures/", hex.EncodeToString(identityBytes)))
}

// EtherHash implements the Ethereum hashing standard used to create signatures
func EtherHash(msg Message) []byte {
	decoratedBytes := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg))
	return ethcrypto.Keccak256(decoratedBytes)
}

func recoverWalletAddress(identityKeyBytes []byte, signature *pb.Signature_ECDSACompact) (types.WalletAddr, error) {

	pubKey := &pb.PublicKey{}
	err := proto.Unmarshal(identityKeyBytes, pubKey)
	if err != nil {
		return "", err
	}

	signatureBytes := append(signature.GetBytes(), uint8(signature.GetRecovery()))

	isrBytes := CreateIdentitySignRequest(identityKeyBytes)
	digest := EtherHash(isrBytes)

	epk, err := ethcrypto.SigToPub(digest, signatureBytes)
	if err != nil {
		//TODO: Log
		return "", err
	}

	addr := ethcrypto.PubkeyToAddress(*epk)
	return types.WalletAddr(addr.String()), nil
}

func VerifyAuthSignature(identityKeyBytes []byte, peerId peer.ID, walletAddr types.WalletAddr, authSig *pb.Signature_ECDSACompact, log *zap.Logger) (peer.ID, types.WalletAddr, error) {

	pubKey := &pb.PublicKey{}
	err := proto.Unmarshal(identityKeyBytes, pubKey)
	if err != nil {
		return "", "", err
	}

	pub, err := crypto.BytesToPublicKey(pubKey.GetSecp256K1Uncompressed().Bytes)
	if err != nil {
		return "", "", err
	}

	messageStr := strings.Join([]string{string(peerId), string(walletAddr)}, "|")
	messageBytes := []byte(messageStr)

	signature, err := crypto.BytesToSignature(authSig.GetBytes())
	if err != nil {
		log.Error("signature decoding", zap.Error(err))
		return "", "", err
	}

	isVerified, err := crypto.Verify(pub, messageBytes, signature)
	if err != nil {
		log.Error("signature verification", zap.Error(err))
		return "", "", err
	}

	if !isVerified {
		return "", "", ErrInvalidSignature
	}

	return peerId, walletAddr, nil
}

func (xmtpAuth *XmtpAuthentication) WriteAuthResponse(stream network.Stream, isAuthenticated bool, errorString string) error {
	writer := protoio.NewDelimitedWriter(stream)
	authRespRPC := &pb.ClientAuthResponse{AuthSuccessful: isAuthenticated, ErrorStr: errorString}
	return writer.WriteMsg(authRespRPC)
}

func (xmtpAuth *XmtpAuthentication) ReadAuthRequest(stream network.Stream) (*pb.ClientAuthRequest, error) {
	reader := protoio.NewDelimitedReader(stream, math.MaxInt32)
	authReqRPC := &pb.ClientAuthRequest{}
	err := reader.ReadMsg(authReqRPC)

	return authReqRPC, err
}
