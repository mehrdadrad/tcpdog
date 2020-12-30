package main

import (
	"C"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/ebpf"
)
import (
	"bytes"
	"fmt"
)

func main() {
	cfg := config.Load()
	sig := make(chan struct{})

	e := ebpf.New(cfg)
	defer e.Close()

	ch := make(chan *bytes.Buffer, 1000)
	for index, tracepoint := range cfg.Tracepoints {
		e.Start(ebpf.TP{
			Name:    tracepoint.Name,
			Index:   index,
			OutChan: ch,
			Workers: 0,
			INet:    tracepoint.Inet,
			Fields:  cfg.GetTPFields(tracepoint.Fields),
		})
	}

	go func() {
		for {
			v := <-ch
			fmt.Println(v.String())
		}
	}()

	<-sig
}
