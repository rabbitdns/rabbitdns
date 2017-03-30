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

func TestMultivalue(t *testing.T) {
	dataStr := []byte(`service:
  type: multivalue
  values:
    A:
    	type: Endpoint
			value: 192.168.0.1
		B:
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
	multivalue, ok := service.(*Multivalue)
	if ok == false {
		t.Fatalf("Multivalue create error %v", multivalue)
	}
	if len(multivalue.Values) == 0 {
		t.Fatalf("geolocation.Values is empty")
	}
	if len(multivalue.Values) != 2 {
		t.Fatalf("must geolocation.Values len is 2")
	}
	for _, val := range geolocation.Values {
		next, ok := val.(*Endpoint)
		if ok == false {
			t.Fatalf("Location next is Endpoint.")
		}
	}
}
