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

var (
	ErrZeroWeight = errors.New("Weight value is zero or empty")
)

type Weight struct {
	Values map[string]*WeightValue
	path   string
}

type WeightValue struct {
	Weight uint8
	Next   Service
	path   string
}

func NewWeight(config *Config, path string, v *viper.Viper) (Service, error) {
	weight := &Weight{Values: map[string]*WeightValue{}, path: path}
	values := v.GetStringMap(path + ".values")
	if values == nil {
		return nil, errors.Wrap(ErrConfigParseError, "format error "+path+".type")
	}
	for value := range values {
		newPath := path + ".values." + strings.ToLower(value)
		next, err := weight.CreateChild(config, newPath, v)
		if err != nil {
			return nil, err
		}
		weight.Values[strings.ToUpper(value)] = next
	}

	return weight, nil
}

func (we *Weight) CreateChild(config *Config, path string, v *viper.Viper) (*WeightValue, error) {
	var err error
	weight := &WeightValue{path: path}
	weight.Weight = uint8(v.GetInt(path + ".weight"))
	if weight.Weight == 0 {
		return nil, errors.Wrap(ErrConfigParseError, "weight is zero or empty "+path+".weight")
	}
	if weight.Next, err = CreateService(config, path+".next", v); err != nil {
		return nil, err
	}
	return weight, nil
}

func (we *Weight) Path() string {
	return we.path
}

func (we *Weight) GetRR(w dns.ResponseWriter, req *dns.Msg) ([]dns.RR, error) {
	var sumWeight uint16
	var err error
	var next_rr []dns.RR

	rrs := map[string][]dns.RR{}

	for name, value := range we.Values {
		next_rr, err = value.Next.GetRR(w, req)
		if err == nil {
			sumWeight += uint16(value.Weight)
			rrs[name] = next_rr
		}
	}
	if sumWeight == 0 {
		return []dns.RR{}, ErrServiceStatusError
	}

	var name string
	mod := req.MsgHdr.Id % sumWeight
	for name, next_rr = range rrs {
		if mod <= uint16(we.Values[name].Weight) {
			break
		}
		mod -= uint16(we.Values[name].Weight)
	}

	return next_rr, nil
}

func init() {
	AddServicePlugin("weight", NewWeight)
}
