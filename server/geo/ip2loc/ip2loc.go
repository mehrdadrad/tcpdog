package ip2loc

import (
	"log"

	"github.com/ip2location/ip2location-go"
)

var db *ip2location.DB

// Geo represents ip2location
type Geo struct{}

// New constructs Geo
func New() *Geo {
	return &Geo{}
}

// Init initializes ip2location database
func (i *Geo) Init(cfg map[string]string) {
	var err error

	db, err = ip2location.OpenDB(cfg["path"])
	if err != nil {
		log.Fatal(err)
	}
}

// Get returns ip address geo information
func (i *Geo) Get(ip string) map[string]string {
	info, err := db.Get_all(ip)
	if err != nil {
		log.Println(err)
		return nil
	}
	return map[string]string{
		"counter_code": info.Country_short,
		"country":      info.Country_long,
		"region":       info.Region,
		"city":         info.City,
	}
}
