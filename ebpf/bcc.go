package ebpf

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"

	bpf "github.com/iovisor/gobpf/bcc"

	"github.com/mehrdadrad/tcpdog/config"
)

// BPF represents eBPF procedures.
type BPF struct {
	m        *bpf.Module
	perfMaps []*bpf.PerfMap
}

// TP represents a tracepoint
type TP struct {
	Name    string
	BufPool *sync.Pool
	OutChan chan *bytes.Buffer
	Index   int
	Workers int
	INet    []int
	Fields  []string
}

// New generates and loads the bpf program.
func New(conf *config.Config) *BPF {
	code, err := GetBPFCode(conf)
	if err != nil {
		log.Fatal(err)
	}

	m := bpf.NewModule(code, []string{})

	return &BPF{m: m}
}

// Start loads and attaches tracepoint and approperiate channel
func (b *BPF) Start(ctx context.Context, tp TP) {
	trace, err := b.m.LoadTracepoint(fmt.Sprintf("sk_trace%d", tp.Index))
	if err != nil {
		log.Fatal(err)
	}

	b.m.AttachTracepoint(tp.Name, trace)

	for _, version := range tp.INet {
		table := bpf.NewTable(b.m.TableId(fmt.Sprintf("ipv%d_events%d", version, tp.Index)), b.m)
		ch := make(chan []byte, 1000)

		perfMap, err := bpf.InitPerfMap(table, ch, nil)
		if err != nil {
			log.Fatal(err)
		}

		for i := 0; i < tp.Workers; i++ {
			go func(version int) {
				data := []byte{}
				d := newDecoder((version == 4))

				for {
					select {
					case data = <-ch:
					case <-ctx.Done():
						return
					}

					buf := tp.BufPool.Get().(*bytes.Buffer)
					buf.Reset()
					d.decode(data, tp.Fields, buf)

					select {
					case tp.OutChan <- buf:
					default:
						log.Println("output channel maxed out")
					}
				}
			}(version)
		}

		perfMap.Start()
		b.perfMaps = append(b.perfMaps, perfMap)
	}
}

// Close cleans up BPF attachments
func (b *BPF) Close() {
	for _, perfMap := range b.perfMaps {
		perfMap.Stop()
	}
	b.m.Close()
}
