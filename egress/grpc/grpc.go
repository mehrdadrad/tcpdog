package grpc

import (
	"bytes"
	"context"
	"os"
	"sync"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/egress/helper"
)

// StartStructPB sends fields to a grpc server with structpb type.
func StartStructPB(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	var (
		err    error
		stream pb.TCPDog_TracepointPBSClient
		conn   *grpc.ClientConn
	)

	cfg := config.FromContext(ctx)
	server, dialOpts := gRPCConfig(cfg.Egress[tp.Egress].Config)
	backoff := helper.NewBackoff(cfg)

	go func() {
		for {
			backoff.Next()

			conn, err = grpc.Dial(server, dialOpts...)
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
func Start(ctx context.Context, grpcConf map[string]string, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	var (
		err     error
		stream  pb.TCPDog_TracepointClient
		conn    *grpc.ClientConn
		backoff = helper.Backoff{}
	)

	server, dialOpts := gRPCConfig(grpcConf)

	go func() {
		for {
			backoff.Next()

			conn, err = grpc.Dial(server, dialOpts...)
			if err != nil {
				continue
			}

			client := pb.NewTCPDogClient(conn)
			stream, err = client.Tracepoint(ctx)
			if err != nil {
				continue
			}

			err = jsonpb(ctx, stream, bufpool, ch)
			if err != nil {
				continue
			}

			break
		}

	}()

	return nil
}

func gRPCConfig(cfg map[string]string) (string, []grpc.DialOption) {
	opts := []grpc.DialOption{}
	if cfg["insecure"] == "true" {
		opts = append(opts, grpc.WithInsecure())
	}

	return cfg["server"], opts
}
