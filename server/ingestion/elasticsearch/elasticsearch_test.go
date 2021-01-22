package elasticsearch

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"testing"
	"time"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/geo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type geoMock struct{}

func (g *geoMock) Init(l *zap.Logger, cfg map[string]string) {}
func (g *geoMock) Get(s string) map[string]string            { return map[string]string{"City": "Los_Angeles"} }

func TestStart(t *testing.T) {
	done := make(chan struct{})

	geo.Reg["foo"] = &geoMock{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		expected := `{"index":{}}
{"Hostname":"foo","PID":123456,"RTT":12345,"Task":"curl","Timestamp":1611118090}
`
		assert.Equal(t, expected, string(body))

		w.Write([]byte(`{}`))
		done <- struct{}{}
	}))

	cfg := &config.Config{
		Geo: config.Geo{
			Type: "foo",
		},
		Ingress: map[string]config.Ingress{},
		Ingestion: map[string]config.Ingestion{
			"foo": {
				Config: map[string]interface{}{
					"urls":          []string{server.URL},
					"FlushInterval": 1,
					"Index":         "foo",
				},
			},
		},
	}

	cfg.SetMockLogger("memory")

	ctx, cancel := context.WithCancel(context.Background())
	ctx = cfg.WithContext(ctx)
	ch := make(chan interface{}, 1)

	Start(ctx, "foo", "json", ch)

	m := map[string]interface{}{}
	b := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"Timestamp":1611118090,"Hostname":"foo"}`)
	json.Unmarshal(b, &m)
	ch <- m

	<-done
	cancel()

	time.Sleep(time.Second)
	server.Close()
}

func TestItemJSON(t *testing.T) {
	e := &elastic{geo: &geoMock{}, cfg: &esConfig{GeoField: "SAddr"}}

	m := map[string]interface{}{}
	b := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"SAddr":"10.0.0.1","Timestamp":1611118090,"Hostname":"foo"}`)
	json.Unmarshal(b, &m)

	item, err := e.itemJSON(m)
	assert.NoError(t, err)

	b, err = ioutil.ReadAll(item.Body)
	assert.NoError(t, err)

	f := pb.Fields{}
	protojson.Unmarshal(b, &f)

	assert.Equal(t, "Los_Angeles", *f.City)
	assert.Equal(t, "foo", *f.Hostname)
	assert.Equal(t, int32(123456), *f.PID)
	assert.Equal(t, int32(12345), *f.RTT)
	assert.Equal(t, "10.0.0.1", *f.SAddr)
	assert.Equal(t, "curl", *f.Task)
	assert.Equal(t, int64(1611118090), *f.Timestamp)
}

func TestItemSPB(t *testing.T) {
	e := &elastic{geo: &geoMock{}, cfg: &esConfig{GeoField: "SAddr"}}

	m := map[string]interface{}{}
	b := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"SAddr":"10.0.0.1","Timestamp":1611118090,"Hostname":"foo"}`)
	json.Unmarshal(b, &m)
	spb, err := structpb.NewStruct(m)
	assert.NoError(t, err)

	item, err := e.itemSPB(&pb.FieldsSPB{Fields: spb})
	assert.NoError(t, err)

	b, err = ioutil.ReadAll(item.Body)
	assert.NoError(t, err)

	f := pb.Fields{}
	protojson.Unmarshal(b, &f)

	assert.Equal(t, "Los_Angeles", *f.City)
	assert.Equal(t, "foo", *f.Hostname)
	assert.Equal(t, int32(123456), *f.PID)
	assert.Equal(t, int32(12345), *f.RTT)
	assert.Equal(t, "10.0.0.1", *f.SAddr)
	assert.Equal(t, "curl", *f.Task)
	assert.Equal(t, int64(1611118090), *f.Timestamp)

}

func TestItemPB(t *testing.T) {
	e := &elastic{geo: &geoMock{}, cfg: &esConfig{GeoField: "SAddr"}}

	b := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"SAddr":"10.0.0.1","Timestamp":1611118090,"Hostname":"foo"}`)
	p := pb.Fields{}
	protojson.Unmarshal(b, &p)

	item, err := e.itemPB(&p)
	assert.NoError(t, err)

	b, err = ioutil.ReadAll(item.Body)
	assert.NoError(t, err)

	f := pb.Fields{}
	protojson.Unmarshal(b, &f)

	assert.Equal(t, "Los_Angeles", *f.City)
	assert.Equal(t, "foo", *f.Hostname)
	assert.Equal(t, int32(123456), *f.PID)
	assert.Equal(t, int32(12345), *f.RTT)
	assert.Equal(t, "10.0.0.1", *f.SAddr)
	assert.Equal(t, "curl", *f.Task)
	assert.Equal(t, int64(1611118090), *f.Timestamp)
}

func TestGetItemMaker(t *testing.T) {
	i := &elastic{}
	expected := runtime.FuncForPC(reflect.ValueOf(i.itemSPB).Pointer()).Name()
	actual := runtime.FuncForPC(reflect.ValueOf(i.getItemMaker("spb")).Pointer()).Name()
	assert.Equal(t, expected, actual)

	expected = runtime.FuncForPC(reflect.ValueOf(i.itemPB).Pointer()).Name()
	actual = runtime.FuncForPC(reflect.ValueOf(i.getItemMaker("pb")).Pointer()).Name()
	assert.Equal(t, expected, actual)

	expected = runtime.FuncForPC(reflect.ValueOf(i.itemJSON).Pointer()).Name()
	actual = runtime.FuncForPC(reflect.ValueOf(i.getItemMaker("json")).Pointer()).Name()
	assert.Equal(t, expected, actual)

	assert.Nil(t, i.getItemMaker("unknown"))
}
