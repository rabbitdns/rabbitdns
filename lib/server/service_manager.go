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
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/miekg/dns"
	"github.com/rabbitdns/rabbitdns/lib/config"
	"github.com/rabbitdns/rabbitdns/lib/service"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNotDefineService      = errors.New("Not define service.")
	ErrMismatchServiceRRtype = errors.New("rrtype mismatch.")
	ErrReadService           = errors.New("failed to read service config.")
	ErrUsingService          = errors.New("faild to delete service. because zone use this service.")
)

type serviceManager struct {
	config            *config.Config
	services          map[string]*service.Config
	loading           map[string]bool
	using             map[string]map[string]bool
	monitoringManager *monitoringManager
}

func NewServiceManager(c *config.Config, m *monitoringManager) *serviceManager {
	return &serviceManager{
		config:            c,
		services:          map[string]*service.Config{},
		loading:           map[string]bool{},
		using:             make(map[string]map[string]bool),
		monitoringManager: m,
	}
}
func (s *serviceManager) GetServices() []map[string]string {
	results := []map[string]string{}
	for name, _ := range s.services {
		result := map[string]string{
			"name": name,
		}
		results = append(results, result)
	}
	return results
}

func (s *serviceManager) GetResources(w dns.ResponseWriter, req *dns.Msg, rrType uint16, name string) ([]dns.RR, error) {
	service, exist := s.services[name]
	if !exist {
		return []dns.RR{}, ErrNotDefineService
	}
	if service.RRType != rrType {
		return []dns.RR{}, ErrMismatchServiceRRtype
	}
	rrs, err := service.Service.GetRR(w, req)
	return rrs, err
}
func (s *serviceManager) GetService(name string) (*service.Config, bool) {
	service, ok := s.services[name]
	return service, ok
}
func (s *serviceManager) RegisterService(service string, zonename string) {
	s.using[service][zonename] = true
}
func (s *serviceManager) UnRegisterService(service string, zonename string) {
	delete(s.using[service], zonename)
}
func (s *serviceManager) addService(confPath string) error {
	name := strings.TrimSuffix(filepath.Base(confPath), ".yml")

	if current, exist := s.services[name]; exist {
		if stat, err := os.Stat(confPath); err == nil {
			if current.ModTime.Equal(stat.ModTime()) {
				return nil
			}
		}
	}

	service, err := service.LoadConfig(confPath)
	if err != nil {
		return err
	}

	for endpoint, monitor := range service.Monitors {
		log.WithFields(log.Fields{
			"Type":    "lib/server/serviceManager",
			"Func":    "addService",
			"monitor": monitor,
			"name":    name,
			"value":   endpoint.Value,
			"RRType":  endpoint.RRType,
		}).Debug("CheckRegisterMonitor")

		err = s.monitoringManager.CheckRegisterMonitor(monitor, name, endpoint.Value, endpoint.RRType)
		if err != nil {
			return err
		}
	}
	for endpoint, monitor := range service.Monitors {
		log.WithFields(log.Fields{
			"Type":    "lib/server/serviceManager",
			"Func":    "addService",
			"monitor": monitor,
			"name":    name,
			"value":   endpoint.Value,
			"RRType":  endpoint.RRType,
		}).Debug("RegisterMonitor")

		s.monitoringManager.RegisterMonitor(monitor, name, endpoint.Path(), endpoint.Value, endpoint.RRType, endpoint.StatusCh)
		go endpoint.StartStatusWatch(context.Background())
	}

	s.services[name] = service
	s.using[name] = make(map[string]bool)
	log.WithFields(log.Fields{
		"Type": "lib/server/serviceManager",
		"Func": "addService",
		"name": name,
	}).Debug("Done create service")

	return nil
}
func (s *serviceManager) LoadServices() error {
	for k := range s.loading {
		s.loading[k] = false
	}

	matches, err := filepath.Glob(s.config.ServicesDir + "/*.yml")
	if err != nil {
		log.WithFields(log.Fields{
			"Type":  "lib/server/serviceManager",
			"Func":  "LoadServices",
			"Error": err,
		}).Warn(ErrGlobZone)
		return ErrGlobZone
	}
	for _, f := range matches {
		err := s.addService(f)
		s.loading[f] = true
		if err != nil {
			log.WithFields(log.Fields{
				"Type":  "lib/server/serviceManager",
				"Func":  "LoadServices",
				"Error": err,
				"file":  f,
			}).Warn(ErrReadService)

			return ErrReadService
		}
	}
	return nil
}

func (s *serviceManager) DeleteServices() {
	for k, v := range s.loading {
		if v == false {
			name := strings.TrimSuffix(filepath.Base(k), ".yml")
			if len(s.using[name]) > 0 {
				log.WithFields(log.Fields{
					"Type":         "lib/server/serviceManager",
					"Func":         "LoadServices",
					"service_name": name,
				}).Warn(ErrUsingService)
				continue
			}
			for endpoint, monitor := range s.services[k].Monitors {
				s.monitoringManager.UnRegisterMonitor(monitor, name, endpoint.Path())
			}
			delete(s.using, name)
		}
	}
}
