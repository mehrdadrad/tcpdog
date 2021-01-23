package config

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/url"
	"os"

	"go.uber.org/zap"
	yml "gopkg.in/yaml.v3"
)

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

// cliRequest represents cli request
type serverCLIRequest struct {
	Config string
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Ingress   map[string]Ingress
	Ingestion map[string]Ingestion
	Flow      []Flow
	Geo       Geo
	Log       *zap.Config

	logger *zap.Logger
}

// Logger returns logger
func (c *ServerConfig) Logger() *zap.Logger {
	return c.logger
}

// SetMockLogger sets the in memory logger
func (c *ServerConfig) SetMockLogger(scheme string) *MemSink {
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
func (c *ServerConfig) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey("cfg"), c)
}

// FromContextServer returns configuration from context
func FromContextServer(ctx context.Context) *ServerConfig {
	return ctx.Value(ctxKey("cfg")).(*ServerConfig)
}

// loadServer reads server yaml configuration
func loadServer(file string) (*ServerConfig, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	c := &ServerConfig{}
	err = yml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// GetServer returns configuration
func GetServer(args []string, version string) (*ServerConfig, error) {
	var (
		config *ServerConfig
		err    error
	)

	defer func() {
		if config != nil {
			setDefaultServer(config)
		}
	}()

	// get cli request
	cli, err := getServer(args, version)
	if err != nil {
		return nil, err
	}

	if len(cli.Config) < 1 {
		cli.Config = "/etc/tcpdog/server.yaml"
	}

	config, err = loadServer(cli.Config)
	if err != nil {
		return nil, err
	}

	config.logger = GetLogger(config.Log)

	return config, nil
}

func setDefaultServer(conf *ServerConfig) {
	if conf.logger == nil {
		conf.logger = GetDefaultLogger()
	}

	setGeoDefault(&conf.Geo)
}

func setGeoDefault(geo *Geo) {
	if geo == nil {
		return
	}

	if geo.Type == "maxmind" && geo.Config["level"] == "" {
		geo.Config["level"] = "city-loc-asn"
	}
}
