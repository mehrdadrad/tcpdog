package output

import (
	"bytes"
	"context"
	"log"
	"sync"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/output/console"
	"github.com/mehrdadrad/tcpdog/output/csv"
	"github.com/mehrdadrad/tcpdog/output/jsonl"
	"github.com/mehrdadrad/tcpdog/output/kafka"
)

// Start starts an output based on the output type at configuration.
func Start(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) {
	var (
		output config.OutputConfig
		err    error
		ok     bool
	)

	cfg := config.FromContext(ctx)
	if output, ok = cfg.Output[tp.Output]; !ok {
		log.Fatal("output not found:", tp.Output)
	}

	switch output.Type {
	case "kafka":
		err = kafka.New(ctx, output.Config, bufpool, ch)
	case "grpc":
		// TODO
	case "csv":
		err = csv.Start(ctx, tp, bufpool, ch)
	case "jsonl":
		err = jsonl.Start(ctx, tp, bufpool, ch)
	default:
		err = console.New(ctx, output.Config, bufpool, ch)
	}

	if err != nil {
		log.Fatal(err)
	}
}
