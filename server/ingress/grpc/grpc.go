package grpc

import (
	"context"
	"log"
	"net"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"google.golang.org/grpc"
)

type Server struct {
	ch chan interface{}
}

func (s *Server) Tracepoint(srv pb.TCPDog_TracepointServer) error {
	for {
		fields, err := srv.Recv()
		if err != nil {
			return err
		}

		s.ch <- fields
	}
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

func Start(ctx context.Context, ch chan interface{}) {
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
