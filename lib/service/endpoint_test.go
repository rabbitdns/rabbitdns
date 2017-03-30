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

func TestEndpointA(t *testing.T) {
	dataStr := []byte(`service:
  type: endpoint
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
	endpoint, ok := service.(*Endpoint)
	if ok == false {
		t.Fatalf("Service is not Endpoint")
	}
	_, ok = endpoint.RR.(*dns.A)
	if ok == false {
		t.Fatalf("endpoint RR is not *dns.A")
	}

	service, err = CreateService(dns.TypeAAAA, "service", v)
	if err == nil {
		t.Errorf("Value is not IPv4, need return error")
	}

}

func TestEndpointAAAA(t *testing.T) {
	dataStr := []byte(`service:
  type: endpoint
  value: 2001:db8::1
`)
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(dataStr))
	if err != nil {
		t.Fatalf(err.Error())
	}

	service, err := CreateService(dns.TypeAAAA, "service", v)
	if err != nil {
		t.Fatalf(err.Error())
	}
	endpoint, ok := service.(*Endpoint)
	if ok == false {
		t.Fatalf("Service is not Endpoint")
	}
	_, ok = endpoint.RR.(*dns.AAAA)
	if ok == false {
		t.Fatalf("endpoint RR is not *dns.AAAA")
	}

	service, err = CreateService(dns.TypeA, "service", v)
	if err == nil {
		t.Errorf("Value is not IPv6, need return error")
	}

}

func TestEndpointTXT(t *testing.T) {
	dataStr := []byte(`service:
  type: endpoint
  value: おこです
`)
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(dataStr))
	if err != nil {
		t.Fatalf(err.Error())
	}

	service, err := CreateService(dns.TypeTXT, "service", v)
	if err != nil {
		t.Fatalf(err.Error())
	}
	endpoint, ok := service.(*Endpoint)
	if ok == false {
		t.Fatalf("Service is not Endpoint")
	}
	_, ok = endpoint.RR.(*dns.TXT)
	if ok == false {
		t.Fatalf("endpoint RR is not *dns.TXT")
	}

}
