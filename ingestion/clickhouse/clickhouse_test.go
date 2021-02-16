package clickhouse

import (
	"encoding/json"
	"net"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/geo"
	pb "github.com/mehrdadrad/tcpdog/proto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type geoMock struct{}

func (g *geoMock) Init(l *zap.Logger, cfg map[string]string) {}
func (g *geoMock) Get(s string) map[string]string            { return map[string]string{"City": "Los_Angeles"} }

func TestStart(t *testing.T) {
	geo.Reg["foo"] = &geoMock{}

	cfg := &config.ServerConfig{
		Geo: config.Geo{
			Type: "foo",
		},
		Ingress: map[string]config.Ingress{},
		Ingestion: map[string]config.Ingestion{
			"foo": {
				Config: map[string]interface{}{
					"DSName":        "tcp://127.0.0.1:9999?username=&debug=true",
					"FlushInterval": 1,
					"Connections":   1,
					"Workers":       1,
					"BatchSize":     1,
					"fields":        []string{"RTT", "SAddr"},
					"columns":       []string{"rtt", "saddr"},
				},
			},
		},
	}

	cfg.SetMockLogger("memory")

	var ln net.Listener
	go func() {
		ln, _ = net.Listen("tcp", "127.0.0.1:9999")
		conn, _ := ln.Accept()

		conn.Write([]byte{0x0, 0x9})
		data := []byte("MockHouse")
		conn.Write(data)
		conn.Write([]byte{0x5, 0x5, 0x5}) // version

		conn.Write([]byte{0x4}) // pong
		time.Sleep(time.Second * 1)

		conn.Write([]byte{0x1})
		conn.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
		conn.Write([]byte{0x2, 0x0})

		conn.Write([]byte{0x3})
		conn.Write([]byte("rtt"))
		conn.Write([]byte{0x6})
		conn.Write([]byte("UInt32"))

		conn.Write([]byte{0x5})
		conn.Write([]byte("saddr"))
		conn.Write([]byte{0x6})
		conn.Write([]byte("String"))

		time.Sleep(time.Second * 2)

		conn.Write([]byte{0x4})

		time.Sleep(time.Second * 5)
	}()
	time.Sleep(time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = cfg.WithContext(ctx)

	ch := make(chan interface{}, 10)

	Start(ctx, "foo", "json", ch)
	time.Sleep(time.Second * 2)

	m := map[string]interface{}{}
	b := []byte(`{"RTT":12345,"SAddr":"10.0.0.1"}`)
	json.Unmarshal(b, &m)

	ch <- m
	time.Sleep(time.Second * 3)

	cancel()
	time.Sleep(time.Second * 3)

	ln.Close()

	// server down
	ctx = cfg.WithContext(context.Background())
	err := Start(ctx, "foo", "json", ch)
	assert.Error(t, err)
}

func TestJSON(t *testing.T) {
	c := clickhouse{
		geo:           &geoMock{},
		cfg:           &chConfig{GeoField: "SAddr", Fields: []string{"RTT", "SAddr", "Timestamp", "Hostname", "City"}},
		serialization: "JSON",
		vFields:       reflect.ValueOf(&pb.Fields{}).Elem(),
	}

	m := map[string]interface{}{}
	b := []byte(`{"RTT":12345,"SAddr":"10.0.0.1","Timestamp":1611118090,"Hostname":"foo"}`)
	json.Unmarshal(b, &m)

	r, err := c.JSON(m)
	assert.NoError(t, err)

	assert.Equal(t, uint32(12345), r[0].(uint32))
	assert.Equal(t, "10.0.0.1", r[1].(string))
	assert.Equal(t, uint64(1611118090), r[2].(uint64))
	assert.Equal(t, "foo", r[3].(string))
	assert.Equal(t, "Los_Angeles", r[4].(string))
}

func TestPB(t *testing.T) {
	c := clickhouse{
		geo:           &geoMock{},
		cfg:           &chConfig{GeoField: "SAddr", Fields: []string{"RTT", "SAddr", "Timestamp", "Hostname", "City"}},
		serialization: "PB",
		vFields:       reflect.ValueOf(&pb.Fields{}).Elem(),
	}

	b := []byte(`{"RTT":12345,"SAddr":"10.0.0.1","Timestamp":1611118090,"Hostname":"foo"}`)
	p := pb.Fields{}
	protojson.Unmarshal(b, &p)

	r, err := c.PB(&p)
	assert.NoError(t, err)

	assert.Equal(t, uint32(12345), r[0].(uint32))
	assert.Equal(t, "10.0.0.1", r[1].(string))
	assert.Equal(t, uint64(1611118090), r[2].(uint64))
	assert.Equal(t, "foo", r[3].(string))
	assert.Equal(t, "Los_Angeles", r[4].(string))
}

func TestSPB(t *testing.T) {
	c := clickhouse{
		geo:           &geoMock{},
		cfg:           &chConfig{GeoField: "SAddr", Fields: []string{"RTT", "SAddr", "Timestamp", "Hostname", "City"}},
		serialization: "SPB",
		vFields:       reflect.ValueOf(&pb.Fields{}).Elem(),
	}

	m := map[string]interface{}{}
	b := []byte(`{"RTT":12345,"SAddr":"10.0.0.1","Timestamp":1611118090,"Hostname":"foo"}`)
	json.Unmarshal(b, &m)
	spb, err := structpb.NewStruct(m)
	assert.NoError(t, err)

	r, err := c.SPB((&pb.FieldsSPB{Fields: spb}))
	assert.NoError(t, err)

	assert.Equal(t, uint32(12345), r[0].(uint32))
	assert.Equal(t, "10.0.0.1", r[1].(string))
	assert.Equal(t, uint64(1611118090), r[2].(uint64))
	assert.Equal(t, "foo", r[3].(string))
	assert.Equal(t, "Los_Angeles", r[4].(string))
}

func TestGetSliceIfMaker(t *testing.T) {
	c := clickhouse{serialization: "pb"}
	expected := runtime.FuncForPC(reflect.ValueOf(c.PB).Pointer()).Name()
	actual := runtime.FuncForPC(reflect.ValueOf(c.getSliceIfMaker()).Pointer()).Name()
	assert.Equal(t, expected, actual)

	c.serialization = "spb"
	expected = runtime.FuncForPC(reflect.ValueOf(c.SPB).Pointer()).Name()
	actual = runtime.FuncForPC(reflect.ValueOf(c.getSliceIfMaker()).Pointer()).Name()
	assert.Equal(t, expected, actual)

	c.serialization = "unknown"
	assert.Nil(t, c.getSliceIfMaker())
}

func BenchmarkReflect(b *testing.B) {
	v := reflect.ValueOf(&pb.Fields{}).Elem()
	name := "RTT"
	for i := 0; i < b.N; i++ {
		switch v.FieldByName(name).Type().Elem().Kind() {
		case reflect.Uint32:
			//a[i] = uint32(f[name].(float64))
		case reflect.Uint64:
			//a[i] = uint64(f[name].(float64))
		case reflect.String:
			//a[i] = f[name]
		}
	}
}

func BenchmarkTypeMap(b *testing.B) {
	m := map[string]int{"RTT": 0}
	name := "RTT"
	for i := 0; i < b.N; i++ {
		switch m[name] {
		case 0:
		case 1:
		case 2:
		}
	}
}
