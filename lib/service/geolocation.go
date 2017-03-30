// Copyright (C) 2018 Manabu Sonoda.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"net"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	. "github.com/rabbitdns/rabbitdns/lib/misc"

	"github.com/miekg/dns"
	maxminddb "github.com/oschwald/maxminddb-golang"
)

var (
	ErrNextEmpty            = errors.New("Rule is empty")
	ErrLocatiosEmpty        = errors.New("Location is empty")
	ErrLocatiosCantUse      = errors.New("Location is can't use")
	ErrLocationsEmpty       = errors.New("Locations is empty")
	ErrGeoip2DBFileEmpty    = errors.New("Geoip2DBFile is empty")
	ErrGeoip2DBCantOpen     = errors.New("geodb can't open")
	ErrDefaultLocationEmpty = errors.New("DEFAULT location is not found")
)

type (
	City struct {
		GeoNameID uint              `maxminddb:"geoname_id"`
		Names     map[string]string `maxminddb:"names"`
	}
	Continent struct {
		Code      string            `maxminddb:"code"`
		GeoNameID uint              `maxminddb:"geoname_id"`
		Names     map[string]string `maxminddb:"names"`
	}
	Country struct {
		GeoNameID uint              `maxminddb:"geoname_id"`
		IsoCode   string            `maxminddb:"iso_code"`
		Names     map[string]string `maxminddb:"names"`
	}
	Subdivision struct {
		GeoNameID uint              `maxminddb:"geoname_id"`
		IsoCode   string            `maxminddb:"iso_code"`
		Names     map[string]string `maxminddb:"names"`
	}
	Records struct {
		City         City          `maxminddb:"city"`
		Continent    Continent     `maxminddb:"continent"`
		Country      Country       `maxminddb:"country"`
		Subdivisions []Subdivision `maxminddb:"subdivisions"`
	}
	Geolocation struct {
		IPv4GeoDB *maxminddb.Reader
		IPv6GeoDB *maxminddb.Reader
		Locations map[string]Service
		path      string
	}
)

func readGeoDB(filepath string) (*maxminddb.Reader, error) {
	db, err := maxminddb.Open(filepath)
	if err != nil {
		return nil, ErrGeoip2DBCantOpen
	}
	return db, nil
}
func NewGeolocation(config *Config, path string, v *viper.Viper) (Service, error) {
	var err error
	geolocation := &Geolocation{Locations: map[string]Service{}, path: path}

	geodbfile := v.Get(path + ".geodbfile")

	if geodbfile == nil {
		return nil, ErrGeoip2DBFileEmpty
	}
	switch db := geodbfile.(type) {
	case string:
		geolocation.IPv4GeoDB, err = maxminddb.Open(db)
		if err != nil {
			return nil, ErrGeoip2DBCantOpen
		}
	case map[string]interface{}:
		if filePath, ok := db["ipv4"].(string); ok {
			geolocation.IPv4GeoDB, err = maxminddb.Open(filePath)
			if err != nil {
				return nil, ErrGeoip2DBCantOpen
			}
		}
		if filePath, ok := db["ipv6"].(string); ok {
			geolocation.IPv6GeoDB, err = maxminddb.Open(filePath)
			if err != nil {
				return nil, ErrGeoip2DBCantOpen
			}
		}
	default:
		return nil, errors.Wrap(ErrConfigParseError, "geodb format error")
	}

	locations := v.GetStringMap(path + ".locations")
	if locations == nil {
		return nil, ErrLocationsEmpty
	}

	for location := range locations {
		newPath := path + ".locations." + strings.ToLower(location)
		next, err := CreateService(config, newPath, v)
		if err != nil {
			return nil, err
		}
		geolocation.Locations[strings.ToUpper(location)] = next
	}

	if _, exist := geolocation.Locations["DEFAULT"]; exist == false {
		return nil, ErrDefaultLocationEmpty
	}

	return geolocation, nil
}
func (g *Geolocation) Path() string {
	return g.path
}

func (g *Geolocation) GetRR(w dns.ResponseWriter, req *dns.Msg) ([]dns.RR, error) {
	for _, record := range g.getNetwork(w, req) {
		// subdivision check
		for location, next := range g.Locations {
			for _, sub := range record.Subdivisions {
				if sub.IsoCode == location {
					return next.GetRR(w, req)
				}
			}
		}
		// country check
		for location, next := range g.Locations {
			if record.Country.IsoCode == location {
				return next.GetRR(w, req)
			}
		}
		// Continent check
		for location, next := range g.Locations {
			if record.Continent.Code == location {
				return next.GetRR(w, req)
			}
		}
	}
	// default
	for location, next := range g.Locations {
		if location == "DEFAULT" {
			return next.GetRR(w, req)
		}
	}
	return []dns.RR{}, nil
}

func (g *Geolocation) getNetwork(w dns.ResponseWriter, req *dns.Msg) []Records {
	var err error
	// EDNS CLIENT SUBNET
	records := []Records{}
	for _, rr := range req.Extra {
		if opt, ok := rr.(*dns.OPT); ok {
			for _, o := range opt.Option {
				if ecs, ok := o.(*dns.EDNS0_SUBNET); ok {
					record := Records{}
					switch ecs.Family {
					case 1:
						err = g.IPv4GeoDB.Lookup(ecs.Address, record)
					case 2:
						err = g.IPv6GeoDB.Lookup(ecs.Address, record)
					default:
						continue
					}
					if err != nil {
						continue
					}
					records = append(records, record)
				}
			}
		}
	}
	addr := w.RemoteAddr()
	record := Records{}
	switch IpFamily(addr.String()) {
	case 4:
		err = g.IPv4GeoDB.Lookup(net.ParseIP(addr.String()), record)
	case 6:
		err = g.IPv6GeoDB.Lookup(net.ParseIP(addr.String()), record)
	}
	return records
}

func init() {
	AddServicePlugin("geolocation", NewGeolocation)
}
