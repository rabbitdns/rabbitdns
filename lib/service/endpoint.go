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

package service

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"github.com/rabbitdns/rabbitdns/lib/config"
	. "github.com/rabbitdns/rabbitdns/lib/misc"
	"github.com/spf13/viper"
)

var (
	ErrEndpointValueEmpty       = errors.New("Value is empty")
	ErrEndpointValueNotString   = errors.New("Value is not string")
	ErrEndpointValueFormatError = errors.New("Value format error")
)

type Endpoint struct {
	Value    string
	RRType   uint16
	RR       dns.RR
	path     string
	status   bool
	StatusCh chan bool
}

func NewEndpoint(config *Config, path string, v *viper.Viper) (Service, error) {
	value := v.GetString(path + ".value")
	if value == "" {
		return nil, ErrEndpointValueEmpty
	}

	var rr dns.RR
	switch config.RRType {
	case dns.TypeA:
		ip := net.ParseIP(value)
		if ip == nil {
			return nil, ErrEndpointValueFormatError
		}
		if IpFamily(value) != 4 {
			return nil, ErrEndpointValueFormatError
		}
		rr = &dns.A{
			A: ip,
		}
	case dns.TypeAAAA:
		ip := net.ParseIP(value)
		if ip == nil {
			return nil, ErrEndpointValueFormatError
		}
		if IpFamily(value) != 6 {
			return nil, ErrEndpointValueFormatError
		}
		rr = &dns.AAAA{
			AAAA: ip.To16(),
		}
	case dns.TypeCNAME:
		if !IsDomainName(value) {
			return nil, ErrEndpointValueFormatError
		}
		rr = &dns.CNAME{
			Target: value,
		}
	case dns.TypeTXT:
		if len(value) > 254 {
			return nil, ErrEndpointValueFormatError
		}
		rr = &dns.TXT{
			Txt: []string{value},
		}
	case dns.TypeMX:
		values := strings.Split(value, " ")
		if len(values) != 2 {
			return nil, ErrEndpointValueFormatError
		}
		i, e := strconv.ParseUint(values[0], 10, 16)
		if e != nil {
			return nil, ErrEndpointValueFormatError
		}

		if !IsDomainName(values[1]) {
			return nil, ErrEndpointValueFormatError
		}

		rr = &dns.MX{
			Preference: uint16(i),
			Mx:         values[1],
		}

	case dns.TypePTR:
		if !IsDomainName(value) {
			return nil, ErrEndpointValueFormatError
		}
		rr = &dns.PTR{
			Ptr: value,
		}
	case dns.TypeSRV:
		if !IsDomainName(value) {
			return nil, ErrEndpointValueFormatError
		}
		values := strings.Split(value, " ")
		if len(values) != 4 {
			return nil, ErrEndpointValueFormatError
		}
		priority, e := strconv.ParseUint(values[0], 10, 16)
		if e != nil {
			return nil, ErrEndpointValueFormatError
		}
		weight, e := strconv.ParseUint(values[1], 10, 16)
		if e != nil {
			return nil, ErrEndpointValueFormatError
		}
		port, e := strconv.ParseUint(values[2], 10, 16)
		if e != nil {
			return nil, ErrEndpointValueFormatError
		}
		if !IsDomainName(values[3]) {
			return nil, ErrEndpointValueFormatError
		}

		rr = &dns.SRV{
			Priority: uint16(priority),
			Weight:   uint16(weight),
			Port:     uint16(port),
			Target:   values[3],
		}
	default:
		return nil, ErrEndpointValueFormatError
	}
	endpoint := &Endpoint{RRType: config.RRType, Value: value, RR: rr, path: path, status: true, StatusCh: make(chan bool, 10)}
	if monitor := v.GetString(path + ".monitor"); monitor != "" {
		config.Monitors[endpoint] = monitor
	}

	return endpoint, nil
}
func (e *Endpoint) Path() string {
	return e.path
}
func (e *Endpoint) GetRR(w dns.ResponseWriter, req *dns.Msg) ([]dns.RR, error) {
	if e.status {
		return []dns.RR{e.RR}, nil
	}
	return []dns.RR{}, ErrServiceStatusError
}

func (e *Endpoint) Verify(syntaxError *config.SyntaxError) {
	return
}
func (e *Endpoint) UpdateStatus() {

}
func (e *Endpoint) StartStatusWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e.status = <-e.StatusCh:
		}
	}
}

func init() {
	AddServicePlugin("endpoint", NewEndpoint)
}
