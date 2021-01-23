package maxmind

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
	"go.uber.org/zap"
)

const (
	LevelASN = iota
	LevelCity
	LevelCityASN
	LevelCityLoc
	LevelCityLocASN
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

type cityLocRecord struct {
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
	Location struct {
		AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
		Latitude       float64 `maxminddb:"latitude"`
		Longitude      float64 `maxminddb:"longitude"`
		MetroCode      uint    `maxminddb:"metro_code"`
		TimeZone       string  `maxminddb:"time_zone"`
	} `maxminddb:"location"`
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
	level  int
}

var validLevel = map[string]int{
	"asn":          LevelASN,
	"city":         LevelCity,
	"city-asn":     LevelCityASN,
	"city-loc":     LevelCityLoc,
	"city-loc-asn": LevelCityLocASN,
}

// New constructs new Maxmind Geo
func New() *Geo {
	return &Geo{}
}

// Init initializes MaxMind database
func (g *Geo) Init(logger *zap.Logger, cfg map[string]string) {
	var err error

	g.validate(cfg)

	g.level = g.getLevel(cfg["level"])

	if cfg["path-city"] != "" {
		g.cityDB, err = maxminddb.Open(cfg["path-city"])
		if err != nil {
			logger.Fatal("maxmind", zap.Error(err))
		}
	}

	if cfg["path-asn"] != "" {
		g.asnDB, err = geoip2.Open(cfg["path-asn"])
		if err != nil {
			logger.Fatal("maxmind", zap.Error(err))
		}
	}

	g.logger = logger
}

func (g *Geo) getASN(ipStr string) map[string]string {

	var r = map[string]string{}

	ip := net.ParseIP(ipStr)

	asn, err := g.asnDB.ASN(ip)
	if err != nil {
		g.logger.Error("maxmind", zap.Error(err))
	} else {
		r["ASN"] = strconv.Itoa(int(asn.AutonomousSystemNumber))
		r["ASNOrg"] = asn.AutonomousSystemOrganization
	}

	return r
}

func (g *Geo) getCity(ipStr string) map[string]string {
	return g.getCityASN(ipStr, false)
}

func (g *Geo) getCityASN(ipStr string, ASN bool) map[string]string {
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

	if !ASN {
		return r
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

func (g *Geo) getCityLoc(ipStr string) map[string]string {
	return g.getCityLocASN(ipStr, false)
}

func (g *Geo) getCityLocASN(ipStr string, ASN bool) map[string]string {
	var (
		cRecord cityLocRecord
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
		r["GeoLocation"] = fmt.Sprintf("%f,%f", cRecord.Location.Latitude, cRecord.Location.Longitude)
	}

	if len(cRecord.Subdivisions) > 0 {
		r["CSCode"] = cRecord.Subdivisions[0].IsoCode
		r["Region"] = cRecord.Subdivisions[0].Names["en"]
	}

	if !ASN {
		return r
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

// Get returns Geo information
func (g *Geo) Get(ipStr string) map[string]string {
	switch g.level {
	case LevelASN:
		return g.getASN(ipStr)
	case LevelCity:
		return g.getCity(ipStr)
	case LevelCityASN:
		return g.getCityASN(ipStr, true)
	case LevelCityLoc:
		return g.getCityLoc(ipStr)
	case LevelCityLocASN:
		return g.getCityLocASN(ipStr, true)
	}

	return nil
}

func (g *Geo) getLevel(level string) int {
	if v, ok := validLevel[strings.ToLower(level)]; ok {
		return v
	}

	panic("unkown level")
}

func (g *Geo) validate(cfg map[string]string) {
	// TODO
	// check cityDB and asnDB nil if approperiate level requested!
}
