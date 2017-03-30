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

package monitor

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/spf13/viper"
)

var (
	ErrMonitorNotFound  = errors.New("Provide not supported monitor type.")
	ErrConfigParseError = errors.New("failed to parse monitor config.")
)

type Monitor interface {
	CheckRegister(*Entry) error
	Run(context.Context, *Entry) bool
	Path() string
}

var MonitorPlugins = map[string]func(string, *viper.Viper) (Monitor, error){}
var MonitorStatus = map[string]int{}

func AddMonitorPlugin(name string, newFunc func(path string, v *viper.Viper) (Monitor, error)) {
	name = strings.ToUpper(name)
	MonitorPlugins[name] = newFunc
}

func CreateMonitor(path string, v *viper.Viper) (Monitor, error) {
	monitor_type := v.GetString(path + ".type")
	if monitor_type == "" {
		return nil, errors.Wrap(ErrConfigParseError, "format error "+path+".type")
	}
	monitor_type = strings.ToUpper(monitor_type)

	newFunc, ok := MonitorPlugins[monitor_type]
	if ok == false {
		return nil, errors.Wrap(ErrMonitorNotFound, "path="+path+",service_name="+monitor_type)
	}

	s, err := newFunc(path, v)

	if err != nil {
		return nil, err
	}
	return s, nil

}
