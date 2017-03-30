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
	"fmt"
	"os"
	"runtime"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/rabbitdns/rabbitdns/api"
	"github.com/rabbitdns/rabbitdns/lib/misc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const maxUint16 = 1<<16 - 1

var ErrExecute = errors.New("execute error")

var rootCmd = &cobra.Command{
	Use: "rabbitdns-controller",
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	runtime.GOMAXPROCS(runtime.NumCPU())

	rootCmd.PersistentFlags().StringP("ip", "i", "127.0.0.1", "connect ip")
	rootCmd.PersistentFlags().StringP("port", "p", "15300", "connect port")
}

func main() {
	cmdVersion := &cobra.Command{
		Use:   "version",
		Short: "show version",
		Long:  `show version.`,
		Args:  cobra.MaximumNArgs(0),
		Run:   version,
	}

	cmdReconfig := &cobra.Command{
		Use:   "reconfig",
		Short: "reconfig config",
		Long:  `reconfig rabbitdns config, service config, monitor config and reload zonefile.`,
		Args:  cobra.MaximumNArgs(0),
		Run:   reconfig,
	}
	cmdReload := &cobra.Command{
		Use:   "reload",
		Short: "reload reload",
		Long:  `reload service config, monitor config and reload zonefile.`,
		Args:  cobra.MaximumNArgs(1),
		Run:   reload,
	}

	cmdZones := &cobra.Command{
		Use:   "zones",
		Short: "Print zone names",
		Long:  `echo is for echoing anything back.`,
		Args:  cobra.MaximumNArgs(0),
		Run:   getZones,
	}
	cmdServices := &cobra.Command{
		Use:   "services",
		Short: "Print zone names",
		Long:  `echo is for echoing anything back.`,
		Args:  cobra.MaximumNArgs(0),
		Run:   getServices,
	}
	cmdMonitors := &cobra.Command{
		Use:   "monitors",
		Short: "Print zone names",
		Long:  `echo is for echoing anything back.`,
		Args:  cobra.MaximumNArgs(0),
		Run:   getMonitors,
	}

	rootCmd.AddCommand(cmdZones, cmdServices, cmdMonitors, cmdReconfig, cmdReload, cmdVersion)

	if err := rootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{
			"Type":  "rabbitdns-server",
			"Func":  "main",
			"Error": err,
		}).Fatal(ErrExecute)
	}
	os.Exit(0)
}

func version(cb *cobra.Command, args []string) {
	fmt.Printf("rabbitdns version %s\n", misc.RabbitdnsVersion)
}

func connect(cb *cobra.Command) api.RabbitDNSClient {
	host, err := rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		log.WithFields(log.Fields{
			"Type": "rabbitdns-controller",
			"Func": "connect",
			"err":  err,
		}).Fatal(err)
	}
	port, err := rootCmd.PersistentFlags().GetString("port")
	if err != nil {
		log.WithFields(log.Fields{
			"Type": "rabbitdns-controller",
			"Func": "connect",
			"err":  err,
		}).Fatal(err)
	}

	connectTo := host + ":" + port

	conn, err := grpc.Dial(connectTo, grpc.WithInsecure())
	if err != nil {
		log.WithFields(log.Fields{
			"Type":      "rabbitdns-controller",
			"Func":      "connect",
			"connectTo": connectTo,
			"err":       err,
		}).Fatal("client connection error")
	}

	return api.NewRabbitDNSClient(conn)
}

func reconfig(cb *cobra.Command, args []string) {
	client := connect(cb)
	message := &empty.Empty{}
	client.Reconfig(context.TODO(), message)
}

func reload(cb *cobra.Command, args []string) {
	client := connect(cb)
	if len(args) == 0 {
		message := &empty.Empty{}
		client.Reload(context.TODO(), message)
	} else {
		message := &api.ReloadRequest{Zonename: args[0]}
		client.ReloadZone(context.TODO(), message)
	}
}

func getZones(cb *cobra.Command, args []string) {
	client := connect(cb)
	message := &empty.Empty{}
	res, err := client.GetZones(context.TODO(), message)
	if err != nil {
		fmt.Printf("error::%#v \n", err)
	}
	if res != nil {
		for _, zone := range res.Zones {
			fmt.Printf("%s\n", zone.Name)
		}
	}
}
func getServices(cb *cobra.Command, args []string) {
	client := connect(cb)
	message := &empty.Empty{}
	res, err := client.GetServices(context.TODO(), message)
	if err != nil {
		fmt.Printf("error::%#v \n", err)
	}
	if res != nil {
		for _, service := range res.Services {
			fmt.Printf("%s\n", service.Name)
		}
	}
}
func getMonitors(cb *cobra.Command, args []string) {
	client := connect(cb)
	message := &empty.Empty{}
	res, err := client.GetMonitors(context.TODO(), message)
	if err != nil {
		fmt.Printf("error::%#v \n", err)
	}
	if res != nil {
		for _, monitor := range res.Monitors {
			fmt.Printf("%s\n", monitor.Name)
		}
	}
}
