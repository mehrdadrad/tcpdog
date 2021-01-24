package grpc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

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

type statsHandler struct {
	remoteAddr net.Addr
	logger     *zap.Logger
}

func (h *statsHandler) TagRPC(ctx context.Context, s *stats.RPCTagInfo) context.Context {
	return ctx
}

func (h *statsHandler) HandleRPC(context.Context, stats.RPCStats) {}

func (h *statsHandler) TagConn(ctx context.Context, s *stats.ConnTagInfo) context.Context {
	h.remoteAddr = s.RemoteAddr
	return ctx
}

func (h *statsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnEnd:
		h.logger.Info("grpc", zap.String("msg", fmt.Sprintf("%s has been disconnected", h.remoteAddr)))
	case *stats.ConnBegin:
		h.logger.Info("grpc", zap.String("msg", fmt.Sprintf("%s has been connected", h.remoteAddr)))
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

	opts, err := getServerOpts(gCfg, logger)
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

func getServerOpts(gCfg *Config, logger *zap.Logger) ([]grpc.ServerOption, error) {
	var opts []grpc.ServerOption

	if gCfg.TLSConfig != nil && gCfg.TLSConfig.Enable {
		creds, err := config.GetCreds(gCfg.TLSConfig)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}

	opts = append(opts, grpc.StatsHandler(&statsHandler{logger: logger}))

	return opts, nil
}
