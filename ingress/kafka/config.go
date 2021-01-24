package kafka

import (
	"log"
	"time"

	"github.com/Shopify/sarama"

	"github.com/mehrdadrad/tcpdog/config"
)

var kafkaVersion = map[string]sarama.KafkaVersion{
	"0.8.2.0":  sarama.V0_8_2_0,
	"0.8.2.1":  sarama.V0_8_2_1,
	"0.8.2.2":  sarama.V0_8_2_2,
	"0.9.0.0":  sarama.V0_9_0_0,
	"0.9.0.1":  sarama.V0_9_0_1,
	"0.10.0.0": sarama.V0_10_0_0,
	"0.10.0.1": sarama.V0_10_0_1,
	"0.10.1.0": sarama.V0_10_1_0,
	"0.10.1.1": sarama.V0_10_1_1,
	"0.10.2.0": sarama.V0_10_2_0,
	"0.10.2.1": sarama.V0_10_2_1,
	"0.11.0.0": sarama.V0_11_0_0,
	"0.11.0.1": sarama.V0_11_0_1,
	"0.11.0.2": sarama.V0_11_0_2,
	"1.0.0.0":  sarama.V1_0_0_0,
	"1.1.0.0":  sarama.V1_1_0_0,
	"1.1.1.0":  sarama.V1_1_1_0,
	"2.0.0.0":  sarama.V2_0_0_0,
	"2.0.1.0":  sarama.V2_0_1_0,
	"2.1.0.0":  sarama.V2_1_0_0,
	"2.2.0.0":  sarama.V2_2_0_0,
	"2.3.0.0":  sarama.V2_3_0_0,
	"2.4.0.0":  sarama.V2_4_0_0,
	"2.5.0.0":  sarama.V2_5_0_0,
}

// Config represents kafka consumer configuration
type Config struct {
	Brokers      []string
	Topic        string
	RetryBackoff int // seconds
	Workers      int
	Version      string

	TLSConfig config.TLSConfig
}

func kafkaConfig(cfg map[string]interface{}) *Config {
	// default configuration
	conf := &Config{
		Brokers:      []string{"localhost:9092"},
		Topic:        "tcpdog",
		RetryBackoff: 2,
		Workers:      2,
		Version:      "0.10.2.1",
	}

	if err := config.Transform(cfg, conf); err != nil {
		log.Fatal(err)
	}

	return conf
}

func saramaConfig(kCfg *Config) (*sarama.Config, error) {
	sConfig := sarama.NewConfig()
	sConfig.ClientID = "tcpdog"
	sConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	sConfig.Consumer.Retry.Backoff = time.Duration(kCfg.RetryBackoff) * time.Second
	sConfig.Version = kafkaVersion[kCfg.Version]

	if kCfg.TLSConfig.Enable {
		tlsConfig, err := config.GetTLS(&kCfg.TLSConfig)
		if err != nil {
			return nil, err
		}

		sConfig.Net.TLS.Enable = true
		sConfig.Net.TLS.Config = tlsConfig
	}

	return sConfig, nil
}
