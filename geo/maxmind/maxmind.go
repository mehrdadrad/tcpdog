package maxmind

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
	"go.uber.org/zap"
)

const (
	// LevelASN : resolve IP to ASN information
	LevelASN = iota + 1
	// LevelCity : resolve IP to city/country information
	LevelCity
	// LevelCityASN : resolve IP to city/country and ASN information
	LevelCityASN
	// LevelCityLoc : resolve IP to city/country/lat/lon information
	LevelCityLoc
	// LevelCityLocASN : resolve IP to city/country/lat/lon and ASN information
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
	fn     func(string) map[string]string
	level  int
	isASN  bool
}

var str2Level = map[string]int{
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

	g.level = str2Level[strings.ToLower(cfg["level"])]

	if err := g.validate(cfg); err != nil {
		logger.Fatal("maxmind", zap.Error(err))
	}

	if g.level != LevelASN {
		g.cityDB, err = maxminddb.Open(cfg["path-city"])
		if err != nil {
			logger.Fatal("maxmind", zap.Error(err))
		}
	}

	if g.level == LevelASN || g.level == LevelCityASN || g.level == LevelCityLocASN {
		g.isASN = true
		g.asnDB, err = geoip2.Open(cfg["path-asn"])
		if err != nil {
			logger.Fatal("maxmind", zap.Error(err))
		}
	}

	g.fn = g.getFunc()

	g.logger = logger

	logger.Info("geo", zap.String("msg", "maxmind has been initialized"))
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

func (g *Geo) getCityASN(ipStr string) map[string]string {
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

	if !g.isASN {
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

func (g *Geo) getCityLocASN(ipStr string) map[string]string {
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

	if !g.isASN {
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
	return g.fn(ipStr)
}

func (g *Geo) getFunc() func(string) map[string]string {
	switch g.level {
	case LevelASN:
		return g.getASN
	case LevelCity, LevelCityASN:
		return g.getCityASN
	case LevelCityLoc, LevelCityLocASN:
		return g.getCityLocASN
	}

	return nil
}

func (g *Geo) validate(cfg map[string]string) error {
	switch g.level {
	case LevelASN:
		if v, isPathASN := cfg["path-asn"]; !isPathASN || len(v) < 1 {
			return errors.New("the maxmind path-asn has not configured")
		}
	case LevelCity, LevelCityLoc:
		if v, isPathCity := cfg["path-city"]; !isPathCity || len(v) < 1 {
			return errors.New("the maxmind path-city has not configured")
		}
	case LevelCityASN, LevelCityLocASN:
		if v, isPathASN := cfg["path-asn"]; !isPathASN || len(v) < 1 {
			return errors.New("the maxmind path-asn has not configured")
		}
		if v, isPathCity := cfg["path-city"]; !isPathCity || len(v) < 1 {
			return errors.New("the maxmind path-city has not configured")
		}
	default:
		return errors.New("unknown maxmind/geo level")
	}

	return nil
}
