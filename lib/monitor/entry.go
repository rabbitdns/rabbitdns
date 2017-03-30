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
	"time"

	log "github.com/sirupsen/logrus"
)

type Entry struct {
	status   bool
	Value    string
	RRtype   uint16
	config   *Config
	statusCh chan bool

	CancelFunc context.CancelFunc

	ngCounter   uint16
	okCounter   uint16
	serviceName string
	path        string
}

func NewEntry(service_name, path, value string, rrtype uint16, mon *Config, statusCh chan bool) *Entry {
	return &Entry{serviceName: service_name, path: path, status: true, Value: value, RRtype: rrtype, config: mon, statusCh: statusCh}
}

func (e *Entry) Status() bool {
	return e.status
}
func (e *Entry) SetStatus(status bool) {
	e.status = status
	e.statusCh <- status
	e.ngCounter = 0
	e.okCounter = 0
	log.WithFields(log.Fields{
		"Type":   "lib/monitor/entry",
		"Func":   "SetStatus",
		"status": status,
	}).Debug("Set Status")

}
func (e *Entry) AddStatus(lastStatus bool) {
	if e.status {
		if lastStatus {
			e.okCounter++
			if e.okCounter >= e.config.OKThreshold {
				e.ngCounter = 0
				e.okCounter = 0
			}
		} else {
			e.okCounter = 0
			e.ngCounter++
			if e.ngCounter >= e.config.NGThreshold {
				e.ngCounter = 0
				e.status = false
				e.statusCh <- false
				log.WithFields(log.Fields{
					"Type":         "lib/monitor/entry",
					"Func":         "AddStatus",
					"service_name": e.serviceName,
					"service_path": e.path,
				}).Info("Status Down")

			}
		}
	} else {
		if lastStatus {
			e.okCounter++
			if e.okCounter >= e.config.UPThreshold {
				e.okCounter = 0
				e.ngCounter = 0
				e.status = true
				e.statusCh <- true
				log.Info("Status Up")
			}
		} else {
			e.okCounter = 0
		}
	}
}
func (e *Entry) Monitor(ctx context.Context) bool {
	return e.config.Monitor.Run(ctx, e)
}
func (e *Entry) MonitorRun(ctx context.Context) {
	resultCh := make(chan bool, 1)
	timeout, canselFunc := context.WithTimeout(ctx, e.config.Timeout*time.Second)
	go func() {
		resultCh <- e.Monitor(timeout)
	}()
	select {
	case <-timeout.Done():
		e.AddStatus(false)
	case <-ctx.Done():
		canselFunc()
	case result := <-resultCh:
		e.AddStatus(result)
		log.WithFields(log.Fields{
			"Type":         "lib/monitor/entry",
			"Func":         "MonitorRun",
			"service_name": e.serviceName,
			"service_path": e.path,
			"status":       e.status,
			"result":       result,
			"okCounter":    e.okCounter,
			"ngCounter":    e.ngCounter,
		}).Debug("recv result")
	}
	return
}
func (e *Entry) Start(ctx context.Context) {
	var loopCtx context.Context
	loopCtx, e.CancelFunc = context.WithCancel(ctx)
	ticker := time.NewTicker(e.config.Interval * time.Second)
	for {
		select {
		case <-loopCtx.Done():
			return
		case <-ticker.C:
			e.MonitorRun(loopCtx)
		}
	}
}
func (e *Entry) Stop() {
	e.CancelFunc()
}
