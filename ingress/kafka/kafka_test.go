package kafka

import (
	"context"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/mehrdadrad/tcpdog/config"
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

func TestStart(t *testing.T) {
	broker := sarama.NewMockBroker(t, 0)
	defer broker.Close()

	// resp := new(sarama.MetadataResponse)
	// broker.Returns(resp)
	mockFetchResponse := sarama.NewMockFetchResponse(t, 1)
	for i := 0; i < 10; i++ {
		mockFetchResponse.SetMessage("tcpdog", 0, int64(i+1234), sarama.StringEncoder(`{"v1":"k1"}`))
	}

	broker.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(broker.Addr(), broker.BrokerID()).
			SetLeader("tcpdog", 0, broker.BrokerID()),
		"OffsetRequest": sarama.NewMockOffsetResponse(t).
			SetOffset("tcpdog", 0, sarama.OffsetOldest, 0).
			SetOffset("tcpdog", 0, sarama.OffsetNewest, 2345),
		"FetchRequest": mockFetchResponse,
	})

	cfg := config.ServerConfig{
		Ingress: map[string]config.Ingress{
			"foo": {
				Config: map[string]interface{}{
					"brokers": []string{broker.Addr()},
					"topic":   "tcpdog",
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = cfg.WithContext(ctx)
	ch := make(chan interface{}, 1)
	err := Start(ctx, "foo", "json", ch)
	assert.NoError(t, err)
}

func TestSaramaConfig(t *testing.T) {
	conf := &Config{
		Brokers:      []string{"localhost:9092"},
		Topic:        "tcpdog",
		RetryBackoff: 2,
		Workers:      2,
		Version:      "0.10.2.1",
		TLSConfig: config.TLSConfig{
			Enable:             true,
			InsecureSkipVerify: true,
		},
	}

	_, err := saramaConfig(conf)
	assert.NoError(t, err)
}
