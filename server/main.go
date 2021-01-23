package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sethvargo/go-signalcontext"
	"go.uber.org/zap"

	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/ingestion/elasticsearch"
	"github.com/mehrdadrad/tcpdog/server/ingestion/influxdb"
	"github.com/mehrdadrad/tcpdog/server/ingress/grpc"
	"github.com/mehrdadrad/tcpdog/server/ingress/kafka"
)

var version string

func main() {
	cfg, err := config.Get(os.Args, version)
	if err != nil {
		exit(err)
	}

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	ctx = cfg.WithContext(ctx)
	logger := cfg.Logger()

	logger.Info("tcpdog", zap.String("version", version))

	for _, flow := range cfg.Flow {
		ch := make(chan interface{}, 1000)
		ingress(ctx, flow, ch)
		ingestion(ctx, flow, ch)
	}

	<-ctx.Done()
}

func ingress(ctx context.Context, flow config.Flow, ch chan interface{}) {
	cfg := config.FromContext(ctx)
	logger := cfg.Logger()

	switch cfg.Ingress[flow.Ingress].Type {
	case "grpc":
		err := grpc.Start(ctx, flow.Ingress, ch)
		if err != nil {
			logger.Fatal("grpc", zap.Error(err))
		}

		logger.Info("grpc", zap.String("msg", "grpc server has been started"))

	case "kafka":
		err := kafka.Start(ctx, flow.Ingress, flow.Serialization, ch)
		if err != nil {
			logger.Fatal("kafka", zap.Error(err))
		}

		logger.Info("kafka", zap.String("msg", "consumer has been started"))
	}
}

func ingestion(ctx context.Context, flow config.Flow, ch chan interface{}) {
	cfg := config.FromContext(ctx)
	logger := cfg.Logger()

	switch cfg.Ingestion[flow.Ingestion].Type {
	case "influxdb":
		err := influxdb.Start(ctx, flow.Ingestion, flow.Serialization, ch)
		if err != nil {
			logger.Fatal("influxdb", zap.Error(err))
		}

		logger.Info("influxdb", zap.String("msg", "client has been started"))

	case "elasticsearch":
		err := elasticsearch.Start(ctx, flow.Ingestion, flow.Serialization, ch)
		if err != nil {
			logger.Fatal("elasticsearch", zap.Error(err))
		}

		logger.Info("elasticsearch", zap.String("msg", "client has been started"))
	}
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
