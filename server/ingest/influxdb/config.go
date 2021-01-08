package influxdb

import (
	"encoding/json"
	"log"
)

type dbConfig struct {
	URL        string `json:"url"`
	Org        string
	Bucket     string
	Timeout    uint
	MaxRetries uint
	BatchSize  uint
	Workers    uint
}

func influxConfig(cfg map[string]interface{}) *dbConfig {
	b, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// default configuration
	conf := &dbConfig{
		URL:        "http://localhost:8086",
		Bucket:     "tcpdog",
		Timeout:    5,
		MaxRetries: 10,
		BatchSize:  200,
		Workers:    2,
	}

	err = json.Unmarshal(b, conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}
