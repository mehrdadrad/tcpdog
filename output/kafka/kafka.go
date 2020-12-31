package kafka

import (
	"bytes"
	"context"
	"sync"
)

func New(ctx context.Context, cfg map[string]string, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	return nil
}
