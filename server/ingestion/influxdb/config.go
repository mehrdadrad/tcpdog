package influxdb

import (
	"log"

	"github.com/mehrdadrad/tcpdog/config"
)

type dbConfig struct {
	URL        string `yaml:"url"`
	Org        string
	Bucket     string
	Timeout    uint
	MaxRetries uint
	BatchSize  uint
	Workers    uint

	GeoField string `yaml:"geoField"`
}

func influxDBConfig(cfg map[string]interface{}) *dbConfig {
	// default configuration
	conf := &dbConfig{
		URL:        "http://localhost:8086",
		Bucket:     "tcpdog",
		Timeout:    5,
		MaxRetries: 10,
		BatchSize:  200,
		Workers:    2,
		GeoField:   "DAddr",
	}

	if err := config.Transform(cfg, conf); err != nil {
		log.Fatal(err)
	}

	return conf
}
