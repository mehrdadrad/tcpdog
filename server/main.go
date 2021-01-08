package main

import (
	"log"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/ingest/influxdb"
	"github.com/mehrdadrad/tcpdog/server/ingress/grpc"
	"github.com/sethvargo/go-signalcontext"
)

func main() {
	var ch = make(chan *pb.FieldsPBS, 10000)

	cfg, err := config.Load("../config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	ctx = cfg.WithContext(ctx)

	grpc.Start(ctx, ch)
	influxdb.Start(ctx, ch)

	<-ctx.Done()
}
