package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/mehrdadrad/tcpdog/geo"
	pb "github.com/mehrdadrad/tcpdog/proto"
)

type elastic struct {
	geo           geo.Geoer
	cfg           *esConfig
	serialization string
}

// Start starts ingestion data points to influxdb
func Start(ctx context.Context, name string, ser string, ch chan interface{}) error {
	var g geo.Geoer

	cfg := config.FromContextServer(ctx)
	logger := cfg.Logger()

	eCfg, err := elasticSearchConfig(cfg.Ingestion[name].Config)
	if err != nil {
		return err
	}

	client, err := elasticsearch.NewClient(eCfg.clientConfig)
	if err != nil {
		return err
	}

	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:        client,
		Index:         eCfg.Index,
		FlushBytes:    eCfg.FlushBytes,
		FlushInterval: time.Duration(eCfg.FlushInterval) * time.Second,

		OnError: func(ctx context.Context, err error) {
			if err != context.Canceled {
				logger.Error("es.bulkindexer", zap.Error(err))
			}
		},
	})
	if err != nil {
		return err
	}

	// if geo is available
	if v, ok := geo.Reg[cfg.Geo.Type]; ok {
		g = v
		g.Init(cfg.Logger(), cfg.Geo.Config)
	}

	e := elastic{geo: g, cfg: eCfg, serialization: ser}

	iCh := make(chan *esutil.BulkIndexerItem, 1000)

	// marshaler workers (encode data)
	for c := 0; c < eCfg.Workers; c++ {
		go e.iWorker(ctx, ch, iCh)
	}

	go func() {
		for {
			select {
			case item := <-iCh:
				err = indexer.Add(ctx, *item)
				if err != nil {
					logger.Error("es.add", zap.Error(err))
				}
			case <-ctx.Done():
				indexer.Close(ctx)
				return
			}
		}
	}()

	return nil
}

// iWorker creates elasticsearch item
func (e *elastic) iWorker(ctx context.Context, ch chan interface{}, iCh chan *esutil.BulkIndexerItem) {
	var fields interface{}

	logger := config.FromContextServer(ctx).Logger()
	getItem := e.getItemMaker(e.serialization)

	for {
		select {
		case fields = <-ch:
			item, err := getItem(fields)
			if err != nil {
				logger.Error("es.worker", zap.Error(err))
				continue
			}

			iCh <- item
		case <-ctx.Done():
			return
		}
	}
}

func (e *elastic) getItemMaker(ser string) func(fi interface{}) (*esutil.BulkIndexerItem, error) {
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

func (e *elastic) itemJSON(fi interface{}) (*esutil.BulkIndexerItem, error) {
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
		return nil, err
	}

	return &esutil.BulkIndexerItem{
		Action: "index",
		Body:   bytes.NewReader(b),
	}, nil
}

func (e *elastic) itemSPB(fi interface{}) (*esutil.BulkIndexerItem, error) {
	f := fi.(*pb.FieldsSPB)

	if e.geo != nil {
		if gv, ok := f.Fields.Fields[e.cfg.GeoField]; ok {
			for k, v := range e.geo.Get(gv.GetStringValue()) {
				f.Fields.Fields[k] = structpb.NewStringValue(v)
			}
		}
	}

	b, err := protojson.Marshal(f.Fields)
	if err != nil {
		return nil, err
	}

	return &esutil.BulkIndexerItem{
		Action: "index",
		Body:   bytes.NewReader(b),
	}, nil
}

func (e *elastic) itemPB(fi interface{}) (*esutil.BulkIndexerItem, error) {
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
		return nil, err
	}

	return &esutil.BulkIndexerItem{
		Action: "index",
		Body:   bytes.NewReader(b),
	}, nil
}
