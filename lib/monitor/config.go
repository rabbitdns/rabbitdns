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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/viper"
)

var (
	ErrReadMonitor       = errors.New("Failed to read monitor config.")
	ErrEmptyMonitor      = errors.New("Empty monitor hash.")
	ErrZeroInterval      = errors.New("Intaval value is zero.")
	ErrZeroUPThreshold   = errors.New("UPThreshold value is zero.")
	ErrZeroOKThreshold   = errors.New("OKThreshold value is zero.")
	ErrZeroNGThreshold   = errors.New("NGThreshold value is zero.")
	ErrTimeoutGTInterval = errors.New("Timeout is grater than Interval.")
)

type Config struct {
	Interval    time.Duration
	Timeout     time.Duration
	UPThreshold uint16
	OKThreshold uint16
	NGThreshold uint16
	Monitor     Monitor
	ModTime     time.Time
}

func LoadConfig(path string) (*Config, error) {
	var err error
	config := &Config{}

	v := viper.New()
	v.SetConfigType("yml")
	v.SetDefault("Interval", 10)
	v.SetDefault("UPThreshold", 20)
	v.SetDefault("OKThreshold", 10)
	v.SetDefault("NGThreshold", 10)
	v.SetConfigName(strings.TrimSuffix(filepath.Base(path), ".yml"))
	v.AddConfigPath(filepath.Dir(path))

	stat, err := os.Stat(path)
	if err != nil {
		return nil, ErrReadMonitor
	}
	config.ModTime = stat.ModTime()
	if err != nil {
		return nil, ErrReadMonitor
	}
	if err = v.ReadInConfig(); err != nil {
		return nil, errors.Wrap(ErrReadMonitor, err.Error())
	}
	config.Interval = time.Duration(v.GetInt("Interval"))
	config.Timeout = time.Duration(v.GetInt("Timeout"))
	config.UPThreshold = uint16(v.GetInt("UPThreshold"))
	config.OKThreshold = uint16(v.GetInt("OKThreshold"))
	config.NGThreshold = uint16(v.GetInt("NGThreshold"))
	if config.Interval == 0 {
		return nil, ErrZeroInterval
	}
	if config.Timeout == 0 {
		config.Timeout = config.Interval / 2
		if config.Timeout == 0 {
			config.Timeout = 1
		}
	}
	if config.Interval < config.Timeout {
		return nil, ErrTimeoutGTInterval
	}
	if config.UPThreshold == 0 {
		return nil, ErrZeroInterval
	}
	if config.OKThreshold == 0 {
		return nil, ErrZeroInterval
	}
	if config.NGThreshold == 0 {
		return nil, ErrZeroInterval
	}
	// RRType
	name := v.Get("monitor")
	if name == nil {
		return nil, ErrEmptyMonitor
	}
	config.Monitor, err = CreateMonitor("monitor", v)
	return config, err
}

func (c *Config) CheckRegister(e *Entry) error {
	return c.Monitor.CheckRegister(e)
}

func (c *Config) Update(current *Config) error {
	return nil
}
