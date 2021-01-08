package grpc

import (
	"context"
	"errors"
	"log"
	"net"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"google.golang.org/grpc"
)

type Server struct {
	ch chan *pb.FieldsPBS
}

func (s *Server) Tracepoint(srv pb.TCPDog_TracepointServer) error {
	return errors.New("does not support")
}

func (s *Server) TracepointPBS(srv pb.TCPDog_TracepointPBSServer) error {
	for {
		fields, err := srv.Recv()
		if err != nil {
			return err
		}

		s.ch <- fields
	}
}

func Start(ctx context.Context, ch chan *pb.FieldsPBS) {
	l, err := net.Listen("tcp", ":8085")
	if err != nil {
		log.Fatal(err)
	}
	srv := Server{
		ch: ch,
	}

	gServer := grpc.NewServer()
	pb.RegisterTCPDogServer(gServer, &srv)

	go func() {
		log.Fatal(gServer.Serve(l))
	}()
}
