package grpc

import (
	"bytes"
	"context"
	"sync"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

// Start sends fields to a grpc server
func Start(ctx context.Context, cfg map[string]string, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	conn, err := grpc.Dial(":8082", grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := pb.NewTCPDogClient(conn)
	stream, err := client.Send(context.Background())
	if err != nil {
		return err
	}

	go func() {
		var data *bytes.Buffer

		for {
			select {
			case data = <-ch:
				m := pb.Fields{}
				protojson.Unmarshal(data.Bytes(), &m)
				stream.Send(&m)
				bufpool.Put(data)
			case <-ctx.Done():
				stream.CloseAndRecv()
			}
		}
	}()

	return nil
}
