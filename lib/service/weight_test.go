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

func TestWeight(t *testing.T) {
	dataStr := []byte(`service:
  type: weight
  values:
		A:
			weight: 10
			next:
				type: Endpoint
				value: 192.168.0.1
		B:
			weight: 20
			next:
				type: Endpoint
				value: 192.168.0.2
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
	weight, ok := service.(*Weight)
	if ok == false {
		t.Fatalf("Multivalue create error %v", weight)
	}
	if len(weight.Values) == 0 {
		t.Fatalf("geolocation.Values is empty")
	}
	if len(weight.Values) != 2 {
		t.Fatalf("geolocation.Values is not len is 2")
	}
	a, _ := weight.Values["a"]
	a, _ := a.(*WeightChild)
	if a.Weight != 10 {
		t.Fatalf("A weight is not 10")
	}
	b, _ := weight.Values["b"]
	b, _ := a.(*WeightChild)
	if b.Weight != 20 {
		t.Fatalf("B weight is not 10")
	}
	if weight.Weight != 30 {
		t.Fatalf("Total weight is not 30")

	}
}
