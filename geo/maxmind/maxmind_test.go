package maxmind

import (
	"testing"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/stretchr/testify/assert"
)

var cfg = map[string]string{
	"path-city": "./test_data/GeoLite2-City-Test.mmdb",
	"path-asn":  "./test_data/GeoLite2-ASN-Test.mmdb",
}

func TestGetCity(t *testing.T) {
	cfg["level"] = "city"

	c := config.Config{}
	c.SetMockLogger("memory")

	g := New()
	g.Init(c.Logger(), cfg)
	r := g.Get("2.125.160.217")

	assert.Equal(t, "GB", r["CCode"])
	assert.Equal(t, "ENG", r["CSCode"])
	assert.Equal(t, "Boxford", r["City"])
	assert.Equal(t, "United Kingdom", r["Country"])
	assert.Equal(t, "England", r["Region"])
}

func TestGetASN(t *testing.T) {
	cfg["level"] = "asn"

	c := config.Config{}
	c.SetMockLogger("memory")

	g := New()
	g.Init(c.Logger(), cfg)

	r := g.Get("70.160.0.1")
	assert.Equal(t, "22773", r["ASN"])
	assert.Equal(t, "Cox Communications Inc.", r["ASNOrg"])
}
func TestGetCityASN(t *testing.T) {
	cfg["level"] = "city-asn"

	c := config.Config{}
	c.SetMockLogger("memory")

	g := New()
	g.Init(c.Logger(), cfg)

	r := g.Get("70.160.0.1")
	assert.Equal(t, "22773", r["ASN"])
	assert.Equal(t, "Cox Communications Inc.", r["ASNOrg"])
}

func TestGetCityLoc(t *testing.T) {
	cfg["level"] = "city-loc"

	c := config.Config{}
	c.SetMockLogger("memory")

	g := New()
	g.Init(c.Logger(), cfg)

	r := g.Get("2.125.160.217")
	assert.Equal(t, "GB", r["CCode"])
	assert.Equal(t, "ENG", r["CSCode"])
	assert.Equal(t, "Boxford", r["City"])
	assert.Equal(t, "United Kingdom", r["Country"])
	assert.Equal(t, "England", r["Region"])
	assert.Equal(t, "51.750000,-1.250000", r["GeoLocation"])
}

func TestGetCityLocASN(t *testing.T) {
	cfg["level"] = "city-loc-asn"

	c := config.Config{}
	c.SetMockLogger("memory")

	g := New()
	g.Init(c.Logger(), cfg)

	r := g.Get("70.160.0.1")
	assert.Equal(t, "22773", r["ASN"])
	assert.Equal(t, "Cox Communications Inc.", r["ASNOrg"])
}
func BenchmarkMaxmindParallel(b *testing.B) {
	c := config.Config{}
	c.SetMockLogger("memory")

	g := New()
	g.Init(c.Logger(), cfg)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.Get("2.125.160.217")
		}
	})
}

func BenchmarkMaxmind(b *testing.B) {
	c := config.ServerConfig{}
	c.SetMockLogger("memory")

	g := New()
	g.Init(c.Logger(), cfg)

	for i := 0; i < b.N; i++ {
		g.Get("68.170.74.242")
	}
}
