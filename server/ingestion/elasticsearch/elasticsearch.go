package elasticsearch

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"

	"github.com/mehrdadrad/tcpdog/server/config"
	"github.com/mehrdadrad/tcpdog/server/geo"
)

type elastic struct {
	g             geo.Geoer
	cfg           *esConfig
	serialization string
}

// Start starts ingestion data points to influxdb
func Start(ctx context.Context, name string, ser string, ch chan interface{}) {
	var g geo.Geoer

	cfg := config.FromContext(ctx)
	eCfg := elasticSearchConfig(cfg.Ingestion[name].Config)

	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: client,
		Index:  "tcpdog",
	})

	// if geo is available
	if v, ok := geo.Reg[cfg.Geo.Type]; ok {
		g = v
		g.Init(cfg.Logger(), cfg.Geo.Config)
	}

	e := elastic{g: g, cfg: eCfg, serialization: ser}

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
					panic(err)
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
	}

	return nil
}

func (e *elastic) itemJSON(fi interface{}) *esutil.BulkIndexerItem {
	f := fi.(map[string]interface{})

	for key, field := range f {
		if value, ok := field.(string); ok {
			if e.g != nil && (key == e.cfg.GeoField) {
				for k1, v1 := range e.g.Get(value) {
					f[k1] = v1
				}
			}
		}
	}

	b, err := json.Marshal(f)
	if err != nil {
		return nil
	}

	return &esutil.BulkIndexerItem{
		Action: "index",
		Body:   strings.NewReader(string(b)),
	}
}
