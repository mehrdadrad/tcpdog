package kafka

import (
	"log"
	"time"

	"github.com/Shopify/sarama"
	"github.com/mehrdadrad/tcpdog/config"
)

// Config represents Kafka configuration
type Config struct {
	Workers        int
	Topic          string
	Brokers        []string
	Serialization  string
	Compression    string
	RetryMax       int
	RequestSizeMax int32
	RetryBackoff   int

	SASLUsername string
	SASLPassword string

	TLSConfig config.TLSConfig
}

func kafkaConfig(cfg map[string]interface{}) *Config {
	c := &Config{
		Brokers:        []string{"localhost:9092"},
		Topic:          "tcpdog",
		Serialization:  "json",
		RequestSizeMax: 104857600,
		RetryMax:       3,
		RetryBackoff:   250, // Millisecond
		Workers:        2,
	}

	if err := config.Transform(cfg, c); err != nil {
		log.Fatal(err)
	}

	return c
}

func saramaConfig(kCfg *Config) (*sarama.Config, error) {
	sConfig := sarama.NewConfig()

	sConfig.ClientID = "tcpdog"
	sConfig.Producer.Retry.Max = kCfg.RetryMax
	sConfig.Producer.Retry.Backoff = time.Duration(kCfg.RetryBackoff) * time.Millisecond
	sarama.MaxRequestSize = kCfg.RequestSizeMax

	if kCfg.TLSConfig.Enable {
		tlsConfig, err := config.GetTLS(&kCfg.TLSConfig)
		if err != nil {
			return nil, err
		}

		sConfig.Net.TLS.Enable = true
		sConfig.Net.TLS.Config = tlsConfig
	}

	switch kCfg.Compression {
	case "gzip":
		sConfig.Producer.Compression = sarama.CompressionGZIP
	case "lz4":
		sConfig.Producer.Compression = sarama.CompressionLZ4
	case "snappy":
		sConfig.Producer.Compression = sarama.CompressionSnappy
	default:
		sConfig.Producer.Compression = sarama.CompressionNone
	}

	return sConfig, nil
}
