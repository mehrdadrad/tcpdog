package main

import (
	"C"
	"bytes"
	"sync"

	"github.com/sethvargo/go-signalcontext"

	"github.com/mehrdadrad/tcpdog/cli"
	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/ebpf"
	"github.com/mehrdadrad/tcpdog/output"
)
import (
	"log"
	"os"
)

func main() {
	r, err := cli.Get(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Get(r)
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
		if _, ok := chMap[tracepoint.Output]; ok {
			continue
		}

		ch := make(chan *bytes.Buffer, 1000)
		chMap[tracepoint.Output] = ch
		output.Start(ctx, tracepoint, bufPool, ch)
	}

	for index, tracepoint := range cfg.Tracepoints {
		e.Start(ctx, ebpf.TP{
			Name:    tracepoint.Name,
			Index:   index,
			BufPool: bufPool,
			OutChan: chMap[tracepoint.Output],
			INet:    tracepoint.Inet,
			Fields:  cfg.GetTPFields(tracepoint.Fields),
		})
	}

	<-ctx.Done()
}

// m := pb.TCPDog{}
// protojson.Unmarshal(v.Bytes(), &m)
// proto.Marshal(&m)
