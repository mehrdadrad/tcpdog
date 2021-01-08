package console

import (
	"bytes"
	"context"
	"fmt"
	"sync"
)

// New ...
func New(ctx context.Context, cfg map[string]string, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	go func() {
		for {
			v := <-ch
			fmt.Println(v.String())
			bufpool.Put(v)
		}
	}()

	return nil
}
