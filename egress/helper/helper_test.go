package helper

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mehrdadrad/tcpdog/config"
)

var cfg = config.Config{
	Egress: map[string]config.EgressConfig{
		"myoutput": {
			Type: "grpc2",
		},
	},
	Fields: map[string][]config.Field{
		"myfields": {
			{Name: "Task"},
			{Name: "Fake1"},
			{Name: "Fake2"},
		},
	},
}

func TestPBStructUnmarshal(t *testing.T) {
	spb := NewStructPB(cfg.Fields["myfields"])
	spb.hostname = "fakehost"
	buf := bytes.NewBufferString(`{"Task":"curl","Fake1":1,"Fake2":2,"Timestamp":1609720926}`)
	r := spb.Unmarshal(buf)

	assert.Equal(t, "curl", r.Fields["Task"].GetStringValue())
	assert.Equal(t, 1.0, r.Fields["Fake1"].GetNumberValue())
	assert.Equal(t, 2.0, r.Fields["Fake2"].GetNumberValue())
	assert.Equal(t, "fakehost", r.Fields["Hostname"].GetStringValue())

}

func TestBackoff(t *testing.T) {
	cfg := config.Config{}
	cfg.SetMockLogger("memory")

	b := NewBackoff(cfg.Logger())
	// first time, zero delay
	now := time.Now()
	b.Next()
	assert.Less(t, time.Since(now).Milliseconds(), int64(100))
	now = time.Now()
	b.Next()
	assert.Less(t, time.Since(now).Milliseconds(), int64(2500))
	assert.Greater(t, time.Since(now).Milliseconds(), int64(2000))

	// resert after 30 mins
	b.last = time.Unix(1612113039, 0)
	now = time.Now()
	b.Next()
	assert.Less(t, time.Since(now).Milliseconds(), int64(100))
}

func BenchmarkPBStructUnmarshal(b *testing.B) {
	spb := NewStructPB(cfg.Fields["myfields"])

	for i := 0; i < b.N; i++ {
		buf := bytes.NewBufferString(`{"Task":"curl","Fake1":1,"Fake2":2,"Timestamp":1609720926}`)
		spb.Unmarshal(buf)
	}
}
