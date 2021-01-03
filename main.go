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
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	r, err := cli.Get(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Get(r)
	err = validation(cfg)
	if err != nil {
		log.Fatal(err)
	}

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
			Workers: tracepoint.Workers,
			Fields:  cfg.GetTPFields(tracepoint.Fields),
		})
	}

	<-ctx.Done()
}

func validation(cfg *config.Config) error {
	for i, tp := range cfg.Tracepoints {
		// fields validation
		err := validationFields(cfg, tp.Fields)
		if err != nil {
			return err
		}

		// tcpstatus validation
		s, err := ebpf.ValidateTCPStatus(tp.TCPState)
		if err != nil {
			return err
		}
		cfg.Tracepoints[i].TCPState = s

		// tracepoint
		err = ebpf.ValidateTracepoint(tp.Name)
		if err != nil {
			return err
		}

		// inet validation and default
		// TODO
		// output validation and default
		// TODO
	}

	return nil
}

func validationFields(cfg *config.Config, name string) error {
	if _, ok := cfg.Fields[name]; !ok {
		return fmt.Errorf("%s not exist", name)
	}

	for i, f := range cfg.Fields[name] {
		cf, err := ebpf.ValidateField(f.Name)
		if err != nil {
			return err
		}

		cfg.Fields[name][i].Name = cf
		cfg.Fields[name][i].Filter = strings.Replace(f.Filter, f.Name, cf, -1)
	}
	return nil
}

// m := pb.TCPDog{}
// protojson.Unmarshal(v.Bytes(), &m)
// proto.Marshal(&m)
