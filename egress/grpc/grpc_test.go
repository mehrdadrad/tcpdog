package grpc

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/mehrdadrad/tcpdog/config"
	pb "github.com/mehrdadrad/tcpdog/proto"
)

var (
	srv  = server{}
	port int
)

type server struct {
	ch1 *pb.Fields
	ch2 *pb.FieldsPBS
}

func (s *server) Tracepoint(srv pb.TCPDog_TracepointServer) error {

	for {
		f, err := srv.Recv()
		if err != nil {
			return err
		}

		s.ch1 = f
	}
}

func (s *server) TracepointPBS(srv pb.TCPDog_TracepointPBSServer) error {
	for {
		f, err := srv.Recv()
		if err != nil {
			return err
		}

		s.ch2 = f
	}
}

func TestGRPC(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	port = l.Addr().(*net.TCPAddr).Port

	srv = server{}
	gServer := grpc.NewServer()
	pb.RegisterTCPDogServer(gServer, &srv)
	go gServer.Serve(l)

	t.Run("StructPB", testStructPB)
	t.Run("testProtoJSON", testProtoJSON)

	t.Cleanup(func() { l.Close() })
}

func testStructPB(t *testing.T) {
	ch := make(chan *bytes.Buffer, 1)
	bufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	tp := config.Tracepoint{Egress: "foo", Fields: "fields01"}
	cfg := config.Config{
		Fields: map[string][]config.Field{
			"fields01": {
				{Name: "F1"},
				{Name: "F2"},
			},
		},
		Egress: map[string]config.EgressConfig{
			"foo": {Config: map[string]string{
				"server":   fmt.Sprintf(":%d", port),
				"insecure": "true"}},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx = cfg.WithContext(ctx)
	ch <- bytes.NewBufferString(`{"F1":5,"F2":6,"Timestamp":1609564925}`)
	StartStructPB(ctx, tp, bufPool, ch)
	time.Sleep(time.Second)

	hostname, _ := os.Hostname()

	assert.NotNil(t, srv.ch2)
	assert.Equal(t, 5.0, srv.ch2.Fields.Fields["F1"].GetNumberValue())
	assert.Equal(t, 6.0, srv.ch2.Fields.Fields["F2"].GetNumberValue())
	assert.Equal(t, hostname, srv.ch2.Fields.Fields["Hostname"].GetStringValue())
	assert.Equal(t, 1609564925.0, srv.ch2.Fields.Fields["Timestamp"].GetNumberValue())

	cancel()
	time.Sleep(time.Second)
}

func testProtoJSON(t *testing.T) {
	ch := make(chan *bytes.Buffer, 1)
	ctx, cancel := context.WithCancel(context.Background())
	bufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	grpcConf := map[string]string{
		"server":   fmt.Sprintf(":%d", port),
		"insecure": "true",
	}

	ch <- bytes.NewBufferString(`{"SRTT":5,"AdvMSS":6,"Timestamp":1609564925}`)
	Start(ctx, grpcConf, bufPool, ch)
	time.Sleep(time.Second)

	hostname, _ := os.Hostname()

	assert.NotNil(t, srv.ch1)
	assert.Equal(t, int32(5), srv.ch1.SRTT)
	assert.Equal(t, int32(6), srv.ch1.AdvMSS)
	assert.Equal(t, hostname, srv.ch1.Hostname)
	assert.Equal(t, int64(1609564925), srv.ch1.Timestamp)

	cancel()
	time.Sleep(time.Second)
}
