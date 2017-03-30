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
	"net"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	ErrEmptyPort            = errors.New("Empty port number.")
	ErrTcpconNotSupportedRR = errors.New("Tcpcon only support A,AAAA and CNAME.")
)

type Tcpcon struct {
	path string
	port string
}

func NewTcpcon(path string, v *viper.Viper) (Monitor, error) {
	port := v.GetString(path + ".port")
	if port == "" {
		return nil, ErrEmptyPort
	}
	return &Tcpcon{path: path, port: port}, nil
}

func (t *Tcpcon) Path() string {
	return t.path
}

func (t *Tcpcon) CheckRegister(e *Entry) error {
	switch e.RRtype {
	case dns.TypeA, dns.TypeAAAA, dns.TypeCNAME:
	default:
		return ErrTcpconNotSupportedRR
	}
	return nil
}

func (t *Tcpcon) Run(ctx context.Context, e *Entry) bool {
	var connectTo string
	switch e.RRtype {
	case dns.TypeAAAA:
		connectTo = "[" + e.Value + "]:" + t.port
	case dns.TypeA, dns.TypeCNAME:
		connectTo = e.Value + ":" + t.port
	}
	conn, err := net.Dial("tcp", connectTo)
	if err != nil {
		return false
	}
	conn.Close()

	return true
}

func init() {
	AddMonitorPlugin("tcpcon", NewTcpcon)
}
