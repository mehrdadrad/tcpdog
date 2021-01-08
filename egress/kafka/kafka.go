package kafka

import (
	"bytes"
	"context"
	"sync"
)

// https://pkg.go.dev/google.golang.org/protobuf/encoding/protojson

func New(ctx context.Context, cfg map[string]string, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	return nil
}
