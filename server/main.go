package main

import (
	"log"

	"github.com/sethvargo/go-signalcontext"

	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/ingest/influxdb"
	"github.com/mehrdadrad/tcpdog/server/ingress/grpc"
	"github.com/mehrdadrad/tcpdog/server/ingress/kafka"
)

func main() {

	cfg, err := config.Load("../config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	ctx = cfg.WithContext(ctx)

	for _, f := range cfg.Flow {
		ch := make(chan interface{}, 1000)

		switch cfg.Ingress[f.Ingress].Type {
		case "grpc":
			grpc.Start(ctx, ch)
		case "kafka":
			kafka.Start(ctx, f.Ingress, f.Serialization, ch)
		}

		influxdb.Start(ctx, f.Ingestion, f.Serialization, ch)
	}

	<-ctx.Done()
}
