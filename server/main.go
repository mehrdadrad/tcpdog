package main

import (
	"log"

	"github.com/sethvargo/go-signalcontext"

	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/ingestion/elasticsearch"
	"github.com/mehrdadrad/tcpdog/server/ingestion/influxdb"
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

	for _, flow := range cfg.Flow {
		ch := make(chan interface{}, 1000)

		switch cfg.Ingress[flow.Ingress].Type {
		case "grpc":
			grpc.Start(ctx, ch)
		case "kafka":
			kafka.Start(ctx, flow.Ingress, flow.Serialization, ch)
		}

		switch cfg.Ingestion[flow.Ingestion].Type {
		case "influxdb":
			influxdb.Start(ctx, flow.Ingestion, flow.Serialization, ch)
		case "elasticsearch":
			elasticsearch.Start(ctx, flow.Ingestion, flow.Serialization, ch)
		}
	}

	<-ctx.Done()
}
