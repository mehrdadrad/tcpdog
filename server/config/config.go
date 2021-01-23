package config

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
	yml "gopkg.in/yaml.v3"

	ac "github.com/mehrdadrad/tcpdog/config"
)

type ctxKey string

// Geo represents a geo
type Geo struct {
	Type   string            `yaml:"type"`
	Config map[string]string `yaml:"config"`
}

// Ingestion represents an ingestion
type Ingestion struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

// Ingress represents an ingress
type Ingress struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

// Flow represents flow from an ingress to an ingestion
type Flow struct {
	Ingress       string
	Ingestion     string
	Serialization string
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enable   bool
	Insecure bool
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
	CAFile   string `yaml:"caFile"`
}

// cliRequest represents cli request
type cliRequest struct {
	Config string
}

// Config represents server configuration
type Config struct {
	Ingress   map[string]Ingress
	Ingestion map[string]Ingestion
	Flow      []Flow
	Geo       Geo
	Log       *zap.Config

	logger *zap.Logger
}

// Logger returns logger
func (c *Config) Logger() *zap.Logger {
	return c.logger
}

// SetMockLogger ...
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

// WithContext returns new context including configuration
func (c *Config) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey("cfg"), c)
}

// FromContext returns configuration from context
func FromContext(ctx context.Context) *Config {
	return ctx.Value(ctxKey("cfg")).(*Config)
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

// Get returns configuration
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

	// get cli request
	cli, err := get(args, version)
	if err != nil {
		return nil, err
	}

	if len(cli.Config) < 1 {
		cli.Config = "/etc/tcpdog/server.yaml"
	}

	config, err = load(cli.Config)
	if err != nil {
		return nil, err
	}

	config.logger = ac.GetLogger(config.Log)

	return config, nil
}

func setDefault(conf *Config) {
	if conf.logger == nil {
		conf.logger = ac.GetDefaultLogger()
	}

	setGeoDefault(&conf.Geo)
}

// Transform converts general configuration to concrete type
func Transform(cfg map[string]interface{}, d interface{}) error {
	return ac.Transform(cfg, d)
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func setGeoDefault(geo *Geo) {
	if geo == nil {
		return
	}

	if geo.Type == "maxmind" && geo.Config["level"] == "" {
		geo.Config["level"] = "city-loc-asn"
	}
}

func validate(c *Config) {
	// TODO serialization validation
}

func GetTLS(cfg *TLSConfig) (*tls.Config, error) {
	return ac.GetTLS(&ac.TLSConfig{
		Enable:   cfg.Enable,
		Insecure: cfg.Insecure,
		CertFile: cfg.CertFile,
		KeyFile:  cfg.KeyFile,
		CAFile:   cfg.CAFile,
	})
}

func GetCreds(cfg *TLSConfig) (credentials.TransportCredentials, error) {
	tlsConfig, err := GetTLS(cfg)
	if err != nil {
		return nil, nil
	}
	return credentials.NewTLS(tlsConfig), nil
}

func GetTLSServer(cfg *TLSConfig) (*tls.Config, error) {
	// TODO
	// mTLS
	return nil, nil
}

func GetCredsServer(cfg *TLSConfig) (credentials.TransportCredentials, error) {
	// TODO
	// mTLS
	return nil, nil
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
