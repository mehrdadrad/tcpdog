package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	yml "gopkg.in/yaml.v3"
)

type ctxKey string

// Config represents tcpstats's config
type Config struct {
	Tracepoints []Tracepoint
	Fields      map[string][]Field
	Egress      map[string]EgressConfig
	Log         *zap.Config

	logger *zap.Logger
}

// EgressConfig represents egress configuration
type EgressConfig struct {
	Type   string
	Config map[string]interface{}
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
	Egress     string
	Config     string
}

func (c *Config) Logger() *zap.Logger {
	return c.logger
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
	Egress   string `yaml:"egress"`
}

// Field represents a field
type Field struct {
	Name   string `yaml:"name"`
	Math   string `yaml:"math,omitempty"`
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

	// set default logger
	if conf.logger == nil {
		conf.logger = getDefaultLogger()
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
			exit(err)
		}

		config.logger = getLogger(config.Log)

		return config
	}

	config, err = cliToConfig(cli)
	if err != nil {
		exit(err)
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
				Egress:   "console",
			},
		},
		Fields: map[string][]Field{
			"cli": cliFieldsStrToSlice(cli.Fields),
		},
		Egress: map[string]EgressConfig{
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

// getDefaultLogger creates default zap logger.
func getDefaultLogger() *zap.Logger {
	var cfg = zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = nil
	cfg.DisableStacktrace = true

	logger, _ := cfg.Build()

	return logger
}

func getLogger(zCfg *zap.Config) *zap.Logger {
	if zCfg == nil {
		return nil
	}

	zCfg.Encoding = "console"
	zCfg.EncoderConfig = zap.NewProductionEncoderConfig()
	zCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zCfg.EncoderConfig.EncodeCaller = nil
	zCfg.DisableStacktrace = true

	logger, err := zCfg.Build()
	if err != nil {
		exit(err)
	}

	return logger
}

func Transform(cfg interface{}, d interface{}) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, d)
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
