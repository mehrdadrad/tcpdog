package influxdb

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/geo"
)

const maxChanSize = 1000

type influxdb struct {
	g   geo.Geoer
	cfg *dbConfig
}

// Start starts ingestion data points to influxdb
func Start(ctx context.Context, ch chan *pb.FieldsPBS) {
	var g geo.Geoer

	cfg := config.FromContext(ctx)
	dCfg := influxConfig(cfg.Ingest.Config)

	client := influxdb2.NewClientWithOptions(dCfg.URL, "", influxdbOpts(dCfg))
	writeAPI := client.WriteAPI(dCfg.Org, dCfg.Bucket)

	// if geo is available
	if v, ok := geo.Reg[cfg.Geo.Type]; ok {
		g = v
		g.Init(cfg.Logger(), cfg.Geo.Config)
	}

	i := influxdb{g: g, cfg: dCfg}

	pCh := make(chan *write.Point, maxChanSize)

	for c := uint(0); c < dCfg.Workers; c++ {
		go i.pWorker(ctx, ch, pCh)
	}

	// main influxdb loop
	go func() {
		for {
			select {
			case p := <-pCh:
				writeAPI.WritePoint(p)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// pWorker creates influxdb points
func (i *influxdb) pWorker(ctx context.Context, ch chan *pb.FieldsPBS, pCh chan *write.Point) {
	fields := &pb.FieldsPBS{}
	for {
		select {
		case fields = <-ch:
			pCh <- i.point(fields)
		case <-ctx.Done():
			return
		}
	}
}

// point returns influxdb point with geo (if available)
func (i *influxdb) point(f *pb.FieldsPBS) *write.Point {
	tags := map[string]string{}
	fields := map[string]interface{}{}

	for key, field := range f.Fields.Fields {
		if value, ok := field.GetKind().(*structpb.Value_StringValue); ok {
			if i.g != nil && (key == i.cfg.GeoField) {
				for k1, v1 := range i.g.Get(value.StringValue) {
					tags[k1] = v1
				}
				continue
			}
			tags[key] = value.StringValue
		} else {
			fields[key] = field.GetNumberValue()
		}
	}

	return influxdb2.NewPoint("tcpdog", tags, fields, time.Now())
}

// influxdbOpts returns influxdb options
func influxdbOpts(cfg *dbConfig) *influxdb2.Options {
	opts := influxdb2.DefaultOptions()
	opts.SetMaxRetries(cfg.MaxRetries)
	opts.SetHTTPRequestTimeout(cfg.Timeout)
	opts.SetBatchSize(cfg.BatchSize)

	return opts
}
