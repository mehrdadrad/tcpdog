package geo

import (
	"github.com/mehrdadrad/tcpdog/geo/maxmind"
	"go.uber.org/zap"
)

// Reg represents geo registry
var Reg = map[string]Geoer{}

// Geoer represents an IP to Geo provider
type Geoer interface {
	Init(*zap.Logger, map[string]string)
	Get(string) map[string]string
}

func init() {
	Reg["maxmind"] = maxmind.New()
}
