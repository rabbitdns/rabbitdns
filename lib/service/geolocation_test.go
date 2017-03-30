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
	"bytes"
	"testing"

	"github.com/miekg/dns"
	"github.com/spf13/viper"
)

func TestGeolocation(t *testing.T) {
	dataStr := []byte(`service:
  type: geolocation
  geodbfile:
    ipv4: ../../MaxMind-DB/test-data/GeoIP2-Country-Test.mmdb
    ipv6: ../../MaxMind-DB/test-data/GeoIP2-Country-Test.mmdb
  locations:
    default:
      type: Endpoint
      value: 192.168.0.1
`)
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(dataStr))
	if err != nil {
		t.Fatalf(err.Error())
	}
	service, err := CreateService(dns.TypeA, "service", v)
	if err != nil {
		t.Fatalf(err.Error())
	}
	geolocation, ok := service.(*Geolocation)
	if ok == false {
		t.Fatalf("geolocation create error %v", geolocation)
	}
	if len(geolocation.Locations) == 0 {
		t.Fatalf("geolocation.Location is empty")
	}
	for loc, next := range geolocation.Locations {
		if loc != "DEFAULT" {
			t.Fatalf("Location is DEFAULT.")
		}
		next, ok := next.(*Endpoint)
		if ok == false {
			t.Fatalf("Location next is Endpoint.")
		}
		if next.Value != "192.168.0.1" {
			t.Fatalf("Endpoint Value is 192.168.0.1")
		}
	}
}
