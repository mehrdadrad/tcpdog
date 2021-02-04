package config

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTransform(t *testing.T) {
	cfg := map[string]interface{}{
		"foo": "bar",
		"key": 5,
	}

	d := &struct {
		Foo string
		Key int
	}{}

	err := Transform(cfg, d)

	assert.NoError(t, err)
	assert.Equal(t, "bar", d.Foo)
	assert.Equal(t, 5, d.Key)

	err = Transform(make(chan int), d)
	assert.Error(t, err)
}

func TestGetLogger(t *testing.T) {
	rawJSON := []byte(`{"level":"info", "encoding": "console", "outputPaths": ["stdout"], "errorOutputPaths":["stderr"]}`)
	zCfg := &zap.Config{}
	json.Unmarshal(rawJSON, zCfg)

	assert.NotNil(t, GetLogger(zCfg))
	assert.Nil(t, GetLogger(nil))
}

func TestGetDefaultLogger(t *testing.T) {
	assert.NotNil(t, GetDefaultLogger())
}

func TestLoad(t *testing.T) {
	ymlContent := `tracepoints:
  - name: sock:inet_sock_set_state
    fields: fields_01
    tcp_state: TCP_CLOSE
    sample: 0
    inet: [4,6]
    egress: console`

	filename := t.TempDir() + "/config.yml"
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	assert.NoError(t, err)
	defer f.Close()
	io.WriteString(f, ymlContent)

	cfg, err := load(filename)
	assert.NoError(t, err)

	assert.Len(t, cfg.Tracepoints, 1)
	assert.Equal(t, "sock:inet_sock_set_state", cfg.Tracepoints[0].Name)
	assert.Equal(t, "fields_01", cfg.Tracepoints[0].Fields)

	// wrong file
	_, err = load("not_exist")
	assert.Error(t, err)

	// wrong yaml
	io.WriteString(f, "\t\nabcde\n")
	_, err = load(filename)
	assert.Error(t, err)
}

func TestGetTPFields(t *testing.T) {
	c := &Config{
		Fields: map[string][]Field{
			"foo": {{Name: "f1"}, {Name: "f2"}},
		},
	}

	s := c.GetTPFields("foo")
	assert.Len(t, s, 2)
	assert.Equal(t, "f1", s[0])
	assert.Equal(t, "f2", s[1])
}

func TestSetDefault(t *testing.T) {
	c := &Config{
		Tracepoints: []Tracepoint{{Name: "foo"}},
	}
	setDefault(c)
	assert.Equal(t, 1, c.Tracepoints[0].Workers)
	assert.NotNil(t, c.logger)
}

func TestCliToConfig(t *testing.T) {
	cli := &cliRequest{
		Tracepoint: "tracepoint1",
		Fields:     []string{"f1", "f2"},
		TCPState:   "foo",
		Egress:     "bar",
		IPv4:       true,
		IPv6:       true,
	}

	c, err := cliToConfig(cli)
	assert.NoError(t, err)
	assert.Len(t, c.Fields["cli"], 2)
	assert.Equal(t, "f1", c.Fields["cli"][0].Name)
	assert.Equal(t, "f2", c.Fields["cli"][1].Name)
	assert.Equal(t, "foo", c.Tracepoints[0].TCPState)
	assert.Len(t, c.Tracepoints[0].INet, 2)
}

func TestGet(t *testing.T) {
	os.Setenv("TCPDOG_TEST", "true")
	c, err := Get([]string{"tcpdog", "-fields", "SAddr,RTT", "-state", "TCP_FOO"}, "0.0.0")
	assert.NoError(t, err)
	assert.Equal(t, "SAddr", c.Fields["cli"][0].Name)
	assert.Equal(t, "RTT", c.Fields["cli"][1].Name)
	assert.Equal(t, 4, c.Tracepoints[0].INet[0])
	assert.Equal(t, "TCP_FOO", c.Tracepoints[0].TCPState)

	// config option
	filename := t.TempDir() + "/config.yml"
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	assert.NoError(t, err)
	defer f.Close()
	io.WriteString(f, "")
	c, err = Get([]string{"tcpdog", "-config", filename}, "0.0.0")
	assert.NoError(t, err)

	// wrong config file
	c, err = Get([]string{"tcpdog", "-config", "foo"}, "0.0.0")
	assert.Error(t, err)
}

func TestGetTLSCreds(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"foo"},
		},

		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 1),

		DNSNames: []string{"foo.com"},

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derCertBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: derCertBytes})
	certPem := buf.String()

	buf.Reset()

	pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: derCertBytes})
	caPem := buf.String()

	buf.Reset()

	pem.Encode(buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	privateKeyPem := buf.String()

	tmpDir := t.TempDir()

	certFile, err := os.OpenFile(tmpDir+"/certFile", os.O_RDWR|os.O_CREATE, 0755)
	assert.NoError(t, err)
	certFile.WriteString(certPem)

	caFile, err := os.OpenFile(tmpDir+"/caFile", os.O_RDWR|os.O_CREATE, 0755)
	assert.NoError(t, err)
	caFile.WriteString(caPem)

	keyFile, err := os.OpenFile(tmpDir+"/keyFile", os.O_RDWR|os.O_CREATE, 0755)
	assert.NoError(t, err)
	keyFile.WriteString(privateKeyPem)

	cfg := &TLSConfig{
		Enable:   true,
		CertFile: certFile.Name(),
		KeyFile:  keyFile.Name(),
		CAFile:   caFile.Name(),
	}

	tlsConfig, err := GetTLS(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, tlsConfig)

	_, err = GetCreds(cfg)
	assert.NoError(t, err)

	// wrong files
	cfg.CAFile = "foo"
	_, err = GetTLS(cfg)
	assert.Error(t, err)

	cfg.KeyFile = ""
	cfg.CertFile = "foo"
	_, err = GetTLS(cfg)
	assert.Error(t, err)

	_, err = GetCreds(cfg)
	assert.Error(t, err)
}

func TestLogMemSink(t *testing.T) {
	m := &MemSink{}
	m.Buffer = bytes.NewBufferString(`{"foo":"bar"}`)
	assert.Contains(t, m.Unmarshal(), "foo")
	m.Sync()
	m.Close()
}

func TestSetMockLogger(t *testing.T) {
	c := &Config{}
	ms := c.SetMockLogger("memory")
	assert.NotNil(t, c.logger)
	assert.NotNil(t, ms)
}

func TestConfigContextLogger(t *testing.T) {
	c := &Config{
		Fields: map[string][]Field{
			"foo": {{Name: "f1"}, {Name: "f2"}},
		},
	}

	c.logger = GetDefaultLogger()

	ctx := c.WithContext(context.Background())
	cFromCTX := FromContext(ctx)
	assert.Equal(t, c, cFromCTX)

	logger := c.Logger()
	assert.Equal(t, c.logger, logger)
}
