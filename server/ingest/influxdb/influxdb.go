package influxdb

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/geo"
	"google.golang.org/protobuf/types/known/structpb"
)

var geoField string

// Start starts ingestion data points to influxdb
func Start(ctx context.Context, ch chan *pb.FieldsPBS) {
	var g geo.Geoer

	cfg := influxConfig(config.FromContext(ctx).Ingest.Config)
	gCfg := config.FromContext(ctx).Geo

	client := influxdb2.NewClientWithOptions(cfg.URL, "", influxdbOpts(cfg))
	writeAPI := client.WriteAPI(cfg.Org, cfg.Bucket)

	if v, ok := geo.Reg[gCfg.Type]; ok {
		g = v
		g.Init(gCfg.Config)
		geoField = gField(gCfg)
	}

	pCh := make(chan *write.Point, 1000)

	for i := uint(0); i < cfg.Workers; i++ {
		go pWorker(ctx, g, ch, pCh)
	}

	for {
		select {
		case p := <-pCh:
			writeAPI.WritePoint(p)
		case <-ctx.Done():
			return
		}
	}
}

func pWorker(ctx context.Context, geo geo.Geoer, ch chan *pb.FieldsPBS, pCh chan *write.Point) {
	fields := &pb.FieldsPBS{}
	for {
		select {
		case fields = <-ch:
			pCh <- point(fields, geo)
		case <-ctx.Done():
			return
		}
	}
}

func point(f *pb.FieldsPBS, g geo.Geoer) *write.Point {
	tags := map[string]string{}
	fields := map[string]interface{}{}

	for key, field := range f.Fields.Fields {
		if value, ok := field.GetKind().(*structpb.Value_StringValue); ok {
			if g != nil && (key == "DAddr") {
				for k1, v1 := range g.Get(value.StringValue) {
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

func gField(cfg config.Geo) string {
	if v, ok := cfg.Config["field"]; ok {
		return v
	}

	return "DAddr"
}

func influxdbOpts(cfg *dbConfig) *influxdb2.Options {
	opts := influxdb2.DefaultOptions()
	opts.SetMaxRetries(cfg.MaxRetries)
	opts.SetHTTPRequestTimeout(cfg.Timeout)
	opts.SetBatchSize(cfg.BatchSize)

	return opts
}
