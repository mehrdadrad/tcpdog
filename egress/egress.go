package egress

import (
	"bytes"
	"context"
	"sync"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/egress/console"
	"github.com/mehrdadrad/tcpdog/egress/csv"
	"github.com/mehrdadrad/tcpdog/egress/grpc"
	"github.com/mehrdadrad/tcpdog/egress/jsonl"
	"github.com/mehrdadrad/tcpdog/egress/kafka"
)

// Start starts an output based on the output type at configuration.
func Start(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	var err error

	cfg := config.FromContext(ctx)
	egress := cfg.Egress[tp.Egress]

	switch egress.Type {
	case "kafka":
		err = kafka.New(ctx, egress.Config, bufpool, ch)
	case "grpc":
		err = grpc.Start(ctx, egress.Config, bufpool, ch)
	case "grpc-spb":
		err = grpc.StartStructPB(ctx, tp, bufpool, ch)
	case "csv":
		err = csv.Start(ctx, tp, bufpool, ch)
	case "jsonl":
		err = jsonl.Start(ctx, tp, bufpool, ch)
	default:
		err = console.New(ctx, egress.Config, bufpool, ch)
	}

	return err
}
