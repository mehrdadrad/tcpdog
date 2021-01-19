package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/credentials"
	yml "gopkg.in/yaml.v3"
)

type ctxKey string

// Config represents server configuration
type Config struct {
	Ingress   map[string]Ingress
	Ingestion map[string]Ingestion
	Flow      []Flow
	Geo       *Geo
	Log       *zap.Config

	logger *zap.Logger
}

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

// Logger returns logger
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

	c.logger = getLogger(c.Log)

	setDefault(c)

	return c, nil
}

func setDefault(conf *Config) {
	// set default logger
	if conf.logger == nil {
		conf.logger = getDefaultLogger()
	}

	// geo
	setGeoDefault(conf.Geo)
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

// Transform converts general configuration to concrete type
func Transform(cfg map[string]interface{}, d interface{}) error {
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

func GetCreds(cfg *TLSConfig) (credentials.TransportCredentials, error) {
	tlsConfig, err := GetTLS(cfg)
	if err != nil {
		return nil, nil
	}
	return credentials.NewTLS(tlsConfig), nil
}
