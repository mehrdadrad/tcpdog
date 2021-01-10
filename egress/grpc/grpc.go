package grpc

import (
	"bytes"
	"context"
	"os"
	"sync"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/egress/helper"
)

// StartStructPB sends fields to a grpc server with structpb type.
func StartStructPB(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	var (
		stream pb.TCPDog_TracepointPBSClient
		conn   *grpc.ClientConn
	)

	cfg := config.FromContext(ctx)
	backoff := helper.NewBackoff(cfg)

	gCfg, err := gRPCConfig(cfg.Egress[tp.Egress].Config)
	if err != nil {
		return err
	}

	opts := dialOpts(gCfg)

	go func() {
		for {
			backoff.Next()

			conn, err = grpc.Dial(gCfg.Server, opts...)
			if err != nil {
				continue
			}

			client := pb.NewTCPDogClient(conn)
			stream, err = client.TracepointPBS(ctx)
			if err != nil {
				continue
			}

			err = structpb(ctx, stream, tp, bufpool, ch)
			if err != nil {
				continue
			}

			break
		}

	}()

	return nil
}

func structpb(ctx context.Context, stream pb.TCPDog_TracepointPBSClient, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	var (
		cfg = config.FromContext(ctx)
		spb = helper.NewStructPB(cfg.Fields[tp.Fields])
		buf *bytes.Buffer
		err error
	)

	for {
		select {
		case buf = <-ch:
			err = stream.Send(&pb.FieldsPBS{
				Fields: spb.Unmarshal(buf),
			})
			if err != nil {
				return err
			}

			bufpool.Put(buf)
		case <-ctx.Done():
			stream.CloseAndRecv()
			return nil
		}
	}
}

func jsonpb(ctx context.Context, stream pb.TCPDog_TracepointClient, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	var buf *bytes.Buffer

	hostname, _ := os.Hostname()

	for {
		select {
		case buf = <-ch:
			m := pb.Fields{}
			protojson.Unmarshal(buf.Bytes(), &m)
			m.Hostname = hostname
			if err := stream.Send(&m); err != nil {
				return err
			}

			bufpool.Put(buf)
		case <-ctx.Done():
			stream.CloseAndRecv()
			return nil
		}
	}
}

// Start sends fields to a grpc server
func Start(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	var (
		stream pb.TCPDog_TracepointClient
		conn   *grpc.ClientConn
	)

	cfg := config.FromContext(ctx)
	logger := cfg.Logger()
	backoff := helper.NewBackoff(cfg)

	gCfg, err := gRPCConfig(cfg.Egress[tp.Egress].Config)
	if err != nil {
		return err
	}

	opts := dialOpts(gCfg)

	go func() {
		for {
			backoff.Next()

			conn, err = grpc.Dial(gCfg.Server, opts...)
			if err != nil {
				logger.Warn("grpc", zap.Error(err))
				continue
			}

			client := pb.NewTCPDogClient(conn)
			stream, err = client.Tracepoint(ctx)
			if err != nil {
				logger.Warn("grpc", zap.Error(err))
				continue
			}

			err = jsonpb(ctx, stream, bufpool, ch)
			if err != nil {
				logger.Warn("grpc", zap.Error(err))
				continue
			}

			break
		}

	}()

	return nil
}

type grpcConf struct {
	Server   string
	Insecure bool
}

func gRPCConfig(cfg map[string]interface{}) (*grpcConf, error) {
	// default config
	gCfg := &grpcConf{
		Server:   "localhost:8085",
		Insecure: true,
	}

	if err := config.Transform(cfg, gCfg); err != nil {
		return nil, err
	}

	return gCfg, nil
}

func dialOpts(gCfg *grpcConf) []grpc.DialOption {
	var opts []grpc.DialOption

	if gCfg.Insecure {
		opts = append(opts, grpc.WithInsecure())
	}

	return opts
}
