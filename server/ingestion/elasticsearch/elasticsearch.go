package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/mehrdadrad/tcpdog/proto"
	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/geo"
)

type elastic struct {
	geo           geo.Geoer
	cfg           *esConfig
	serialization string
}

// Start starts ingestion data points to influxdb
func Start(ctx context.Context, name string, ser string, ch chan interface{}) {
	var g geo.Geoer

	cfg := config.FromContext(ctx)
	eCfg := elasticSearchConfig(cfg.Ingestion[name].Config)
	logger := cfg.Logger()

	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		logger.Fatal("elasticsearch", zap.Error(err))
	}

	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: client,
		Index:  "tcpdog",
		OnError: func(ctx context.Context, err error) {
			logger.Error("elasticsearch", zap.Error(err))
		},
	})

	// if geo is available
	if v, ok := geo.Reg[cfg.Geo.Type]; ok {
		g = v
		g.Init(cfg.Logger(), cfg.Geo.Config)
	}

	e := elastic{geo: g, cfg: eCfg, serialization: ser}

	iCh := make(chan *esutil.BulkIndexerItem, 1000)
	for c := 0; c < eCfg.Workers; c++ {
		go e.iWorker(ctx, ch, iCh)
	}

	go func() {
		for {
			select {
			case item := <-iCh:
				err = indexer.Add(ctx, *item)
				if err != nil {
					logger.Error("elasticsearch", zap.Error(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

}

func (e *elastic) iWorker(ctx context.Context, ch chan interface{}, iCh chan *esutil.BulkIndexerItem) {
	var fields interface{}

	getItem := e.getItemMaker(e.serialization)

	for {
		select {
		case fields = <-ch:
			item := getItem(fields)
			if item != nil {
				iCh <- item
			}
		case <-ctx.Done():
			return
		}
	}
}

func (e *elastic) getItemMaker(ser string) func(fi interface{}) *esutil.BulkIndexerItem {
	switch ser {
	case "json":
		return e.itemJSON
	case "spb":
		return e.itemSPB
	case "pb":
		return e.itemPB
	}

	return nil
}

func (e *elastic) itemJSON(fi interface{}) *esutil.BulkIndexerItem {
	f := fi.(map[string]interface{})

	if e.geo != nil {
		if gv, ok := f[e.cfg.GeoField]; ok {
			for k, v := range e.geo.Get(gv.(string)) {
				f[k] = v
			}
		}
	}

	b, err := json.Marshal(f)
	if err != nil {
		return nil
	}

	return &esutil.BulkIndexerItem{
		Action: "index",
		Body:   bytes.NewReader(b),
	}
}

func (e *elastic) itemSPB(fi interface{}) *esutil.BulkIndexerItem {
	f := fi.(*pb.FieldsPBS)

	if e.geo != nil {
		if gv, ok := f.Fields.Fields[e.cfg.GeoField]; ok {
			for k, v := range e.geo.Get(gv.GetStringValue()) {
				f.Fields.Fields[k] = structpb.NewStringValue(v)
			}
		}
	}

	b, err := protojson.Marshal(f.Fields)
	if err != nil {
		return nil
	}

	return &esutil.BulkIndexerItem{
		Action: "index",
		Body:   bytes.NewReader(b),
	}
}

func (e *elastic) itemPB(fi interface{}) *esutil.BulkIndexerItem {
	var geoKV map[string]string

	f := fi.(*pb.Fields)

	value := reflect.ValueOf(f).Elem()

	if e.geo != nil {
		gv := value.FieldByName(e.cfg.GeoField)
		if gv.IsValid() {
			geoKV = e.geo.Get(gv.Elem().String())
		}
	}

	for k, v := range geoKV {
		v := v
		fv := value.FieldByName(k)
		if fv.IsValid() {
			fv.Set(reflect.ValueOf(&v))
		}
	}

	b, err := protojson.Marshal(f)
	if err != nil {
		return nil
	}

	return &esutil.BulkIndexerItem{
		Action: "index",
		Body:   bytes.NewReader(b),
	}
}
