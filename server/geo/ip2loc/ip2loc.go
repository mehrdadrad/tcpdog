package ip2loc

import (
	"log"

	"github.com/ip2location/ip2location-go"
)

var db *ip2location.DB

// Geo represents ip2location
type Geo struct {
	usCodes map[string]string
}

// New constructs Geo
func New() *Geo {
	return &Geo{
		usCodes: usCodes(),
	}
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
		"CCode":   info.Country_short,
		"Country": info.Country_long,
		"Region":  info.Region,
		"City":    info.City,
		"USCode":  i.usCodes[info.Region],
	}
}

func usCodes() map[string]string {
	return map[string]string{
		"Alabama":              "AL",
		"Alaska":               "AK",
		"Arizona":              "AZ",
		"Arkansas":             "AR",
		"California":           "CA",
		"Colorado":             "CO",
		"Connecticut":          "CT",
		"Delaware":             "DE",
		"District Of Columbia": "DC",
		"Florida":              "FL",
		"Georgia":              "GA",
		"Hawaii":               "HI",
		"Idaho":                "ID",
		"Illinois":             "IL",
		"Indiana":              "IN",
		"Iowa":                 "IA",
		"Kansas":               "KS",
		"Kentucky":             "KY",
		"Louisiana":            "LA",
		"Maine":                "ME",
		"Maryland":             "MD",
		"Massachusetts":        "MA",
		"Michigan":             "MI",
		"Minnesota":            "MN",
		"Mississippi":          "MS",
		"Missouri":             "MO",
		"Montana":              "MT",
		"Nebraska":             "NE",
		"Nevada":               "NV",
		"New Hampshire":        "NH",
		"New Jersey":           "NJ",
		"New Mexico":           "NM",
		"New York":             "NY",
		"North Carolina":       "NC",
		"North Dakota":         "ND",
		"Ohio":                 "OH",
		"Oklahoma":             "OK",
		"Oregon":               "OR",
		"Pennsylvania":         "PA",
		"Rhode Island":         "RI",
		"South Carolina":       "SC",
		"South Dakota":         "SD",
		"Tennessee":            "TN",
		"Texas":                "TX",
		"Utah":                 "UT",
		"Vermont":              "VT",
		"Virginia":             "VA",
		"Washington":           "WA",
		"West Virginia":        "WV",
		"Wisconsin":            "WI",
		"Wyoming":              "WY",
	}
}
