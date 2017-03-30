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
	ErrSamePriority = errors.New("Can't set same priority.")
)

type Failover struct {
	Values []*FailoverValue
	Weight uint32
	path   string
}

type FailoverValue struct {
	Priority uint8
	Next     Service
	path     string
}

func NewFailover(config *Config, path string, v *viper.Viper) (Service, error) {
	failover := &Failover{Values: []*FailoverValue{}, path: path}
	values := v.GetStringMap(path + ".values")
	if values == nil {
		return nil, errors.Wrap(ErrConfigParseError, "format error "+path+".type")
	}
	for value := range values {
		newPath := path + ".values." + strings.ToLower(value)
		fv, err := failover.CreateChild(config, newPath, v)
		if err != nil {
			return nil, err
		}
		if err := failover.insertChild(fv); err != nil {
			return nil, err
		}
	}

	return failover, nil
}

func (f *Failover) insertChild(fv *FailoverValue) error {
	for i, v := range f.Values {
		if v.Priority > fv.Priority {
			f.Values = append(f.Values, fv)
			copy(f.Values[i+1:], f.Values[i:])
			f.Values[i] = fv
			return nil
		}
		if v.Priority == fv.Priority {
			return ErrSamePriority
		}
	}
	f.Values = append(f.Values, fv)
	return nil
}
func (f *Failover) CreateChild(config *Config, path string, v *viper.Viper) (*FailoverValue, error) {
	var err error
	fv := &FailoverValue{path: path}
	fv.Priority = uint8(v.GetInt(path + ".priority"))
	if fv.Priority == 0 {
		return nil, errors.Wrap(ErrConfigParseError, "Priority is zero or empty "+path+".weight (1 is highest)")
	}
	if fv.Next, err = CreateService(config, path+".next", v); err != nil {
		return nil, err
	}
	return fv, nil
}

func (f *Failover) Path() string {
	return f.path
}

func (f *Failover) GetRR(w dns.ResponseWriter, req *dns.Msg) (rrs []dns.RR, err error) {
	for _, value := range f.Values {
		rrs, err = value.Next.GetRR(w, req)
		if err == nil {
			return rrs, nil
		}
	}
	return []dns.RR{}, err
}

func init() {
	AddServicePlugin("failover", NewFailover)
}
