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

	"github.com/rabbitdns/rabbitdns/lib/config"
	"github.com/rabbitdns/rabbitdns/lib/monitor"

	log "github.com/sirupsen/logrus"
)

var (
	ErrReadMonitor      = errors.New("failed to read monitor config.")
	ErrNotDefineMonitor = errors.New("Not define monitor")
	ErrEmptyState       = errors.New("empty state")
)

type monitoringManager struct {
	config   *config.Config
	loading  map[string]bool
	monitors map[string]*monitor.Config
	entries  map[string]map[string]map[string]*monitor.Entry
	states   map[string]map[string]map[string]bool
}

func NewMonitoringManager(c *config.Config) *monitoringManager {
	return &monitoringManager{
		config:   c,
		loading:  map[string]bool{},
		monitors: map[string]*monitor.Config{},
		entries:  map[string]map[string]map[string]*monitor.Entry{},
		states:   map[string]map[string]map[string]bool{},
	}
}

func (m *monitoringManager) GetMonitors() []map[string]string {
	results := []map[string]string{}
	for name, _ := range m.monitors {
		result := map[string]string{
			"name": name,
		}
		results = append(results, result)
	}
	return results
}

func (m *monitoringManager) CheckRegisterMonitor(monitor_name, service, value string, rrtype uint16) error {
	mon, ok := m.monitors[monitor_name]
	if ok == false {
		return ErrNotDefineMonitor
	}
	entry := &monitor.Entry{Value: value, RRtype: rrtype}

	err := mon.CheckRegister(entry)
	if err != nil {
		return err
	}
	return nil
}
func (m *monitoringManager) RegisterMonitor(monitor_name, service, path, value string, rrtype uint16, statusCh chan bool) error {
	mon, ok := m.monitors[monitor_name]
	if ok == false {
		return ErrNotDefineMonitor
	}
	if m.entries[monitor_name][service] == nil {
		m.entries[monitor_name][service] = make(map[string]*monitor.Entry)
	}

	entry := monitor.NewEntry(service, path, value, rrtype, mon, statusCh)
	var status bool
	var err error
	if current, ok := m.entries[monitor_name][service][path]; ok {
		status = current.Status()
		current.CancelFunc()
		log.WithFields(log.Fields{
			"Type":          "lib/server/monitoringManager",
			"Func":          "RegisterMonitor",
			"monitor_name":  monitor_name,
			"service_name":  service,
			"endpoint_path": path,
			"status":        status,
		}).Debug("stop current monitor")

	} else {
		status, err = m.getState(monitor_name, service, path)
		if err != nil {
			status = entry.Monitor(context.Background())
		}
	}
	entry.SetStatus(status)
	m.entries[monitor_name][service][path] = entry
	log.WithFields(log.Fields{
		"Type":            "lib/server/monitoringManager",
		"Func":            "RegisterMonitor",
		"monitor_name":    monitor_name,
		"service_name":    service,
		"endpoint_path":   path,
		"endpoint_value":  value,
		"endpoint_rrtype": rrtype,
	}).Debug("start monitoring")

	go entry.Start(context.Background())
	return nil
}
func (m *monitoringManager) UnRegisterMonitor(monitor_name, service, path string) {
	if m.entries[monitor_name][service] != nil {
		if entry, exist := m.entries[monitor_name][service][path]; exist {
			entry.Stop()
			delete(m.entries[monitor_name][service], path)
		}
		if len(m.entries[monitor_name][service]) == 0 {
			delete(m.entries[monitor_name], service)
		}
	}
}

func (m *monitoringManager) addMonitor(confPath string) error {
	name := strings.TrimSuffix(filepath.Base(confPath), ".yml")

	if current, exist := m.monitors[name]; exist {
		if stat, err := os.Stat(confPath); err == nil {
			if current.ModTime.Equal(stat.ModTime()) {
				return nil
			}
		}
	}

	mon, err := monitor.LoadConfig(confPath)
	if err != nil {
		return err
	}
	m.monitors[name] = mon
	m.entries[name] = make(map[string]map[string]*monitor.Entry)

	return nil
}
func (m *monitoringManager) getState(monitor_name, service_name, path string) (bool, error) {
	if mmap, ok := m.states[monitor_name]; ok {
		if smap, ok := mmap[service_name]; ok {
			if state, ok := smap[path]; ok {
				return state, nil
			}
		}
	}
	return false, ErrEmptyState
}
func (m *monitoringManager) SaveStates() {
	for mon_name, _ := range m.entries {
		m.states[mon_name] = make(map[string]map[string]bool)
		for service_name, _ := range m.entries[mon_name] {
			m.states[mon_name][service_name] = make(map[string]bool)
			for path, entry := range m.entries[mon_name][service_name] {
				m.states[mon_name][service_name][path] = entry.Status()
			}
		}
	}
}
func (m *monitoringManager) LoadMonitors() error {
	for k := range m.loading {
		m.loading[k] = false
	}
	matches, err := filepath.Glob(m.config.MonitorsDir + "/*.yml")
	if err != nil {
		log.WithFields(log.Fields{
			"Type":  "lib/server/monitoringManager",
			"Func":  "LoadMonitor",
			"Error": err,
		}).Warn(ErrGlobZone)
		return ErrGlobZone
	}
	for _, f := range matches {
		err := m.addMonitor(f)
		m.loading[f] = true
		if err != nil {
			log.WithFields(log.Fields{
				"Type":  "lib/server/monitoringManager",
				"Func":  "LoadMonitor",
				"Error": err,
				"file":  f,
			}).Warn(ErrReadMonitor)

			return ErrReadMonitor
		}
	}
	return nil
}

func (m *monitoringManager) DeleteMonitors() {
	for k, v := range m.loading {
		if v == false {
			name := strings.TrimSuffix(filepath.Base(k), ".yml")
			if len(m.entries[name]) > 0 {
				log.WithFields(log.Fields{
					"Type":         "lib/server/serviceManager",
					"Func":         "LoadServices",
					"service_name": name,
				}).Warn(ErrUsingService)
				continue
			}
			delete(m.entries, name)
			delete(m.monitors, name)
		}
	}
}
