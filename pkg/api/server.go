package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	grpcvalidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	proto "github.com/xmtp/proto/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	messagev1 "github.com/xmtp/xmtp-node-go/pkg/api/message/v1"
)

type Server struct {
	*Parameters

	grpcListener net.Listener
	httpListener net.Listener
	cert         *tls.Certificate
	certPool     *x509.CertPool
	messagev1    *messagev1.Service
	wg           sync.WaitGroup
	ctx          context.Context
}

func New(config *Parameters) (*Server, error) {
	if err := config.check(); err != nil {
		return nil, err
	}

	s := &Server{
		Parameters: config,
	}

	// Initialize TLS.
	var err error
	s.cert, s.certPool, err = generateCert()
	if err != nil {
		return nil, errors.Wrap(err, "generating tls")
	}

	s.ctx = context.Background()

	// Start gRPC services.
	err = s.startGRPC()
	if err != nil {
		return nil, err
	}

	// Start HTTP gateway.
	err = s.startHTTP()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) startGRPC() error {
	var err error

	s.grpcListener, err = net.Listen("tcp", addrString(s.GRPCAddress, s.GRPCPort))
	if err != nil {
		return errors.Wrap(err, "creating grpc listener")
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(credentials.NewServerTLSFromCert(s.cert)),
		grpc.UnaryInterceptor(grpcvalidator.UnaryServerInterceptor()),
		grpc.StreamInterceptor(grpcvalidator.StreamServerInterceptor()),
	)

	s.messagev1, err = messagev1.NewService(s.Waku, s.Log)
	if err != nil {
		return errors.Wrap(err, "creating message service")
	}
	proto.RegisterMessageApiServer(grpcServer, s.messagev1)

	tracing.GoPanicWrap(s.ctx, &s.wg, "grpc", func(ctx context.Context) {
		s.Log.Info("serving grpc", zap.String("address", s.grpcListener.Addr().String()))
		err := grpcServer.Serve(s.grpcListener)
		if err != nil && !isErrUseOfClosedConnection(err) {
			s.Log.Error("serving grpc", zap.Error(err))
		}
	})

	return nil
}

func (s *Server) startHTTP() error {

	mux := http.NewServeMux()
	gwmux := runtime.NewServeMux(
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		runtime.WithStreamErrorHandler(runtime.DefaultStreamErrorHandler),
	)
	mux.Handle("/", gwmux)

	conn, err := s.dialGRPC(s.ctx)
	if err != nil {
		return errors.Wrap(err, "dialing grpc server")
	}

	err = proto.RegisterMessageApiHandler(s.ctx, gwmux, conn)
	if err != nil {
		return errors.Wrap(err, "registering message handler")
	}

	addr := addrString(s.HTTPAddress, s.HTTPPort)
	s.httpListener, err = net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "creating grpc-gateway listener")
	}

	server := http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*s.cert},
		},
		Handler: allowCORS(mux),
	}

	tracing.GoPanicWrap(s.ctx, &s.wg, "http", func(ctx context.Context) {
		s.Log.Info("serving http", zap.String("address", s.httpListener.Addr().String()))
		err = server.ServeTLS(s.httpListener, "", "")
		if err != nil && err != http.ErrServerClosed && !isErrUseOfClosedConnection(err) {
			s.Log.Error("serving http", zap.Error(err))
		}
	})

	return nil
}

func (s *Server) Close() {
	s.Log.Info("closing")
	if s.messagev1 != nil {
		s.messagev1.Close()
	}

	if s.httpListener != nil {
		err := s.httpListener.Close()
		if err != nil {
			s.Log.Error("closing http listener", zap.Error(err))
		}
	}

	if s.grpcListener != nil {
		err := s.grpcListener.Close()
		if err != nil {
			s.Log.Error("closing grpc listener", zap.Error(err))
		}
	}

	(&s.wg).Wait()
	s.Log.Info("closed")
}

func (s *Server) dialGRPC(ctx context.Context) (*grpc.ClientConn, error) {
	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	dialAddr := fmt.Sprintf("passthrough://localhost/%s", s.grpcListener.Addr().String())
	return grpc.DialContext(
		ctx,
		dialAddr,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(s.certPool, "")),
		// grpc.WithBlock(),
	)
}

func isErrUseOfClosedConnection(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
