package authn

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/xmtp/xmtp-node-go/pkg/crypto"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidPeerId     = errors.New("invalid peerId")
	ErrInvalidWalletAddr = errors.New("invalid walletAddress")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrNoHandler         = errors.New("no handler found for given request version")
	ErrWalletMismatch    = errors.New("wallet address mismatch")
	ErrWrongPeerId       = errors.New("wrong peerId supplied")
)

type XmtpAuthentication struct {
	ctx context.Context
	log *zap.Logger
}

func NewXmtpAuthentication(ctx context.Context, log *zap.Logger) *XmtpAuthentication {
	xmtpAuth := new(XmtpAuthentication)
	xmtpAuth.ctx = ctx
	xmtpAuth.log = log.Named("client-authn")

	return xmtpAuth
}

func (xmtpAuth *XmtpAuthentication) Start() {
	xmtpAuth.log.Info("Auth protocol started")
}

func validateRequest(ctx context.Context, request *V1ClientAuthRequest, connectingPeerId types.PeerId) (peer types.PeerId, wallet types.WalletAddr, err error) {
	logger := logging.From(ctx)

	// Validate WalletSignature
	recoveredWalletAddress, err := recoverWalletAddress(request.IdentityKeyBytes, request.WalletSignature.GetEcdsaCompact())
	if err != nil {
		logger.Error("verifying wallet signature", zap.Error(err))
		return peer, wallet, err
	}

	// Validate AuthSignature
	suppliedPeerId, suppliedWalletAddress, err := verifyAuthSignature(ctx, request.IdentityKeyBytes, request.AuthDataBytes, request.AuthSignature.GetEcdsaCompact())
	if err != nil {
		logger.Error("verifying authn signature", zap.Error(err))
		return peer, wallet, err
	}

	// To protect against spoofing, ensure the walletAddresses match in both signatures
	if recoveredWalletAddress != suppliedWalletAddress {
		logger.Error("wallet address mismatch", zap.Error(err), zap.String("recovered", recoveredWalletAddress.String()), zap.String("supplied", suppliedWalletAddress.String()))
		return peer, wallet, ErrWalletMismatch
	}

	// To protect against spoofing, ensure the AuthRequest originated from the same peerID that was authenticated.
	if connectingPeerId != suppliedPeerId {
		logger.Error("peerId Mismatch", zap.Error(err), zap.String("supplied", suppliedPeerId.Raw().String()))
		return peer, wallet, ErrWrongPeerId
	}

	return suppliedPeerId, recoveredWalletAddress, nil
}

func CreateIdentitySignRequest(identityBytes []byte) crypto.Message {
	return []byte(fmt.Sprintf("XMTP : Create Identity\n%s\n\nFor more info: https://xmtp.org/signatures/", hex.EncodeToString(identityBytes)))
}

func unpackAuthData(authDataBytes []byte) (*AuthData, error) {
	authData := &AuthData{}
	err := proto.Unmarshal(authDataBytes, authData)

	return authData, err
}

func recoverWalletAddress(identityKeyBytes []byte, signature *Signature_ECDSACompact) (wallet types.WalletAddr, err error) {

	isrBytes := CreateIdentitySignRequest(identityKeyBytes)

	sig, err := crypto.SignatureFromBytes(signature.GetBytes())
	if err != nil {
		return wallet, err
	}
	return crypto.RecoverWalletAddress(isrBytes, sig, uint8(signature.GetRecovery()))
}

func verifyAuthSignature(ctx context.Context, identityKeyBytes []byte, authDataBytes []byte, authSig *Signature_ECDSACompact) (peer types.PeerId, wallet types.WalletAddr, err error) {
	logger := logging.From(ctx)

	pubKey := &PublicKey{}
	err = proto.Unmarshal(identityKeyBytes, pubKey)
	if err != nil {
		return peer, wallet, err
	}

	pub, err := crypto.PublicKeyFromBytes(pubKey.GetSecp256K1Uncompressed().Bytes)
	if err != nil {
		return peer, wallet, err
	}

	signature, err := crypto.SignatureFromBytes(authSig.GetBytes())
	if err != nil {
		logger.Error("signature decoding", zap.Error(err))
		return peer, wallet, err
	}

	isVerified := crypto.Verify(pub, authDataBytes, signature)
	if !isVerified {
		return peer, wallet, ErrInvalidSignature
	}

	authData, err := unpackAuthData(authDataBytes)
	if err != nil {
		logger.Error("unpacking authn data", zap.Error(err))
		return peer, wallet, err
	}

	return types.PeerId(authData.PeerId), types.WalletAddr(authData.WalletAddr), nil
}
