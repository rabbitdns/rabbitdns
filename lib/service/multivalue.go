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
	"strings"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Multivalue struct {
	Values map[string]Service
	path   string
}

func NewMultivalue(config *Config, path string, v *viper.Viper) (Service, error) {
	multivalue := &Multivalue{Values: map[string]Service{}, path: path}
	values := v.GetStringMap(path + ".values")
	if values == nil {
		return nil, errors.Wrap(ErrConfigParseError, "format error "+path+".type")
	}
	for value := range values {
		newPath := path + ".values." + strings.ToLower(value)
		next, err := CreateService(config, newPath, v)
		if err != nil {
			return nil, err
		}
		multivalue.Values[strings.ToUpper(value)] = next
	}

	return multivalue, nil
}

func (m *Multivalue) Path() string {
	return m.path
}

func (m *Multivalue) GetRR(w dns.ResponseWriter, req *dns.Msg) ([]dns.RR, error) {
	rrs := []dns.RR{}
	status := false
	for _, service := range m.Values {
		endpoint_rrs, err := service.GetRR(w, req)
		if err != nil {
			status = true
			rrs = append(rrs, endpoint_rrs...)
		}
	}
	if status {
		return rrs, nil
	}
	return rrs, ErrServiceStatusError
}

func init() {
	AddServicePlugin("multivalue", NewMultivalue)
}
