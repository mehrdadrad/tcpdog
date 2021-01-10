package console

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/mehrdadrad/tcpdog/config"
)

// New encodes the tcp fields on the console.
func New(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	go func() {
		for {
			v := <-ch
			fmt.Println(string(v.Bytes()[1 : v.Len()-1]))
			bufpool.Put(v)
		}
	}()

	return nil
}
