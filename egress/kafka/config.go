package kafka

import (
	"log"
	"time"

	"github.com/Shopify/sarama"
	"github.com/mehrdadrad/tcpdog/config"
)

// Config represents Kafka configuration
type Config struct {
	Workers        int      `yaml:"workers"`
	Topic          string   `yaml:"topic"`
	Brokers        []string `yaml:"brokers"`
	Serialization  string   `yaml:"serialization"`
	Compression    string   `yaml:"compression" `
	RetryMax       int      `yaml:"retry-max"`
	RequestSizeMax int32    `yaml:"request-size-max"`
	RetryBackoff   int      `yaml:"retry-backoff"`
	TLSEnabled     bool     `yaml:"tls-enabled"`
	TLSCertFile    string   `yaml:"tls-cert"`
	TLSKeyFile     string   `yaml:"tls-key"`
	CAFile         string   `yaml:"ca-file"`
	TLSSkipVerify  bool     `yaml:"tls-skip-verify"`
	SASLUsername   string   `yaml:"sasl-username"`
	SASLPassword   string   `yaml:"sasl-password"`
}

func kafkaConfig(cfg map[string]interface{}) *Config {
	c := &Config{
		Brokers:        []string{"localhost:9092"},
		Topic:          "tcpdog",
		Serialization:  "json",
		RequestSizeMax: 104857600,
		RetryMax:       2,
		RetryBackoff:   10,
		TLSSkipVerify:  true,
		Workers:        2,
	}

	if err := config.Transform(cfg, c); err != nil {
		log.Fatal(err)
	}

	return c
}

func saramaConfig(c *Config) *sarama.Config {
	config := sarama.NewConfig()

	config.ClientID = "tcpdog"
	config.Producer.Retry.Max = c.RetryMax
	config.Producer.Retry.Backoff = time.Duration(c.RetryBackoff) * time.Millisecond
	sarama.MaxRequestSize = c.RequestSizeMax

	switch c.Compression {
	case "gzip":
		config.Producer.Compression = sarama.CompressionGZIP
	case "lz4":
		config.Producer.Compression = sarama.CompressionLZ4
	case "snappy":
		config.Producer.Compression = sarama.CompressionSnappy
	default:
		config.Producer.Compression = sarama.CompressionNone
	}

	return config
}
