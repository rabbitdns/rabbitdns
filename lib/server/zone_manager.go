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
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/miekg/dns"
	"github.com/rabbitdns/rabbitdns/lib/config"
	. "github.com/rabbitdns/rabbitdns/lib/misc"
	log "github.com/sirupsen/logrus"
)

var (
	ErrGlobZone        = errors.New("failed to glob zone files.")
	ErrParseRR         = errors.New("failed to parse RR.")
	ErrParseZone       = errors.New("failed to parse zone file.")
	ErrServiceNotFount = errors.New("Service is not found.")
)

type zoneManager struct {
	config         *config.Config
	zoneSet        *Tree
	loading        map[string]bool
	serviceManager *serviceManager
}

func NewZoneManager(c *config.Config, s *serviceManager) *zoneManager {
	return &zoneManager{
		config:         c,
		zoneSet:        NewTree(),
		loading:        map[string]bool{},
		serviceManager: s,
	}
}
func (m *zoneManager) GetZones() []map[string]string {
	results := []map[string]string{}
	for file, _ := range m.loading {
		origin := filepath.Base(file)
		result := map[string]string{
			"name": origin,
		}
		results = append(results, result)
	}
	return results
}

func (m *zoneManager) ReadZone(zoneFile string) error {
	file, err := os.Open(zoneFile)
	if err != nil {
		return err
	}
	stat, err := os.Stat(zoneFile)
	if err != nil {
		return err
	}
	origin := filepath.Base(zoneFile)
	origin = FQDN(origin)
	origin_labels := Labels(origin)

	zoneNode := m.zoneSet.AddNode(origin_labels)
	zoneNode.Set("provide", true)
	if v, ok := zoneNode.Get("ModTime"); ok == true {
		switch modTime := v.(type) {
		case time.Time:
			if modTime.Equal(stat.ModTime()) {
				return nil
			}
		}
	}

	services := []string{}
	zoneTree := NewTree()
	zoneTree.Auth = true
	var RRs []dns.RR
	success := true
	for x := range dns.ParseZone(file, origin, "") {
		if x.Error != nil {
			log.WithFields(log.Fields{
				"Type":  "lib/server/zoneManager",
				"Func":  "LoadZones",
				"Error": x.Error,
			}).Warn(ErrParseRR)
			success = false

			return ErrParseZone
		} else {
			if dyn, ok := x.RR.(*dns.PrivateRR); ok {
				if rdata, ok := dyn.Data.(*DYNRR); ok {
					if _, ok := m.serviceManager.GetService(rdata.Resource); ok == false {
						return errors.Wrap(ErrServiceNotFount, "name:"+dyn.Header().Name+",ServiceName:"+rdata.Resource)
					}
					services = append(services, rdata.Resource)
				}
			}
			node := zoneTree.AddRR(x.RR)
			if x.RR.Header().Rrtype == dns.TypeNS && FQDN(x.RR.Header().Name) != origin {
				node.Auth = false
			}
		}
	}
	if err := zoneTree.VerifyZone(origin_labels); err != nil {
		zoneNode.Set("state", LOAD_ERROR)
		return err

	}
	if success != true {
		zoneNode.Set("state", LOAD_ERROR)
		return ErrParseZone
	}
	zoneNode.Set("ModTime", stat.ModTime())
	zoneNode.Set("Records", RRs)
	zoneNode.Set("state", OK)
	zoneNode.Set("ZoneTree", zoneTree)
	zoneNode.Set("services", services)
	for _, service_name := range services {
		m.serviceManager.RegisterService(service_name, origin)
	}

	log.WithFields(log.Fields{
		"Type":     "lib/server/zoneManager",
		"Func":     "readZone",
		"zonename": origin,
	}).Info("load zone")

	return nil
}

func (m *zoneManager) LoadZones() error {
	for k, _ := range m.loading {
		m.loading[k] = false
	}
	matches, err := filepath.Glob(m.config.ZonesDir + "/*")
	if err != nil {
		log.WithFields(log.Fields{
			"Type":  "lib/server/zoneManager",
			"Func":  "LoadZones",
			"Error": err,
		}).Warn(ErrGlobZone)
		return ErrGlobZone
	}
	for _, f := range matches {
		err := m.ReadZone(f)
		m.loading[f] = true
		if err != nil {
			log.WithFields(log.Fields{
				"Type":     "lib/server/zoneManager",
				"Func":     "LoadZones",
				"Error":    err,
				"filename": f,
			}).Warn(err)
			return err
		}
	}
	return nil
}
func (m *zoneManager) DeleteZones() {
	for k, v := range m.loading {
		if v == false {
			origin := filepath.Base(k)
			origin = FQDN(origin)
			origin_labels := Labels(origin)

			if node := m.zoneSet.SearchNode(origin_labels, true); node != nil {
				if s, ok := node.Get("services"); ok {
					if services, ok := s.([]string); ok {
						for _, service_name := range services {
							m.serviceManager.UnRegisterService(service_name, origin)
						}
					}
				}
				node.DeleteAll()
			}

			m.zoneSet.DeleteNode(origin_labels, false)
		}
	}
}
