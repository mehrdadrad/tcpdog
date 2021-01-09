package console

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/mehrdadrad/tcpdog/config"
)

// New ...
func New(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	go func() {
		for {
			v := <-ch
			fmt.Println(v.String())
			bufpool.Put(v)
		}
	}()

	return nil
}
