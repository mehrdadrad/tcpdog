package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/ingestion/elasticsearch"
	"github.com/mehrdadrad/tcpdog/ingestion/influxdb"
	"github.com/mehrdadrad/tcpdog/ingress/grpc"
	"github.com/mehrdadrad/tcpdog/ingress/kafka"
)

func ingress(ctx context.Context, flow config.Flow, ch chan interface{}) {
	cfg := config.FromContextServer(ctx)
	logger := cfg.Logger()

	switch cfg.Ingress[flow.Ingress].Type {
	case "grpc":
		err := grpc.Start(ctx, flow.Ingress, ch)
		if err != nil {
			logger.Fatal("grpc", zap.Error(err))
		}

		logger.Info("grpc", zap.String("msg", flow.Ingress+" has been started"))

	case "kafka":
		err := kafka.Start(ctx, flow.Ingress, flow.Serialization, ch)
		if err != nil {
			logger.Fatal("kafka", zap.Error(err))
		}

		logger.Info("kafka", zap.String("msg", flow.Ingress+" has been started"))
	}
}

func ingestion(ctx context.Context, flow config.Flow, ch chan interface{}) {
	cfg := config.FromContextServer(ctx)
	logger := cfg.Logger()

	switch cfg.Ingestion[flow.Ingestion].Type {
	case "influxdb":
		err := influxdb.Start(ctx, flow.Ingestion, flow.Serialization, ch)
		if err != nil {
			logger.Fatal("influxdb", zap.Error(err))
		}

		logger.Info("influxdb", zap.String("msg", flow.Ingestion+" has been started"))

	case "elasticsearch":
		err := elasticsearch.Start(ctx, flow.Ingestion, flow.Serialization, ch)
		if err != nil {
			logger.Fatal("elasticsearch", zap.Error(err))
		}

		logger.Info("elasticsearch", zap.String("msg", flow.Ingestion+" has been started"))
	}
}

func validate(cfg *config.ServerConfig) error {
	return validateFlow(cfg)
}

func validateFlow(cfg *config.ServerConfig) error {
	if len(cfg.Flow) < 1 {
		return errors.New("there isn't any flow configuration")
	}

	for _, f := range cfg.Flow {
		if len(f.Ingestion) < 1 {
			return errors.New("ingestion hasn't configured")
		}

		if len(f.Ingress) < 1 {
			return errors.New("ingress hasn't configured")
		}

		if len(f.Serialization) < 1 {
			return errors.New("serialization hasn't configured")
		}

		if _, ok := cfg.Ingestion[f.Ingestion]; !ok {
			return fmt.Errorf("ingestion %s is not available", f.Ingestion)
		}

		if _, ok := cfg.Ingress[f.Ingress]; !ok {
			return fmt.Errorf("ingress %s is not available", f.Ingress)
		}
	}

	return nil
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
