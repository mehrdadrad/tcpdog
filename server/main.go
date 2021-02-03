package main

import (
	"os"

	"github.com/sethvargo/go-signalcontext"
	"go.uber.org/zap"

	"github.com/mehrdadrad/tcpdog/config"
)

var version string

func main() {
	cfg, err := config.GetServer(os.Args, version)
	if err != nil {
		exit(err)
	}

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	ctx = cfg.WithContext(ctx)
	logger := cfg.Logger()

	logger.Info("tcpdog", zap.String("version", version), zap.String("type", "server"))

	for _, flow := range cfg.Flow {
		ch := make(chan interface{}, 1000)
		ingress(ctx, flow, ch)
		ingestion(ctx, flow, ch)
	}

	<-ctx.Done()
}
