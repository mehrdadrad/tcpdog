package config

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	yml "gopkg.in/yaml.v3"
)

type ctxKey string

// Config represents tcpstats's config
type Config struct {
	Tracepoints []Tracepoint
	Fields      map[string][]Field
	Output      map[string]OutputConfig
}

// OutputConfig represents output's configuration
type OutputConfig struct {
	Type   string
	Config map[string]string
}

// WithContext returns new context including configuration
func (c *Config) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey("cfg"), c)
}

// FromContext returns configuration from context
func FromContext(ctx context.Context) *Config {
	return ctx.Value(ctxKey("cfg")).(*Config)
}

// Tracepoint represents a tracepoint's config
type Tracepoint struct {
	Name            string `yaml:"name"`
	Fields          string `yaml:"fields"`
	TCPState        string `yaml:"tcp_state"`
	PollingInterval int    `yaml:"pollingInterval"`
	Inet            []int  `yaml:"inet"`
	Geo             string `yaml:"geo"`
	Output          string `yaml:"output"`
}

// Field represents a field
type Field struct {
	Name   string `yaml:"name"`
	Func   string `yaml:"func,omitempty"`
	Filter string `yaml:"filter,omitempty"`
}

// GetTPFields returns a tracepoint fields slice
func (c *Config) GetTPFields(name string) []string {
	fields := []string{}
	if v, ok := c.Fields[name]; ok {
		for _, f := range v {
			fields = append(fields, f.Name)
		}
	}

	return fields
}

// Load reads yaml configuration
func Load() *Config {
	file, err := os.Open("./config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	c := &Config{}
	err = yml.Unmarshal(b, c)
	if err != nil {
		log.Fatal(err)
	}

	setDefault(c)

	return c
}

func setDefault(conf *Config) {
	for _, tp := range conf.Tracepoints {
		if len(tp.Inet) < 1 {
			tp.Inet = append(tp.Inet, 4)
		}
	}
}
