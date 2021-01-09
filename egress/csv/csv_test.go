package csv

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	tp := config.Tracepoint{
		Egress: "myegress",
		Fields: "myfields",
	}
	ch := make(chan *bytes.Buffer, 1)
	bufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	filename := t.TempDir() + "testfile.jsonl"

	cfg := config.Config{
		Egress: map[string]config.EgressConfig{
			"myegress": {
				Type: "jsonl",
				Config: map[string]interface{}{
					"filename": filename,
				},
			},
		},
		Fields: map[string][]config.Field{
			"myfields": {
				{Name: "F1"},
				{Name: "F2"},
			},
		},
	}

	ctx = cfg.WithContext(ctx)

	go Start(ctx, tp, bufPool, ch)

	b := new(bytes.Buffer)
	b.WriteString(`{"F1":5,"F2":6,"Timestamp":1609564925}`)
	ch <- b
	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)
	f, err := os.Open(filename)
	assert.NoError(t, err)
	fb, err := ioutil.ReadAll(f)

	assert.NoError(t, err)
	assert.Equal(t, "F1,F2,timestamp\n5,6,1609564925\n", string(fb))

	cfg = config.Config{
		Egress: map[string]config.EgressConfig{
			"myegress": {
				Type: "jsonl",
				Config: map[string]interface{}{
					"filename": "",
				},
			},
		},
	}
	ctx = cfg.WithContext(context.Background())
	err = Start(ctx, tp, bufPool, ch)
	assert.Error(t, err)
}
