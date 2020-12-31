package output

import (
	"bytes"
	"context"
	"log"
	"sync"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/output/console"
	"github.com/mehrdadrad/tcpdog/output/kafka"
)

// Create creates a new output based on the configuration
func Create(ctx context.Context, name string, bufpool *sync.Pool, ch chan *bytes.Buffer) {
	var (
		output config.OutputConfig
		err    error
		ok     bool
	)

	cfg := config.FromContext(ctx)
	if output, ok = cfg.Output[name]; !ok {
		log.Fatal("output not found")
	}

	switch output.Type {
	case "kafka":
		err = kafka.New(ctx, output.Config, bufpool, ch)
	case "grpc":
		// TODO
	case "csv":
		// TODO
	case "jsonl":
		// TODO
	default:
		err = console.New(ctx, output.Config, bufpool, ch)
	}

	if err != nil {
		log.Fatal(err)
	}
}
