package helper

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
	"go.uber.org/zap"

	"github.com/mehrdadrad/tcpdog/config"
)

var comma = []byte(",")[0]

// Backoff represents backoff strategy
type Backoff struct {
	duration time.Duration
	last     time.Time
	cfg      *config.Config
}

// StructPB represents the conversion between
// json bytes to StructPB.
type StructPB struct {
	fieldsLen  []int
	fieldsName []string
	isString   map[string]bool
	hostname   string
	buffer     *bytes.Buffer
}

// NewStructPB constructs and initializes a struct pb.
func NewStructPB(fields []config.Field) *StructPB {
	s := &StructPB{}
	s.init(fields)
	return s
}

func (s *StructPB) init(fields []config.Field) {
	var err error
	s.isString = map[string]bool{
		"Task":  true,
		"SAddr": true,
		"DAddr": true,
	}

	s.hostname, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range fields {
		s.fieldsLen = append(s.fieldsLen, len(f.Name)+3)
		s.fieldsName = append(s.fieldsName, f.Name)
	}
}

// Unmarshal decode bytes to protobuf struct
func (s *StructPB) Unmarshal(buf *bytes.Buffer) *pbstruct.Struct {
	r := &pbstruct.Struct{Fields: make(map[string]*pbstruct.Value)}

	buf.Next(1) // skip bracket
	for i, l := range s.fieldsLen {
		buf.Next(l)
		name := s.fieldsName[i]

		if s.isString[name] {
			v, err := buf.ReadBytes(comma)
			if err != nil {
				log.Fatal(err)
			}
			r.Fields[name] = &pbstruct.Value{
				Kind: &pbstruct.Value_StringValue{StringValue: string(v[1 : len(v)-2])},
			}
		} else {
			v, err := buf.ReadBytes(comma)
			if err != nil {
				log.Fatal(err)
			}
			vi, err := strconv.Atoi(string(v[:len(v)-1]))
			if err != nil {
				log.Fatal(err)
			}
			r.Fields[name] = &pbstruct.Value{
				Kind: &pbstruct.Value_NumberValue{NumberValue: float64(vi)},
			}
		}
	}

	// timestamp
	buf.Next(12)
	vv, err := strconv.Atoi(string(buf.Next(10)))
	if err != nil {
		log.Println(err)
	}
	r.Fields["Hostname"] = &pbstruct.Value{
		Kind: &pbstruct.Value_StringValue{StringValue: s.hostname},
	}
	r.Fields["Timestamp"] = &pbstruct.Value{
		Kind: &pbstruct.Value_NumberValue{NumberValue: float64(vv)},
	}

	return r
}

// NewBackoff constructs a new backoff
func NewBackoff(cfg *config.Config) *Backoff {
	return &Backoff{cfg: cfg}
}

// Next waits for a specific backoff time
func (b *Backoff) Next() {
	if b.duration == 0 {
		b.reset()
		return
	}

	if time.Since(b.last).Minutes() > 30 {
		b.reset()
		return
	}

	if b.duration.Minutes() < 2 {
		b.duration += b.duration * 15 / 100
		b.last = time.Now()
	}

	b.cfg.Logger().Info("backoff", zap.String("delay", fmt.Sprintf("%.2fs", b.duration.Seconds())))
	time.Sleep(b.duration)
}

func (b *Backoff) reset() {
	b.last = time.Now()
	b.duration = time.Duration(2 * time.Second)
}
