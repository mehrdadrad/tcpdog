package kafka

import (
	"log"

	"github.com/mehrdadrad/tcpdog/server/config"
)

// Config represents kafka consumer configuration
type Config struct {
	Brokers []string
	Topic   string
	Workers int
}

func influxConfig(cfg map[string]interface{}) *Config {
	// default configuration
	conf := &Config{
		Brokers: []string{"localhost:9092"},
		Topic:   "tcpdog",
		Workers: 2,
	}

	if err := config.Transform(cfg, conf); err != nil {
		log.Fatal(err)
	}

	return conf
}
