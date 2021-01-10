package main

import (
	"C"
	"bytes"
	"os"
	"sync"

	"github.com/sethvargo/go-signalcontext"
	"go.uber.org/zap"

	"github.com/mehrdadrad/tcpdog/cli"
	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/ebpf"
	"github.com/mehrdadrad/tcpdog/egress"
)

func main() {
	r, err := cli.Get(os.Args)
	if err != nil {
		exit(err)
	}

	cfg := config.Get(r)
	err = validation(cfg)
	if err != nil {
		exit(err)
	}

	logger := cfg.Logger()

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	ctx = cfg.WithContext(ctx)

	e := ebpf.New(cfg)
	defer e.Close()

	bufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	chMap := map[string]chan *bytes.Buffer{}
	for _, tracepoint := range cfg.Tracepoints {
		if _, ok := chMap[tracepoint.Egress]; ok {
			continue
		}

		ch := make(chan *bytes.Buffer, 1000)
		chMap[tracepoint.Egress] = ch
		err := egress.Start(ctx, tracepoint, bufPool, ch)
		if err != nil {
			logger.Fatal("egress", zap.Error(err))
		}
	}

	for index, tracepoint := range cfg.Tracepoints {
		e.Start(ctx, ebpf.TP{
			Name:    tracepoint.Name,
			Index:   index,
			BufPool: bufPool,
			OutChan: chMap[tracepoint.Egress],
			INet:    tracepoint.Inet,
			Workers: tracepoint.Workers,
			Fields:  cfg.GetTPFields(tracepoint.Fields),
		})
	}

	<-ctx.Done()
}
