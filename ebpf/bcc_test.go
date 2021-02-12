package ebpf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	cfg := &config.Config{
		Tracepoints: []config.Tracepoint{{
			Name:     "sock:inet_sock_set_state",
			Fields:   "fields01",
			TCPState: "TCP_CLOSE",
			Workers:  1,
			INet:     []int{4},
			Egress:   "console",
		}},
		Fields: map[string][]config.Field{
			"fields01": {{
				Name: "RTT",
			}},
		},
		Egress: map[string]config.EgressConfig{
			"console": {Type: "console"},
		},
	}

	cfg.SetMockLogger("memory")

	bufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	eBPF := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = cfg.WithContext(ctx)
	ch := make(chan *bytes.Buffer, 10)

	eBPF.Start(ctx, TP{
		Name:    "sock:inet_sock_set_state",
		BufPool: bufPool,
		OutChan: ch,
		Index:   0,
		Workers: 1,
		INet:    []int{4},
		Fields:  []string{"RTT"},
	})

	// create a tcp connection
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, ebpf")
	}))
	defer ts.Close()

	conn, err := net.Dial("tcp", ts.Listener.Addr().String())
	assert.NoError(t, err)
	conn.Close()

	select {
	case data := <-ch:
		v := struct {
			RTT       int
			Timestamp uint64
		}{}

		json.Unmarshal(data.Bytes(), &v)
		assert.Greater(t, v.RTT, 0)
		assert.Greater(t, v.Timestamp, uint64(1613009084))
	case <-time.After(time.Second * 5):
		t.Fatal("time exceeded")
	}
}
