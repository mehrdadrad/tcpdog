package main

import (
	"fmt"
	"os"

	"github.com/sethvargo/go-signalcontext"

	"github.com/mehrdadrad/tcpdog/server/cli"
	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/ingestion/elasticsearch"
	"github.com/mehrdadrad/tcpdog/server/ingestion/influxdb"
	"github.com/mehrdadrad/tcpdog/server/ingress/grpc"
	"github.com/mehrdadrad/tcpdog/server/ingress/kafka"
)

func main() {
	r, err := cli.Get(os.Args)
	if err != nil {
		exit(err)
	}

	cfg, err := config.Get(r)
	if err != nil {
		exit(err)
	}

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	ctx = cfg.WithContext(ctx)

	for _, flow := range cfg.Flow {
		ch := make(chan interface{}, 1000)

		switch cfg.Ingress[flow.Ingress].Type {
		case "grpc":
			grpc.Start(ctx, flow.Ingress, ch)
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

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
