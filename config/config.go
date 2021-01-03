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

// CLIRequest represents cli requests.
type CLIRequest struct {
	Tracepoint string
	Fields     []string
	IPv4       bool
	IPv6       bool
	Workers    int
	Sample     int
	TCPState   string
	Output     string
	Config     string
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
	Name     string `yaml:"name"`
	Fields   string `yaml:"fields"`
	TCPState string `yaml:"tcp_state"`
	Sample   int    `yaml:"sample"`
	Workers  int    `yaml:"workers"`
	Inet     []int  `yaml:"inet"`
	Geo      string `yaml:"geo"`
	Output   string `yaml:"output"`
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

// load reads yaml configuration
func load(file string) (*Config, error) {
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

func setDefault(conf *Config) {
	for i := range conf.Tracepoints {
		if len(conf.Tracepoints[i].Inet) < 1 {
			conf.Tracepoints[i].Inet = append(conf.Tracepoints[i].Inet, 4)
		}
		if conf.Tracepoints[i].Workers < 1 {
			conf.Tracepoints[i].Workers = 1
		}
	}
}

// Get returns the configuration based on the file or cli
func Get(cli *CLIRequest) *Config {
	var (
		config *Config
		err    error
	)

	defer func() {
		setDefault(config)
	}()

	if cli.Config != "" {
		config, err = load(cli.Config)
		if err != nil {
			log.Fatal(err)
		}
		return config
	}

	config, err = cliToConfig(cli)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func cliToConfig(cli *CLIRequest) (*Config, error) {
	var inet []int

	if cli.IPv4 {
		inet = append(inet, 4)
	}
	if cli.IPv6 {
		inet = append(inet, 6)
	}

	config := &Config{
		Tracepoints: []Tracepoint{
			{
				Name:     cli.Tracepoint,
				Fields:   "cli",
				TCPState: cli.TCPState,
				Workers:  cli.Workers,
				Sample:   cli.Sample,
				Inet:     inet,
				Output:   "console",
			},
		},
		Fields: map[string][]Field{
			"cli": cliFieldsStrToSlice(cli.Fields),
		},
		Output: map[string]OutputConfig{
			"console": {
				Type: "console",
			},
		},
	}

	return config, nil
}

func cliFieldsStrToSlice(fs []string) []Field {
	fields := []Field{}

	for _, f := range fs {
		fields = append(fields, Field{Name: f})
	}

	return fields
}
