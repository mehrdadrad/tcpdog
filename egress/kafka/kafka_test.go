package kafka

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/mehrdadrad/tcpdog/config"
	pb "github.com/mehrdadrad/tcpdog/proto"
)

func TestStartJSON(t *testing.T) {
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

	seedBroker := sarama.NewMockBroker(t, 1)
	defer seedBroker.Close()
	seedBroker.Returns(new(sarama.MetadataResponse))

	cfg := config.Config{
		Egress: map[string]config.EgressConfig{
			"myegress": {
				Type: "kafka",
				Config: map[string]interface{}{
					"brokers":        []string{seedBroker.Addr()},
					"RequestSizeMax": 104857600,
					"RetryMax":       2,
					"RetryBackoff":   10,
					"Workers":        2,
					"Serialization":  "json",
					"Topic":          "tcpdog",
					"TLSSkipVerify":  true,
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

	err := Start(ctx, tp, bufPool, ch)
	assert.NoError(t, err)

	data := `{"F1":5,"F2":6,"Timestamp":1609564925}`
	buf0 := bufPool.Get().(*bytes.Buffer)
	buf0.WriteString(data)
	ch <- buf0

	time.Sleep(time.Second)

	// check through recycled buffer
	buf1 := bufPool.Get().(*bytes.Buffer)
	assert.Equal(t, data, buf1.String())

	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestStartSPB(t *testing.T) {
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

	seedBroker := sarama.NewMockBroker(t, 1)
	defer seedBroker.Close()
	seedBroker.Returns(new(sarama.MetadataResponse))

	cfg := config.Config{
		Egress: map[string]config.EgressConfig{
			"myegress": {
				Type: "kafka",
				Config: map[string]interface{}{
					"brokers":        []string{seedBroker.Addr()},
					"RequestSizeMax": 104857600,
					"RetryMax":       2,
					"RetryBackoff":   10,
					"Workers":        2,
					"Serialization":  "spb",
					"Topic":          "tcpdog",
					"TLSSkipVerify":  true,
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

	err := Start(ctx, tp, bufPool, ch)
	assert.NoError(t, err)

	data := `{"F1":5,"F2":6,"Timestamp":1609564925}`
	buf := bufPool.Get().(*bytes.Buffer)
	buf.WriteString(data)
	ch <- buf

	time.Sleep(time.Second)

	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestStartPB(t *testing.T) {
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

	seedBroker := sarama.NewMockBroker(t, 1)
	defer seedBroker.Close()
	seedBroker.Returns(new(sarama.MetadataResponse))

	cfg := config.Config{
		Egress: map[string]config.EgressConfig{
			"myegress": {
				Type: "kafka",
				Config: map[string]interface{}{
					"brokers":        []string{seedBroker.Addr()},
					"RequestSizeMax": 104857600,
					"RetryMax":       2,
					"RetryBackoff":   10,
					"Workers":        2,
					"Serialization":  "pb",
					"Topic":          "tcpdog",
					"TLSSkipVerify":  true,
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

	err := Start(ctx, tp, bufPool, ch)
	assert.NoError(t, err)

	data := `{"F1":5,"F2":6,"Timestamp":1609564925}`
	buf := bufPool.Get().(*bytes.Buffer)
	buf.WriteString(data)
	ch <- buf

	time.Sleep(time.Second)

	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestWorkerSPB(t *testing.T) {
	bufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	k := kafka{
		dCh:     make(chan *bytes.Buffer, 1),
		bCh:     make(chan []byte, 1),
		bufpool: bufPool,
	}

	cfg := config.Config{}
	ctx := cfg.WithContext(context.Background())

	go k.workerSPB(ctx, []config.Field{{Name: "F1"}, {Name: "F2"}})
	k.dCh <- bytes.NewBufferString(`{"F1":5,"F2":6,"Timestamp":1609564925}`)

	time.Sleep(time.Second)

	b := <-k.bCh
	spb := pb.FieldsSPB{}
	err := proto.Unmarshal(b, &spb)
	assert.NoError(t, err)

	assert.Contains(t, spb.Fields.Fields, "F1")
	assert.Contains(t, spb.Fields.Fields, "F2")
	assert.Contains(t, spb.Fields.Fields, "Timestamp")
	assert.Contains(t, spb.Fields.Fields, "Hostname")

	assert.Equal(t, float64(5), spb.Fields.Fields["F1"].GetNumberValue())
	assert.Equal(t, float64(6), spb.Fields.Fields["F2"].GetNumberValue())
	assert.Equal(t, float64(1609564925), spb.Fields.Fields["Timestamp"].GetNumberValue())
}

func TestWorkerPB(t *testing.T) {
	bufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	k := kafka{
		dCh:     make(chan *bytes.Buffer, 1),
		bCh:     make(chan []byte, 1),
		bufpool: bufPool,
	}

	cfg := config.Config{}
	ctx := cfg.WithContext(context.Background())

	go k.workerPB(ctx, []config.Field{{Name: "RTT"}, {Name: "AdvMSS"}})
	k.dCh <- bytes.NewBufferString(`{"RTT":5,"AdvMSS":1400,"Timestamp":1609564925}`)

	time.Sleep(time.Second)

	b := <-k.bCh
	p := pb.Fields{}
	err := proto.Unmarshal(b, &p)
	assert.NoError(t, err)

	assert.Equal(t, int32(5), *p.RTT)
	assert.Equal(t, int32(1400), *p.AdvMSS)
	assert.Equal(t, int64(1609564925), *p.Timestamp)
}
