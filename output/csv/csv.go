package csv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/mehrdadrad/tcpdog/config"
)

type csv struct {
	fieldsLen  []int
	fieldsName []string
	file       io.WriteCloser
	buffer     *bytes.Buffer
}

func (c *csv) init(conf map[string]string, fields []config.Field) error {
	var err error

	for _, f := range fields {
		c.fieldsLen = append(c.fieldsLen, len(f.Name)+3)
		c.fieldsName = append(c.fieldsName, f.Name)
	}

	filename := conf["filename"]
	c.file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	return err
}

func (c *csv) marshal(buf *bytes.Buffer) {
	buf.Next(1) // skip bracket

	for _, l := range c.fieldsLen {
		buf.Next(l)
		v, _ := buf.ReadBytes([]byte(",")[0])
		c.buffer.Write(v)
	}

	buf.Next(12)                 // skip timestamp key
	c.buffer.Write(buf.Next(10)) // write timestamp
}

func (c *csv) header() {
	m := strings.Join(c.fieldsName, ",")
	c.buffer.WriteString(fmt.Sprintf("%s,timestamp", m))
}
func (c *csv) flush() {
	c.buffer.WriteRune('\n')
	c.file.Write(c.buffer.Bytes())
	c.buffer.Reset()
}

func (c *csv) cleanup() {
	c.file.Close()
}

// Start encodes and writes tcp fields to a specific file in csv format
func Start(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	var (
		c   = &csv{buffer: new(bytes.Buffer)}
		err error
	)

	cfg := config.FromContext(ctx)
	err = c.init(cfg.Output[tp.Output].Config, cfg.Fields[tp.Fields])
	if err != nil {
		return err
	}

	c.header()
	c.flush()

	go func() {
		defer c.cleanup()
		buf := new(bytes.Buffer)

		for {
			select {
			case buf = <-ch:
			case <-ctx.Done():
				return
			}

			c.marshal(buf)
			c.flush()

			bufpool.Put(buf)
		}
	}()

	return nil
}
