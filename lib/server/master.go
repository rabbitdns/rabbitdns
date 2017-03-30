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

package server

import (
	"context"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	api "github.com/rabbitdns/rabbitdns/api"
	. "github.com/rabbitdns/rabbitdns/lib/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	LOAD_ERROR int = iota
	OK
)

var (
	ErrReloadError = errors.New("reload error")
)

type Master struct {
	workers []*worker
	config  *Config

	monitoringManager *monitoringManager
	serviceManager    *serviceManager
	zoneManager       *zoneManager
	reloadCh          chan bool
	mutex             sync.Mutex
}

func NewMaster() *Master {
	m := Master{
		reloadCh: make(chan bool),
	}
	return &m
}

func (m *Master) StartServ(ctx context.Context, c *Config) error {
	m.config = c
	m.monitoringManager = NewMonitoringManager(c)
	m.serviceManager = NewServiceManager(c, m.monitoringManager)
	m.zoneManager = NewZoneManager(c, m.serviceManager)

	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "StartServ",
	}).Info("start to load monitoring config")

	if err := m.monitoringManager.LoadMonitors(); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "StartServ",
	}).Info("start to load service config")

	if err := m.serviceManager.LoadServices(); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "StartServ",
	}).Info("start to load zone data")

	if err := m.zoneManager.LoadZones(); err != nil {
		return err
	}
	protocols := []string{"tcp", "udp"}
	for _, addr := range m.config.Listens {
		for _, proto := range protocols {
			m.workers = append(m.workers, NewWorker(m.config, m.zoneManager, m.serviceManager, addr, proto))
		}
	}

	for _, worker := range m.workers {
		worker.Run()
	}

	go m.updateConfig(ctx)
	server := grpc.NewServer()
	api.RegisterRabbitDNSServer(server, m)
	for _, addr := range m.config.CtlListens {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"Type": "lib/server/Master",
			"Func": "StartServ",
			"addr": addr,
		}).Info("api server will start")

		go server.Serve(lis)
	}

	return nil
}
func (m *Master) updateConfig(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.reloadCh:
			m.mutex.Lock()
			m.monitoringManager.LoadMonitors()
			m.serviceManager.LoadServices()
			m.zoneManager.LoadZones()
			m.zoneManager.DeleteZones()
			m.serviceManager.DeleteServices()
			m.monitoringManager.DeleteMonitors()
			m.mutex.Unlock()

		case <-ticker.C:
			m.mutex.Lock()
			if m.config.AutoMonitorReconfig {
				m.monitoringManager.LoadMonitors()
			}
			if m.config.AutoServiceReconfig {
				m.serviceManager.LoadServices()
			}
			if m.config.AutoZoneReload {
				m.zoneManager.LoadZones()
				m.zoneManager.DeleteZones()
			}
			if m.config.AutoServiceReconfig {
				m.serviceManager.DeleteServices()
			}
			if m.config.AutoMonitorReconfig {
				m.monitoringManager.DeleteMonitors()
			}
			m.mutex.Unlock()
		}
	}
}

func (m *Master) Reconfig(context.Context, *empty.Empty) (*empty.Empty, error) {
	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "Reconfig",
	}).Info("Receive request to reconfig.")

	response := &empty.Empty{}
	syscall.Kill(os.Getpid(), syscall.SIGHUP)
	return response, nil
}
func (m *Master) Reload(context.Context, *empty.Empty) (*empty.Empty, error) {
	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "Reload",
	}).Info("Receive request to reload all zones.")

	response := &empty.Empty{}
	m.reloadCh <- true
	return response, nil
}
func (m *Master) ReloadZone(ctx context.Context, request *api.ReloadRequest) (*empty.Empty, error) {
	log.WithFields(log.Fields{
		"Type":     "lib/server/Master",
		"Func":     "Reload",
		"zonename": request.Zonename,
	}).Info("Receive request to reload a zone.")

	response := &empty.Empty{}
	filePath := m.config.ZonesDir + "/" + request.Zonename

	m.mutex.Lock()
	err := m.zoneManager.ReadZone(filePath)
	m.mutex.Unlock()

	if err != nil {
		return nil, ErrReloadError
	}
	return response, nil
}
func (m *Master) GetZones(context.Context, *empty.Empty) (*api.GetZonesResponse, error) {
	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "GetZones",
	}).Info("Receive request to get zones.")

	m.mutex.Lock()
	response := &api.GetZonesResponse{Zones: []*api.Zone{}}
	m.mutex.Unlock()

	for _, v := range m.zoneManager.GetZones() {
		zone := &api.Zone{
			Name: v["name"],
		}
		response.Zones = append(response.Zones, zone)
	}
	return response, nil
}
func (m *Master) GetMonitors(context.Context, *empty.Empty) (*api.GetMonitorsResponse, error) {
	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "GetMonitors",
	}).Info("Receive request to get monitors.")

	m.mutex.Lock()
	response := &api.GetMonitorsResponse{Monitors: []*api.Monitor{}}
	m.mutex.Unlock()

	for _, v := range m.monitoringManager.GetMonitors() {
		monitor := &api.Monitor{
			Name: v["name"],
		}
		response.Monitors = append(response.Monitors, monitor)
	}
	return response, nil

}
func (m *Master) GetServices(context.Context, *empty.Empty) (*api.GetServicesResponse, error) {
	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "GetServices",
	}).Info("Receive request to get services.")

	m.mutex.Lock()
	response := &api.GetServicesResponse{Services: []*api.Service{}}
	m.mutex.Unlock()

	for _, v := range m.serviceManager.GetServices() {
		service := &api.Service{
			Name: v["name"],
		}
		response.Services = append(response.Services, service)
	}
	return response, nil

}
