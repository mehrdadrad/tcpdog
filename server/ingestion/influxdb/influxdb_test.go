package influxdb

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
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/geo"
)

type geoMock struct{}

func (g *geoMock) Init(l *zap.Logger, cfg map[string]string) {}
func (g *geoMock) Get(s string) map[string]string            { return map[string]string{"City": "Los_Angeles"} }

func TestStart(t *testing.T) {
	done := make(chan struct{})

	geo.Reg["foo"] = &geoMock{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		expected := "tcpdog,Hostname=foo,Task=curl PID=123456,RTT=12345 1611118090000000000\n"
		assert.Equal(t, expected, string(body))
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
					"url": server.URL,
				},
			},
		},
	}
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

func TestPointJSON(t *testing.T) {
	i := &influxdb{geo: &geoMock{}, cfg: &dbConfig{GeoField: "SAddr"}}

	m := map[string]interface{}{}
	b := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"SAddr":"10.0.0.1","Timestamp":1611118090,"Hostname":"foo"}`)
	json.Unmarshal(b, &m)

	point := i.pointJSON(m)

	assert.Len(t, point.TagList(), 3)
	assert.Equal(t, "City", point.TagList()[0].Key)
	assert.Equal(t, "Los_Angeles", point.TagList()[0].Value)
	assert.Equal(t, "Hostname", point.TagList()[1].Key)
	assert.Equal(t, "foo", point.TagList()[1].Value)
	assert.Equal(t, "Task", point.TagList()[2].Key)
	assert.Equal(t, "curl", point.TagList()[2].Value)

	assert.Len(t, point.FieldList(), 2)
	assert.Equal(t, "PID", point.FieldList()[0].Key)
	assert.Equal(t, float64(123456), point.FieldList()[0].Value)
	assert.Equal(t, "RTT", point.FieldList()[1].Key)
	assert.Equal(t, float64(12345), point.FieldList()[1].Value)
}

func TestPointPB(t *testing.T) {
	i := &influxdb{geo: &geoMock{}, cfg: &dbConfig{GeoField: "SAddr"}}

	p := pb.Fields{}
	b := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"SAddr":"10.0.0.1","Timestamp":1611118090,"Hostname":"foo"}`)
	protojson.Unmarshal(b, &p)

	point := i.pointPB(&p)

	assert.Len(t, point.TagList(), 3)
	assert.Equal(t, "City", point.TagList()[0].Key)
	assert.Equal(t, "Los_Angeles", point.TagList()[0].Value)
	assert.Equal(t, "Hostname", point.TagList()[1].Key)
	assert.Equal(t, "foo", point.TagList()[1].Value)
	assert.Equal(t, "Task", point.TagList()[2].Key)
	assert.Equal(t, "curl", point.TagList()[2].Value)

	assert.Len(t, point.FieldList(), 2)
	assert.Equal(t, "PID", point.FieldList()[0].Key)
	assert.Equal(t, int64(123456), point.FieldList()[0].Value)
	assert.Equal(t, "RTT", point.FieldList()[1].Key)
	assert.Equal(t, int64(12345), point.FieldList()[1].Value)

}

func TestPointSPB(t *testing.T) {
	i := &influxdb{geo: &geoMock{}, cfg: &dbConfig{GeoField: "SAddr"}}

	m := map[string]interface{}{}
	b := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"SAddr":"10.0.0.1","Timestamp":1611118090,"Hostname":"foo"}`)
	json.Unmarshal(b, &m)
	spb, err := structpb.NewStruct(m)
	assert.NoError(t, err)
	point := i.pointSPB(&pb.FieldsSPB{Fields: spb})

	assert.Len(t, point.TagList(), 3)
	assert.Equal(t, "City", point.TagList()[0].Key)
	assert.Equal(t, "Los_Angeles", point.TagList()[0].Value)
	assert.Equal(t, "Hostname", point.TagList()[1].Key)
	assert.Equal(t, "foo", point.TagList()[1].Value)
	assert.Equal(t, "Task", point.TagList()[2].Key)
	assert.Equal(t, "curl", point.TagList()[2].Value)

	assert.Len(t, point.FieldList(), 2)
	assert.Equal(t, "PID", point.FieldList()[0].Key)
	assert.Equal(t, float64(123456), point.FieldList()[0].Value)
	assert.Equal(t, "RTT", point.FieldList()[1].Key)
	assert.Equal(t, float64(12345), point.FieldList()[1].Value)
}

func TestGetPointMaker(t *testing.T) {
	i := &influxdb{}
	expected := runtime.FuncForPC(reflect.ValueOf(i.pointSPB).Pointer()).Name()
	actual := runtime.FuncForPC(reflect.ValueOf(i.getPointMaker("spb")).Pointer()).Name()
	assert.Equal(t, expected, actual)

	expected = runtime.FuncForPC(reflect.ValueOf(i.pointPB).Pointer()).Name()
	actual = runtime.FuncForPC(reflect.ValueOf(i.getPointMaker("pb")).Pointer()).Name()
	assert.Equal(t, expected, actual)

	expected = runtime.FuncForPC(reflect.ValueOf(i.pointJSON).Pointer()).Name()
	actual = runtime.FuncForPC(reflect.ValueOf(i.getPointMaker("json")).Pointer()).Name()
	assert.Equal(t, expected, actual)

	assert.Nil(t, i.getPointMaker("unknown"))
}

func BenchmarkPointJSON(b *testing.B) {
	i := &influxdb{}

	m := map[string]interface{}{}
	bb := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"Timestamp":1611118090,"Hostname":"foo"}`)
	json.Unmarshal(bb, &m)

	for n := 0; n < b.N; n++ {
		i.pointJSON(m)
	}
}

func BenchmarkPointPB(b *testing.B) {
	i := &influxdb{}

	p := pb.Fields{}
	bb := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"Timestamp":1611118090,"Hostname":"foo"}`)
	protojson.Unmarshal(bb, &p)

	for n := 0; n < b.N; n++ {
		i.pointPB(&p)
	}
}

func BenchmarkPointSPB(b *testing.B) {
	i := &influxdb{}

	m := map[string]interface{}{}
	bb := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"Timestamp":1611118090,"Hostname":"foo"}`)
	json.Unmarshal(bb, &m)
	spb, _ := structpb.NewStruct(m)

	for n := 0; n < b.N; n++ {
		i.pointSPB(&pb.FieldsSPB{Fields: spb})
	}
}
