package maxmind

import (
	"net"
	"strconv"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
	"go.uber.org/zap"
)

type cityRecord struct {
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Subdivisions []struct {
		IsoCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
}

type asnRecord struct {
	AutonomousSystemNumber       uint   `maxminddb:"autonomous_system_number"`
	AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
}

// Geo represents Maxmind
type Geo struct {
	logger *zap.Logger
	cityDB *maxminddb.Reader
	asnDB  *geoip2.Reader
}

// New constructs new Maxmind Geo
func New() *Geo {
	return &Geo{}
}

// Init initializes MaxMind database
func (g *Geo) Init(logger *zap.Logger, cfg map[string]string) {
	var err error

	g.cityDB, err = maxminddb.Open(cfg["path-city"])
	if err != nil {
		logger.Error("maxmind", zap.Error(err))
	}

	g.asnDB, err = geoip2.Open(cfg["path-asn"])
	if err != nil {
		logger.Error("maxmind", zap.Error(err))
	}

	g.logger = logger
}

// Get returns ip address geo information
func (g *Geo) Get(ipStr string) map[string]string {
	var (
		cRecord cityRecord
		r       = map[string]string{}
	)

	ip := net.ParseIP(ipStr)
	err := g.cityDB.Lookup(ip, &cRecord)
	if err != nil {
		g.logger.Error("maxmind", zap.Error(err))
	} else {
		r["CCode"] = cRecord.Country.ISOCode
		r["Country"] = cRecord.Country.Names["en"]
		r["City"] = cRecord.City.Names["en"]
	}

	if len(cRecord.Subdivisions) > 0 {
		r["CSCode"] = cRecord.Subdivisions[0].IsoCode
		r["Region"] = cRecord.Subdivisions[0].Names["en"]
	}

	asn, err := g.asnDB.ASN(ip)
	if err != nil {
		g.logger.Error("maxmind", zap.Error(err))
	} else {
		r["ASN"] = strconv.Itoa(int(asn.AutonomousSystemNumber))
		r["ASNOrg"] = asn.AutonomousSystemOrganization
	}

	return r
}
