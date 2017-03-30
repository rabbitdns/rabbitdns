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

	"github.com/pkg/errors"

	"github.com/miekg/dns"
	"github.com/spf13/viper"
)

var (
	ErrConfigParseError   = errors.New("Failed to parse service config.")
	ErrServiceNotFound    = errors.New("Service not found")
	ErrServiceExist       = errors.New("Service is already exist. ")
	ErrServiceStatusError = errors.New("Monitoring status is failed")
)

var ServicePlugin = map[string]func(*Config, string, *viper.Viper) (Service, error){}

type Service interface {
	GetRR(dns.ResponseWriter, *dns.Msg) ([]dns.RR, error)
	Path() string
}

type ServiceGenerator struct {
	ServiceType int
	NewFunc     func() Service
}

func AddServicePlugin(name string, newFunc func(config *Config, path string, v *viper.Viper) (Service, error)) error {
	name = strings.ToUpper(name)
	if _, exist := ServicePlugin[name]; exist {
		return ErrServiceExist
	}
	ServicePlugin[name] = newFunc
	return nil
}

func CreateService(config *Config, path string, v *viper.Viper) (Service, error) {
	service_name := v.GetString(path + ".type")
	if service_name == "" {
		return nil, errors.Wrap(ErrConfigParseError, "format error "+path+".type")
	}
	service_name = strings.ToUpper(service_name)

	newFunc, ok := ServicePlugin[service_name]
	if ok == false {
		return nil, errors.Wrap(ErrServiceNotFound, "path="+path+",service_name="+service_name)
	}

	s, err := newFunc(config, path, v)
	if err != nil {
		return nil, err
	}
	return s, nil
}
