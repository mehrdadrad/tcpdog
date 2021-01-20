package elasticsearch

import (
	"net"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mehrdadrad/tcpdog/server/config"
)

type esConfig struct {
	URLs          []string // elasticsearch cluster ip addresses
	Username      string   // Username for HTTP Basic Authentication.
	Password      string   // Password for HTTP Basic Authentication.
	CloudID       string   // Endpoint for the Elastic Service (https://elastic.co/cloud).
	APIKey        string   // Base64-encoded token for authorization; if set, overrides username and password.
	Index         string   // elasticsearch index name
	Workers       int      // number of marshaler workers
	FlushBytes    int      // flush threshold in bytes
	FlushInterval int      // periodic flush interval
	GeoField      string   // field supposed to resolve to Geo

	TLSConfig config.TLSConfig // TLS configuration

	clientConfig elasticsearch.Config // elasticsearch HTTP client configuration
}

func elasticSearchConfig(cfg map[string]interface{}) (*esConfig, error) {
	var err error
	// default configuration
	es := &esConfig{
		URLs:          []string{"http://localhost:9200"},
		Index:         "tcpdog",
		GeoField:      "DAddr",
		Workers:       2,
		FlushBytes:    5 & 1 << 20,
		FlushInterval: 30,
	}

	if err := config.Transform(cfg, es); err != nil {
		return nil, err
	}

	// add client config
	es.clientConfig, err = clientConfig(es)

	return es, err
}

func clientConfig(c *esConfig) (elasticsearch.Config, error) {
	cfg := elasticsearch.Config{
		Addresses: c.URLs,
		Username:  c.Username,
		Password:  c.Password,
		CloudID:   c.CloudID,
		APIKey:    c.APIKey,
	}

	if c.TLSConfig.Enable {
		tlsConfig, err := config.GetTLS(&c.TLSConfig)
		if err != nil {
			return elasticsearch.Config{}, err
		}

		cfg.Transport = &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig:       tlsConfig,
		}

	}

	return cfg, nil
}
