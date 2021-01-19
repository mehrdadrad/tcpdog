package grpc

import (
	"log"

	"github.com/mehrdadrad/tcpdog/server/config"
)

// Config represents kafka consumer configuration
type Config struct {
	Addr             string
	NumStreamWorkers uint32
	TLSConfig        *config.TLSConfig
}

func grpcConfig(cfg map[string]interface{}) *Config {
	// default configuration
	conf := &Config{
		Addr: ":8085",
	}

	if err := config.Transform(cfg, conf); err != nil {
		log.Fatal(err)
	}

	return conf
}
