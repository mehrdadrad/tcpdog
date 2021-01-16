package influxdb

import (
	"context"
	"reflect"
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
	g             geo.Geoer
	cfg           *dbConfig
	serialization string
}

// Start starts ingestion data points to influxdb
func Start(ctx context.Context, name string, ser string, ch chan interface{}) {
	var g geo.Geoer

	cfg := config.FromContext(ctx)
	dCfg := influxConfig(cfg.Ingestion[name].Config)

	client := influxdb2.NewClientWithOptions(dCfg.URL, "", influxdbOpts(dCfg))
	writeAPI := client.WriteAPI(dCfg.Org, dCfg.Bucket)

	// if geo is available
	if v, ok := geo.Reg[cfg.Geo.Type]; ok {
		g = v
		g.Init(cfg.Logger(), cfg.Geo.Config)
	}

	i := influxdb{g: g, cfg: dCfg, serialization: ser}

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

// pWorker creates influxdb points from struct-pb
func (i *influxdb) pWorker(ctx context.Context, ch chan interface{}, pCh chan *write.Point) {
	var fields interface{}

	point := i.getPointMaker(i.serialization)

	for {
		select {
		case fields = <-ch:
			p := point(fields)
			if p != nil {
				pCh <- p
			}
			//_ = p
		case <-ctx.Done():
			return
		}
	}
}

func (i *influxdb) getPointMaker(ser string) func(fi interface{}) *write.Point {
	switch ser {
	case "json":
		return i.pointJSON
	case "spb":
		return i.pointPBS
	case "pb":
		return i.pointPB
	}

	return nil
}

// pointPBS returns influxdb pointPBS with geo (if available)
func (i *influxdb) pointPBS(fi interface{}) *write.Point {
	var (
		tags      = map[string]string{}
		fields    = map[string]interface{}{}
		timestamp time.Time
	)

	f := fi.(*pb.FieldsPBS)

	for key, field := range f.Fields.Fields {
		if value, ok := field.GetKind().(*structpb.Value_StringValue); ok {
			if i.g != nil && (key == i.cfg.GeoField) {
				for k1, v1 := range i.g.Get(value.StringValue) {
					tags[k1] = v1
				}
				continue
			}
			tags[key] = value.StringValue
		} else if key != "Timestamp" {
			fields[key] = field.GetNumberValue()
		} else {
			timestamp = time.Unix(int64(field.GetNumberValue()), 0)
		}
	}

	return influxdb2.NewPoint("tcpdog", tags, fields, timestamp)
}

// point returns influxdb point with geo (if available)
func (i *influxdb) pointPB(fi interface{}) *write.Point {
	var (
		tags      = map[string]string{}
		fields    = map[string]interface{}{}
		timestamp time.Time
	)

	f := fi.(*pb.Fields)

	v := reflect.ValueOf(f).Elem()

	for n := 0; n < v.NumField(); n++ {
		switch v.Field(n).Type().Kind() {
		case reflect.Ptr:
			if v.Field(n).Pointer() != 0 {
				switch v.Field(n).Addr().Elem().Elem().Kind() {
				case reflect.String:
					if i.g != nil && (v.Type().Field(n).Name == i.cfg.GeoField) {
						for k1, v1 := range i.g.Get(v.Field(n).Elem().String()) {
							tags[k1] = v1
						}
						continue
					}
					tags[v.Type().Field(n).Name] = v.Field(n).Elem().String()
				case reflect.Int32:
					fields[v.Type().Field(n).Name] = v.Field(n).Elem().Int()
				case reflect.Int64:
					if v.Type().Field(n).Name != "Timestamp" {
						fields[v.Type().Field(n).Name] = v.Field(n).Elem().Int()
					} else {
						timestamp = time.Unix(v.Field(n).Elem().Int(), 0)
					}
				}
			}
		}
	}

	return influxdb2.NewPoint("tcpdog", tags, fields, timestamp)
}

func (i *influxdb) pointJSON(fi interface{}) *write.Point {
	var (
		tags      = map[string]string{}
		fields    = map[string]interface{}{}
		timestamp time.Time
	)

	f := fi.(map[string]interface{})

	for key, field := range f {
		if value, ok := field.(string); ok {
			if i.g != nil && (key == i.cfg.GeoField) {
				for k1, v1 := range i.g.Get(value) {
					tags[k1] = v1
				}
				continue
			}
			tags[key] = value
		} else if key != "Timestamp" {
			fields[key] = field.(float64)
		} else {
			timestamp = time.Unix(int64(field.(float64)), 0)
		}
	}

	return influxdb2.NewPoint("tcpdog", tags, fields, timestamp)
}

// influxdbOpts returns influxdb options
func influxdbOpts(cfg *dbConfig) *influxdb2.Options {
	opts := influxdb2.DefaultOptions()
	opts.SetMaxRetries(cfg.MaxRetries)
	opts.SetHTTPRequestTimeout(cfg.Timeout)
	opts.SetBatchSize(cfg.BatchSize)

	return opts
}
