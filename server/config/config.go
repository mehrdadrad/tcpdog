package config

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	yml "gopkg.in/yaml.v3"
)

type ctxKey string

type Config struct {
	Geo    Geo
	Ingest Ingest
}

type Geo struct {
	Type   string            `yaml:"type"`
	Config map[string]string `yaml:"config"`
}

type Ingest struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

// WithContext returns new context including configuration
func (c *Config) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey("cfg"), c)
}

// FromContext returns configuration from context
func FromContext(ctx context.Context) *Config {
	return ctx.Value(ctxKey("cfg")).(*Config)
}

// Load reads yaml configuration
func Load(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func Transform(cfg map[string]interface{}, d interface{}) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, d)
}
