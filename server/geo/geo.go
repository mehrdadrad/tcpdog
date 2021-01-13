package geo

import (
	"github.com/mehrdadrad/tcpdog/server/geo/ip2loc"
	"github.com/mehrdadrad/tcpdog/server/geo/maxmind"
	"go.uber.org/zap"
)

var Reg = map[string]Geoer{}

// Geoer represents an IP to Geo provider
type Geoer interface {
	Init(*zap.Logger, map[string]string)
	Get(string) map[string]string
}

func init() {
	Reg["ip2loc"] = ip2loc.New()
	Reg["maxmind"] = maxmind.New()
}
