package grpc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/mehrdadrad/tcpdog/config"
	pb "github.com/mehrdadrad/tcpdog/proto"
)

func TestStart(t *testing.T) {
	// server
	cfg := config.ServerConfig{
		Ingress: map[string]config.Ingress{
			"foo": {
				Type: "grpc",
				Config: map[string]interface{}{
					"addr": ":8085",
				},
			},
		},
	}

	ms := cfg.SetMockLogger("memory")

	ctx, cancel := context.WithCancel(context.Background())
	ctx = cfg.WithContext(ctx)
	ch := make(chan interface{}, 1)

	Start(ctx, "foo", ch)

	// client
	conn, err := grpc.Dial("localhost:8085", grpc.WithInsecure())
	assert.NoError(t, err)
	client := pb.NewTCPDogClient(conn)

	// SPB
	streamSPB, err := client.TracepointSPB(ctx)
	assert.NoError(t, err)

	m := map[string]interface{}{
		"PID":       float64(123456),
		"Task":      "curl",
		"RTT":       float64(12345),
		"SAddr":     "10.0.0.1",
		"Timestamp": float64(1611118090),
	}

	spb, err := structpb.NewStruct(m)
	assert.NoError(t, err)

	err = streamSPB.Send(&pb.FieldsSPB{
		Fields: spb,
	})
	assert.NoError(t, err)

	select {
	case a := <-ch:
		b, err := a.(*pb.FieldsSPB).Fields.MarshalJSON()
		assert.NoError(t, err)

		mm := map[string]interface{}{}
		err = json.Unmarshal(b, &mm)
		assert.NoError(t, err)
		assert.Equal(t, m, mm)
	case <-time.After(time.Second):
		t.Fatal("time exceeded")
	}

	// drop SPB
	ms.Reset()
	for i := 0; i < 2; i++ {
		err = streamSPB.Send(&pb.FieldsSPB{Fields: spb})
		assert.NoError(t, err)
	}
	time.Sleep(time.Second)
	assert.Equal(t, "data has been dropped", ms.Unmarshal()["msg"])

	<-ch // empty the channel

	// PB
	rtt := uint32(10)
	task := "curl"
	timestamp := uint64(1611118090)

	streamPB, err := client.Tracepoint(ctx)
	assert.NoError(t, err)

	err = streamPB.Send(&pb.Fields{
		RTT:       &rtt,
		Task:      &task,
		Timestamp: &timestamp,
	})
	assert.NoError(t, err)

	select {
	case a := <-ch:
		assert.Equal(t, uint32(10), *a.(*pb.Fields).RTT)
		assert.Equal(t, uint64(1611118090), *a.(*pb.Fields).Timestamp)
		assert.Equal(t, "curl", *a.(*pb.Fields).Task)
	case <-time.After(time.Second):
		t.Fatal("time exceeded")
	}

	// drop PB
	ms.Reset()
	for i := 0; i < 2; i++ {
		err = streamPB.Send(&pb.Fields{})
		assert.NoError(t, err)
	}
	time.Sleep(time.Second)
	assert.Equal(t, "data has been dropped", ms.Unmarshal()["msg"])

	// recv err
	cancel()
	time.Sleep(time.Second)
}
