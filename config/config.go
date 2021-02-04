package config

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/credentials"
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

// TLSConfig represents TLS configuration.
type TLSConfig struct {
	Enable   bool
	Insecure bool
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
	CAFile   string `yaml:"caFile"`
}

// EgressConfig represents egress configuration.
type EgressConfig struct {
	Type   string
	Config map[string]interface{}
}

// cliRequest represents cli requests.
type cliRequest struct {
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

// Logger returns logger.
func (c *Config) Logger() *zap.Logger {
	return c.logger
}

// WithContext returns new context including configuration.
func (c *Config) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey("cfg"), c)
}

// SetMockLogger sets in memory logger
func (c *Config) SetMockLogger(scheme string) *MemSink {
	var err error

	cfg := zap.NewDevelopmentConfig()
	ms := &MemSink{new(bytes.Buffer)}
	zap.RegisterSink(scheme, func(*url.URL) (zap.Sink, error) {
		return ms, nil
	})

	cfg.OutputPaths = []string{scheme + "://"}
	cfg.DisableStacktrace = true
	cfg.Encoding = "json"

	c.logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}

	return ms
}

// FromContext returns configuration from context.
func FromContext(ctx context.Context) *Config {
	return ctx.Value(ctxKey("cfg")).(*Config)
}

// Tracepoint represents a tracepoint's config.
type Tracepoint struct {
	Name     string `yaml:"name"`
	Fields   string `yaml:"fields"`
	TCPState string `yaml:"tcp_state"`
	Sample   int    `yaml:"sample"`
	Workers  int    `yaml:"workers"`
	INet     []int  `yaml:"inet"`
	Egress   string `yaml:"egress"`
}

// Field represents a field.
type Field struct {
	Name   string `yaml:"name"`
	Math   string `yaml:"math,omitempty"`
	Filter string `yaml:"filter,omitempty"`
}

// GetTPFields returns a tracepoint fields.
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
		if len(conf.Tracepoints[i].INet) < 1 {
			conf.Tracepoints[i].INet = append(conf.Tracepoints[i].INet, 4)
		}
		if conf.Tracepoints[i].Workers < 1 {
			conf.Tracepoints[i].Workers = 1
		}
	}

	// set default logger
	if conf.logger == nil {
		conf.logger = GetDefaultLogger()
	}
}

// Get returns the configuration based on the file or cli
func Get(args []string, version string) (*Config, error) {
	var (
		config *Config
		err    error
	)

	defer func() {
		if config != nil {
			setDefault(config)
		}
	}()

	cli, err := get(args, version)
	if err != nil {
		return nil, err
	}

	if cli.Config != "" {
		config, err = load(cli.Config)
		if err != nil {
			return nil, err
		}

		config.logger = GetLogger(config.Log)

		return config, nil
	}

	config, err = cliToConfig(cli)

	return config, err
}

func cliToConfig(cli *cliRequest) (*Config, error) {
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
				INet:     inet,
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

// GetDefaultLogger creates default zap logger.
func GetDefaultLogger() *zap.Logger {
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

// GetLogger returns logger based on the configuration.
func GetLogger(zCfg *zap.Config) *zap.Logger {
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

// Transform transforms one data structure to another
// it uses to transform map[string]interface{} config
// to specific struct.
func Transform(cfg interface{}, d interface{}) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, d)
}

// GetTLS returns tls.config based on the configuration.
func GetTLS(cfg *TLSConfig) (*tls.Config, error) {
	var (
		tlsConfig  = &tls.Config{}
		caCertPool *x509.CertPool
	)

	if cfg.CertFile != "" {
		if cfg.KeyFile == "" {
			cfg.KeyFile = cfg.CertFile
		}

		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, err
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if cfg.CAFile != "" {
		caCert, err := ioutil.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}

		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig.RootCAs = caCertPool
	}

	tlsConfig.InsecureSkipVerify = cfg.Insecure

	return tlsConfig, nil
}

// GetCreds returns transport credentials based on the tls config.
func GetCreds(cfg *TLSConfig) (credentials.TransportCredentials, error) {
	tlsConfig, err := GetTLS(cfg)
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(tlsConfig), nil
}

// MemSink represents logging in memory
type MemSink struct {
	*bytes.Buffer
}

// Close is required method for sink interface.
func (s *MemSink) Close() error { return nil }

// Sync is required method for sink interface.
func (s *MemSink) Sync() error { return nil }

// Unmarshal returns decoded data as key value and reset the buffer.
func (s *MemSink) Unmarshal() map[string]string {
	defer s.Reset()
	v := make(map[string]string)
	json.Unmarshal(s.Bytes(), &v)
	return v
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
