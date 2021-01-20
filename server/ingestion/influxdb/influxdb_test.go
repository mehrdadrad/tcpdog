package influxdb

import (
	"encoding/json"
	"testing"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestPointJSON(t *testing.T) {
	i := &influxdb{}

	m := map[string]interface{}{}
	b := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"Timestamp":1611118090,"Hostname":"foo"}`)
	json.Unmarshal(b, &m)

	p := i.pointJSON(m)

	assert.Len(t, p.TagList(), 2)
	assert.Equal(t, "Hostname", p.TagList()[0].Key)
	assert.Equal(t, "foo", p.TagList()[0].Value)
	assert.Equal(t, "Task", p.TagList()[1].Key)
	assert.Equal(t, "curl", p.TagList()[1].Value)

	assert.Len(t, p.FieldList(), 2)
	assert.Equal(t, "PID", p.FieldList()[0].Key)
	assert.Equal(t, float64(123456), p.FieldList()[0].Value)
	assert.Equal(t, "RTT", p.FieldList()[1].Key)
	assert.Equal(t, float64(12345), p.FieldList()[1].Value)
}

func TestPointPB(t *testing.T) {
	p := pb.Fields{}
	b := []byte(`{"PID":123456,"Task":"curl","RTT":12345,"Timestamp":1611118090,"Hostname":"foo"}`)
	protojson.Unmarshal(b, &p)

	assert.Equal(t, "foo", *p.Hostname)
	assert.Equal(t, int32(123456), *p.PID)
	assert.Equal(t, "curl", *p.Task)
	assert.Equal(t, int32(12345), *p.RTT)
	assert.Equal(t, int64(1611118090), *p.Timestamp)
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
