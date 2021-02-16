package clickhouse

import (
	"net/url"

	chgo "github.com/ClickHouse/clickhouse-go"

	"github.com/mehrdadrad/tcpdog/config"
)

type chConfig struct {
	DSName   string
	Table    string
	GeoField string

	Columns []string
	Fields  []string

	Connections int // number of connections to database
	Workers     int // number of workers to prepare data

	BatchSize     int
	FlushInterval int
	ConnTimeout   int

	TLSConfig config.TLSConfig // TLS configuration
}

func clickhouseConfig(cfg map[string]interface{}) (*chConfig, error) {

	chConfig := &chConfig{
		DSName:        "tcp://127.0.0.1:9000?username=&debug=true",
		Table:         "tcpdog",
		GeoField:      "SAddr",
		Connections:   1,
		Workers:       2,
		BatchSize:     100,
		FlushInterval: 2,
		ConnTimeout:   300,
	}

	if err := config.Transform(cfg, chConfig); err != nil {
		return nil, err
	}

	if chConfig.TLSConfig.Enable {
		tlsConfig, err := config.GetTLS(&chConfig.TLSConfig)
		if err != nil {
			return nil, err
		}
		err = chgo.RegisterTLSConfig("tcpdog", tlsConfig)
		if err != nil {
			return nil, err
		}
		chConfig.DSName, err = addQString(chConfig.DSName, "tls_config", "tcpdog")
		if err != nil {
			return nil, err
		}
	}

	return chConfig, nil
}

func addQString(dsName, key, value string) (string, error) {
	u, err := url.Parse(dsName)
	if err != nil {
		return "", err
	}

	qString := u.Query()
	qString.Add(key, value)
	u.RawQuery = qString.Encode()

	return u.ResolveReference(u).String(), nil
}
