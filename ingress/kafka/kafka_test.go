package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/mehrdadrad/tcpdog/proto"
)

func TestGetUnmarshalJSON(t *testing.T) {
	f := getUnmarshal("json")
	b := []byte(`{"F1":5,"Timestamp":1611634115,"Hostname":"foo"}`)
	v, err := f(b)
	assert.NoError(t, err)

	m := v.(map[string]interface{})
	assert.Equal(t, float64(5), m["F1"])
	assert.Equal(t, float64(1611634115), m["Timestamp"])
	assert.Equal(t, "foo", m["Hostname"])
}

func TestGetUnmarshalPB(t *testing.T) {
	f := getUnmarshal("pb")
	r := uint32(5)
	s := "foo"
	p := pb.Fields{RTT: &r, Hostname: &s}
	b, err := proto.Marshal(&p)
	assert.NoError(t, err)

	v, err := f(b)
	assert.NoError(t, err)

	m := v.(*pb.Fields)
	assert.Equal(t, uint32(5), *m.RTT)
	assert.Equal(t, "foo", *m.Hostname)
}

func TestGetUnmarshalSPB(t *testing.T) {
	f := getUnmarshal("spb")
	b := []byte(`{"F1":5,"Timestamp":1611634115,"Hostname":"foo"}`)
	p := structpb.Struct{}
	protojson.Unmarshal(b, &p)
	b, err := proto.Marshal(&pb.FieldsSPB{Fields: &p})
	assert.NoError(t, err)

	v, err := f(b)
	assert.NoError(t, err)

	m := v.(*pb.FieldsSPB)

	assert.Equal(t, float64(5), m.Fields.Fields["F1"].GetNumberValue())
	assert.Equal(t, float64(1611634115), m.Fields.Fields["Timestamp"].GetNumberValue())
	assert.Equal(t, "foo", m.Fields.Fields["Hostname"].GetStringValue())

	// cover nil
	f = getUnmarshal("unknown")
	assert.Nil(t, f)
}
