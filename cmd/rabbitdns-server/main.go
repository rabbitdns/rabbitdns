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

package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/rabbitdns/rabbitdns/lib/config"
	"github.com/rabbitdns/rabbitdns/lib/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const maxUint16 = 1<<16 - 1

var ErrExecute = errors.New("execute error")

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	rootCmd := &cobra.Command{
		Use: "rabbitdns-server",
		Run: serv,
	}

	rootCmd.PersistentFlags().StringP("config_dir", "c", "config.toml", "config dir path")
	rootCmd.PersistentFlags().StringP("log_level", "l", "info", "log level")

	viper.BindPFlag("LogLevel", rootCmd.PersistentFlags().Lookup("level"))

	if err := rootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{
			"Type":  "rabbitdns-server",
			"Func":  "main",
			"Error": err,
		}).Fatal(ErrExecute)
	}
	os.Exit(0)
}
func serv(cb *cobra.Command, args []string) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)

	master := server.NewMaster()
	var configPath, logLevel string
	configPath, _ = cb.PersistentFlags().GetString("config_dir")
	logLevel, _ = cb.PersistentFlags().GetString("log_level")

	config.SetLogLevel(logLevel)
	// syslog or stdout or file
	var c *config.Config
	configCh := make(chan *config.Config)

	go config.LoadConfig(configPath, configCh)
	log.WithFields(log.Fields{
		"Type": "rabbitdns-server",
		"Func": "serv",
	}).Info("rabbitdns started")

	for {
		select {
		case <-sigCh:
			return
		case newConfig := <-configCh:
			if c == nil {
				c = newConfig
				config.SetLogLevel(c.LogLevel)
				if err := master.StartServ(context.Background(), c); err != nil {
					log.Fatal(err)
				}
			} else {
				// reload config
				config.SetLogLevel(newConfig.LogLevel)

			}
		}
	}
}
