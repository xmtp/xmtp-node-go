package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/pkg/errors"
	swgui "github.com/swaggest/swgui/v3"
	wakupb "github.com/waku-org/go-waku/waku/v2/protocol/pb"
	wakurelay "github.com/waku-org/go-waku/waku/v2/protocol/relay"
	identityv1pb "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	proto "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	mlsv1pb "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	messagev1openapi "github.com/xmtp/xmtp-node-go/pkg/proto/openapi"
	"github.com/xmtp/xmtp-node-go/pkg/ratelimiter"
	"github.com/xmtp/xmtp-node-go/pkg/topic"
	"github.com/xmtp/xmtp-node-go/pkg/tracing"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pires/go-proxyproto"
	messagev1 "github.com/xmtp/xmtp-node-go/pkg/api/message/v1"
	apicontext "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/context"
	identityv1 "github.com/xmtp/xmtp-node-go/pkg/identity/api/v1"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/mls/api/v1"
)

const (
	authorizationMetadataKey = "authorization"
)

var prometheusOnce sync.Once

type Server struct {
	*Config

	grpcListener net.Listener
	httpListener net.Listener
	messagev1    *messagev1.Service
	mlsv1        *mlsv1.Service
	identityv1   *identityv1.Service
	wg           sync.WaitGroup
	ctx          context.Context
	ctxCancel    func()
	natsServer   *server.Server
	wakuRelaySub *wakurelay.Subscription

	authorizer *WalletAuthorizer
}

func New(config *Config) (*Server, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	s := &Server{
		Config: config,
	}

	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	// Start gRPC services.
	err := s.startGRPC()
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

	grpcListener, err := net.Listen("tcp", addrString(s.GRPCAddress, s.GRPCPort))
	s.grpcListener = &proxyproto.Listener{
		Listener:          grpcListener,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err != nil {
		return errors.Wrap(err, "creating grpc listener")
	}

	prometheusOnce.Do(func() {
		prometheus.EnableHandlingTimeHistogram()
	})
	unary := []grpc.UnaryServerInterceptor{prometheus.UnaryServerInterceptor}
	stream := []grpc.StreamServerInterceptor{prometheus.StreamServerInterceptor}

	telemetryInterceptor := NewTelemetryInterceptor(s.Log)
	gatingInterceptor := NewGatingInterceptor(s.Log)
	unary = append(unary, telemetryInterceptor.Unary(), gatingInterceptor.Unary())
	stream = append(stream, telemetryInterceptor.Stream(), gatingInterceptor.Stream())

	// Initialize nats for API subscribers.
	s.natsServer, err = server.NewServer(&server.Options{
		Port: server.RANDOM_PORT,
	})
	if err != nil {
		return err
	}
	go s.natsServer.Start()
	if !s.natsServer.ReadyForConnections(4 * time.Second) {
		return errors.New("nats not ready")
	}

	if s.Authn.Enable {
		limiter := ratelimiter.NewTokenBucketRateLimiter(s.ctx, s.Log)
		// Expire buckets after 1 hour of inactivity,
		// sweep for expired buckets every 10 minutes.
		// Note: entry expiration should be at least some multiple of
		// maximum (limit max / limit rate) minutes.
		go limiter.Janitor(10*time.Minute, 1*time.Hour)
		s.authorizer = NewWalletAuthorizer(&AuthnConfig{
			AuthnOptions: s.Authn,
			Limiter:      limiter,
			AllowLister:  s.AllowLister,
			Log:          s.Log.Named("authn"),
		})
		unary = append(unary, s.authorizer.Unary())
		stream = append(stream, s.authorizer.Stream())
	}

	options := []grpc.ServerOption{
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(middleware.ChainUnaryServer(unary...)),
		grpc.StreamInterceptor(middleware.ChainStreamServer(stream...)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: 5 * time.Minute,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
			MinTime:             15 * time.Second,
		}),
		grpc.MaxRecvMsgSize(s.MaxMsgSize),
	}
	grpcServer := grpc.NewServer(options...)
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(grpcServer, healthcheck)

	publishToWakuRelay := func(ctx context.Context, msg *wakupb.WakuMessage) error {
		_, err := s.Waku.Relay().Publish(ctx, msg)
		return err
	}

	s.messagev1, err = messagev1.NewService(s.Log, s.Store, s.natsServer, publishToWakuRelay)
	if err != nil {
		return errors.Wrap(err, "creating message service")
	}
	proto.RegisterMessageApiServer(grpcServer, s.messagev1)

	// Enable the MLS and identity servers if a store is provided
	if s.MLSStore != nil && s.MLSValidator != nil && s.EnableMls {
		s.mlsv1, err = mlsv1.NewService(
			s.Log,
			s.MLSStore,
			s.MLSValidator,
			s.natsServer,
			publishToWakuRelay,
		)
		if err != nil {
			return errors.Wrap(err, "creating mls service")
		}
		mlsv1pb.RegisterMlsApiServer(grpcServer, s.mlsv1)

		s.identityv1, err = identityv1.NewService(s.Log, s.MLSStore, s.MLSValidator)
		if err != nil {
			return errors.Wrap(err, "creating identity service")
		}
		identityv1pb.RegisterIdentityApiServer(grpcServer, s.identityv1)
	}

	// Initialize waku relay subscription.
	s.wakuRelaySub, err = s.Waku.Relay().Subscribe(s.ctx)
	if err != nil {
		return errors.Wrap(err, "subscribing to relay")
	}
	tracing.GoPanicWrap(s.ctx, &s.wg, "broadcast", func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case wakuEnv := <-s.wakuRelaySub.Ch:
				if wakuEnv == nil || wakuEnv.Message() == nil {
					continue
				}
				wakuMsg := wakuEnv.Message()

				if topic.IsMLSV1(wakuMsg.ContentTopic) {
					if s.mlsv1 != nil {
						err := s.mlsv1.HandleIncomingWakuRelayMessage(wakuEnv.Message())
						if err != nil {
							s.Log.Error(
								"error handling waku relay message by mlsv1 service",
								zap.Error(err),
							)
						}
					}
				} else {
					if s.messagev1 != nil {
						err := s.messagev1.HandleIncomingWakuRelayMessage(wakuEnv.Message())
						if err != nil {
							s.Log.Error("error handling waku relay message by messagev1 service", zap.Error(err))
						}
					}
				}

			}
		}
	})

	prometheus.Register(grpcServer)

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
		runtime.WithMarshalerOption("application/x-protobuf", &runtime.ProtoMarshaller{}),
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		runtime.WithStreamErrorHandler(runtime.DefaultStreamErrorHandler),
		runtime.WithIncomingHeaderMatcher(incomingHeaderMatcher),
	)

	swaggerUI := swgui.NewHandler("API", "/swagger.json", "/")
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/" || strings.HasPrefix(r.URL.Path, "/swagger-ui"):
			swaggerUI.ServeHTTP(w, r)
		case r.URL.Path == "/swagger.json":
			_, err := w.Write(messagev1openapi.JSON)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		default:
			gwmux.ServeHTTP(w, r)
		}
	})

	conn, err := s.dialGRPC(s.ctx)
	if err != nil {
		return errors.Wrap(err, "dialing grpc server")
	}

	err = proto.RegisterMessageApiHandler(s.ctx, gwmux, conn)
	if err != nil {
		return errors.Wrap(err, "registering message handler")
	}

	if s.MLSStore != nil && s.EnableMls {
		err = mlsv1pb.RegisterMlsApiHandler(s.ctx, gwmux, conn)
		if err != nil {
			return errors.Wrap(err, "registering mls handler")
		}

		err = identityv1pb.RegisterIdentityApiHandler(s.ctx, gwmux, conn)
		if err != nil {
			return errors.Wrap(err, "registering identity handler")
		}
	}

	addr := addrString(s.HTTPAddress, s.HTTPPort)
	s.httpListener, err = net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "creating grpc-gateway listener")
	}

	// Add two handler wrappers to mux: gzipWrapper and allowCORS
	server := http.Server{
		Addr:    addr,
		Handler: allowCORS(gzipWrapper(mux)),
	}

	tracing.GoPanicWrap(s.ctx, &s.wg, "http", func(ctx context.Context) {
		s.Log.Info("serving http", zap.String("address", s.httpListener.Addr().String()))
		if s.httpListener == nil {
			s.Log.Fatal("no http listener")
		}
		err = server.Serve(s.httpListener)
		if err != nil && err != http.ErrServerClosed && !isErrUseOfClosedConnection(err) {
			s.Log.Error("serving http", zap.Error(err))
		}
	})

	return nil
}

func (s *Server) Close() {
	s.Log.Info("closing")

	if s.ctxCancel != nil {
		s.ctxCancel()
	}

	if s.wakuRelaySub != nil {
		s.wakuRelaySub.Unsubscribe()
	}

	if s.messagev1 != nil {
		s.messagev1.Close()
	}
	if s.mlsv1 != nil {
		s.mlsv1.Close()
	}

	if s.natsServer != nil {
		s.natsServer.Shutdown()
	}

	if s.httpListener != nil {
		err := s.httpListener.Close()
		if err != nil {
			s.Log.Error("closing http listener", zap.Error(err))
		}
		s.httpListener = nil
	}

	if s.grpcListener != nil {
		err := s.grpcListener.Close()
		if err != nil {
			s.Log.Error("closing grpc listener", zap.Error(err))
		}
		s.grpcListener = nil
	}

	s.wg.Wait()
	s.Log.Info("closed")
}

func (s *Server) dialGRPC(ctx context.Context) (*grpc.ClientConn, error) {
	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	dialAddr := fmt.Sprintf("passthrough://localhost/%s", s.grpcListener.Addr().String())
	return grpc.DialContext(
		ctx,
		dialAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(s.MaxMsgSize),
		),
	)
}

func (s *Server) httpListenAddr() string {
	return "http://" + s.httpListener.Addr().String()
}

func isErrUseOfClosedConnection(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{
		"Content-Type",
		"Accept",
		"Authorization",
		"X-App-Version",
		"X-Client-Version",
		"X-Libxmtp-Version",
		"Baggage",
		"DNT",
		"Sec-CH-UA",
		"Sec-CH-UA-Mobile",
		"Sec-CH-UA-Platform",
		"Sentry-Trace",
		"User-Agent",
		"x-libxmtp-version",
		"x-app-version",
	}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
			preflightHandler(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func incomingHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case apicontext.AppVersionMetadataKey:
		return key, true
	case apicontext.ClientVersionMetadataKey:
		return key, true
	case apicontext.LibxmtpVersionMetadataKey:
		return key, true
	default:
		return key, false
	}
}
