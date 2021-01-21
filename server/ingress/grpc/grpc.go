package grpc

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"github.com/mehrdadrad/tcpdog/server/config"
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
func Start(ctx context.Context, name string, ch chan interface{}) {
	gCfg := grpcConfig(config.FromContext(ctx).Ingress[name].Config)
	logger := config.FromContext(ctx).Logger()

	l, err := net.Listen("tcp", gCfg.Addr)
	if err != nil {
		logger.Fatal("grpc", zap.Error(err))
	}
	srv := Server{
		ch:     ch,
		logger: logger,
	}

	opts, err := getServerOpts(gCfg)
	if err != nil {
		logger.Fatal("grpc", zap.Error(err))
	}

	gServer := grpc.NewServer(opts...)
	pb.RegisterTCPDogServer(gServer, &srv)

	go func() {
		err := gServer.Serve(l)
		logger.Fatal("grpc", zap.Error(err))
	}()
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

	return opts, nil
}
