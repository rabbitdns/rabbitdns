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

package config

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

var (
	ErrReadConfig             = errors.New("can't read config_file")
	ErrParseConfig            = errors.New("can't parse config file")
	ErrSyntaxNoListen         = errors.New("Listens parameter is required")
	ErrSyntaxInvalidListen    = errors.New("Listens parameter is invalid format")
	ErrSyntaxUnknownUser      = errors.New("User parameter is Unknown")
	ErrSyntaxNoCtlListen      = errors.New("CtlListens parameter is required")
	ErrSyntaxCtlInvalidListen = errors.New("CtlListens parameter is invalid format")
	ErrSyntaxMinTCPQueries    = errors.New("MaxTCPQueries parameter must grater than 0")
)

type Config struct {
	Listens             []string
	User                string
	CtlListens          []string
	LogLevel            string
	MaxTCPQueries       int
	ZonesDir            string
	ServicesDir         string
	MonitorsDir         string
	StateFile           string
	MinimumResponse     bool
	AutoZoneReload      bool
	AutoServiceReconfig bool
	AutoMonitorReconfig bool
}

func SetLogLevel(logLevel string) {
	switch logLevel {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	}
}

func LoadConfig(configFile string, ch chan *Config) {
	hupCh := make(chan os.Signal, 1)
	signal.Notify(hupCh, syscall.SIGHUP)
	first := true
	for {
		var err error
		c := &Config{}

		v := viper.New()
		v.SetDefault("Listens", []string{"0.0.0.0:53", "[::]:53"})
		v.SetDefault("User", "rabbitdns")
		v.SetDefault("CtlListens", []string{"127.0.0.1:8053", "[::1]:8053"})
		v.SetDefault("LogLevel", "info")
		v.SetDefault("MaxTCPQueries", 1000)
		v.SetDefault("ZonesDir", "zones")
		v.SetDefault("ServicesDir", "services")
		v.SetDefault("MonitorsDir", "monitors")
		v.SetDefault("StateFile", "/tmp/rabbitdns-state.dat")
		v.SetDefault("MinimumResponse", false)
		v.SetDefault("AutoZoneReload", true)
		v.SetDefault("AutoServiceReconfig", true)
		v.SetDefault("AutoMonitorReconfig", true)

		v.SetConfigType("toml")
		v.SetConfigName("config")
		v.AddConfigPath(filepath.Dir(configFile))
		v.AddConfigPath(".")
		v.AddConfigPath("../../test")

		if err = v.ReadInConfig(); err != nil {
			goto ERROR
		}
		if err = v.UnmarshalExact(c); err != nil {
			goto ERROR
		}
		if err = c.Check(); err != nil {
			goto ERROR
		}
		ch <- c
		goto NEXT
	ERROR:
		if first {
			log.WithFields(log.Fields{
				"Type":  "lib/config/Config",
				"Func":  "LoadConfig",
				"Error": err,
				"file":  configFile,
			}).Fatal(ErrReadConfig)
		} else {
			log.WithFields(log.Fields{
				"Type":  "lib/config/Config",
				"Func":  "LoadConfig",
				"Error": err,
				"file":  configFile,
			}).Warn(ErrReadConfig)

		}
	NEXT:
		<-hupCh
	}
}

func (c *Config) Check() error {
	syntaxError := &SyntaxError{}

	if len(c.Listens) == 0 {
		syntaxError.Add(ErrSyntaxNoListen)
	}
	for _, listen := range c.Listens {
		_, err := net.ResolveTCPAddr("tcp", listen)
		if err != nil {
			syntaxError.Add(ErrSyntaxInvalidListen)
		}
	}
	_, err := user.Lookup(c.User)
	if err != nil {
		syntaxError.Add(ErrSyntaxUnknownUser)
	}
	if len(c.CtlListens) == 0 {
		syntaxError.Add(ErrSyntaxNoCtlListen)
	}
	for _, listen := range c.CtlListens {
		_, err := net.ResolveTCPAddr("tcp", listen)
		if err != nil {
			syntaxError.Add(ErrSyntaxCtlInvalidListen)
		}
	}
	if c.MaxTCPQueries == 0 {
		syntaxError.Add(ErrSyntaxMinTCPQueries)
	}
	return syntaxError.Return()
}
