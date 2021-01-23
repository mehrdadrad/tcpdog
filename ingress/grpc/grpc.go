package grpc

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"github.com/mehrdadrad/tcpdog/config"
	pb "github.com/mehrdadrad/tcpdog/proto"
)

// Server represents gRPC server
type Server struct {
	ch     chan interface{}
	logger *zap.Logger
}

// Tracepoint receives protobuf messages
func (s *Server) Tracepoint(srv pb.TCPDog_TracepointServer) error {
	for {
		fields, err := srv.Recv()
		if err != nil {
			return err
		}

		select {
		case s.ch <- fields:
		default:
			s.logger.Error("grpc", zap.String("msg", "data has been dropped"))
		}
	}
}

// TracepointSPB receives struct protobuf messages
func (s *Server) TracepointSPB(srv pb.TCPDog_TracepointSPBServer) error {
	for {
		fields, err := srv.Recv()
		if err != nil {
			return err
		}

		select {
		case s.ch <- fields:
		default:
			s.logger.Error("grpc", zap.String("msg", "data has been dropped"))
		}
	}
}

// Start starts gRPC server
func Start(ctx context.Context, name string, ch chan interface{}) error {
	gCfg := grpcConfig(config.FromContextServer(ctx).Ingress[name].Config)
	logger := config.FromContextServer(ctx).Logger()

	l, err := net.Listen("tcp", gCfg.Addr)
	if err != nil {
		return err
	}
	srv := Server{
		ch:     ch,
		logger: logger,
	}

	opts, err := getServerOpts(gCfg)
	if err != nil {
		return err
	}

	gServer := grpc.NewServer(opts...)
	pb.RegisterTCPDogServer(gServer, &srv)

	go func() {
		err := gServer.Serve(l)
		logger.Fatal("grpc", zap.Error(err))
	}()

	return nil
}

func getServerOpts(gCfg *Config) ([]grpc.ServerOption, error) {
	var opts []grpc.ServerOption

	if gCfg.TLSConfig != nil && gCfg.TLSConfig.Enable {
		creds, err := config.GetCreds(gCfg.TLSConfig)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}

	opts = append(opts, grpc.StreamInterceptor(serverInterceptor))

	return opts, nil
}

func serverInterceptor(srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if p, ok := peer.FromContext(ss.Context()); ok {
		srv.(*Server).logger.Info("grpc.connect", zap.String("peer", p.Addr.String()))
	}

	return handler(srv, ss)
}
