package elasticsearch

import (
	"log"

	"github.com/mehrdadrad/tcpdog/config"
)

type esConfig struct {
	Addrs []string
	Index string

	Workers int

	GeoField string `yaml:"geoField"`
}

func elasticSearchConfig(cfg map[string]interface{}) *esConfig {
	// default configuration
	es := &esConfig{
		Addrs:   []string{"localhost:9200"},
		Index:   "tcpdog",
		Workers: 2,
	}

	if err := config.Transform(cfg, es); err != nil {
		log.Fatal(err)
	}

	return es
}
