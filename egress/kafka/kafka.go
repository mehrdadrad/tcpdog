package kafka

import (
	"bytes"
	"context"
	"sync"

	"github.com/mehrdadrad/tcpdog/config"
)

// https://pkg.go.dev/google.golang.org/protobuf/encoding/protojson

// KafkaConfig represents Kafka configuration
type KafkaConfig struct {
	Brokers        []string `yaml:"brokers"`
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

func New(ctx context.Context, tp config.Tracepoint, bufpool *sync.Pool, ch chan *bytes.Buffer) error {
	//cfg := config.FromContext(ctx)
	// := cfg.Egress[tp.Egress].Config
	//_ = egress
	return nil
}

func kafkaConf(cfg map[string]interface{}) {

}
