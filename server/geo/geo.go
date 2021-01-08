package geo

import (
	"github.com/mehrdadrad/tcpdog/server/geo/ip2loc"
)

var Reg = map[string]Geoer{}

type Geoer interface {
	Init(map[string]string)
	Get(string) map[string]string
}

func init() {
	Reg["ip2loc"] = ip2loc.New()
}
