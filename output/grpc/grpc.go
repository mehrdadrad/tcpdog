package grpc

import (
	"bytes"
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
	pb "github.com/mehrdadrad/tcpdog/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/mehrdadrad/tcpdog/config"
)

var comma = []byte(",")[0]

type grpc2 struct {
	fieldsLen  []int
	fieldsName []string
	isString   map[string]bool
	hostname   string
	buffer     *bytes.Buffer
}

func (g *grpc2) init(fields []config.Field) {
	var err error
	g.isString = map[string]bool{
		"Task":  true,
		"SAddr": true,
		"DAddr": true,
	}

	g.hostname, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range fields {
		g.fieldsLen = append(g.fieldsLen, len(f.Name)+3)
		g.fieldsName = append(g.fieldsName, f.Name)
	}
}

func (g *grpc2) pbStructUnmarshal(buf *bytes.Buffer) *pbstruct.Struct {
	r := &pbstruct.Struct{Fields: make(map[string]*pbstruct.Value)}

	buf.Next(1) // skip bracket
	for i, l := range g.fieldsLen {
		buf.Next(l)
		name := g.fieldsName[i]

		if g.isString[name] {
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
		Kind: &pbstruct.Value_StringValue{StringValue: g.hostname},
	}
	r.Fields["Timestamp"] = &pbstruct.Value{
		Kind: &pbstruct.Value_NumberValue{NumberValue: float64(vv)},
	}

	return r
}

// Start2 ...
func Start2(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	g := &grpc2{buffer: new(bytes.Buffer)}
	cfg := config.FromContext(ctx)
	g.init(cfg.Fields[tp.Fields])

	conn, err := grpc.Dial(":8085", grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := pb.NewTCPDogClient(conn)
	stream, err := client.TracepointPBS(context.Background())
	if err != nil {
		return err
	}

	go func() {
		var buf *bytes.Buffer
		for {
			buf = <-ch
			t := time.Now()
			r := g.pbStructUnmarshal(buf)
			log.Println("ELAPSED:", time.Since(t).Nanoseconds())
			err = stream.Send(&pb.FieldsPBS{
				Fields: r,
			})
			if err != nil {
				log.Println(err)
			}
			bufpool.Put(buf)
		}
	}()

	return nil
}

// Start sends fields to a grpc server
func Start(ctx context.Context, cfg map[string]string, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	conn, err := grpc.Dial(":8085", grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := pb.NewTCPDogClient(conn)
	stream, err := client.Tracepoint(context.Background())
	if err != nil {
		return err
	}

	go func() {
		var data *bytes.Buffer

		for {
			select {
			case data = <-ch:
				m := pb.Fields{}
				t := time.Now()
				protojson.Unmarshal(data.Bytes(), &m)
				log.Println("JSON ELAPSED:", time.Since(t).Nanoseconds())
				stream.Send(&m)
				bufpool.Put(data)
			case <-ctx.Done():
				stream.CloseAndRecv()
			}
		}
	}()

	return nil
}
