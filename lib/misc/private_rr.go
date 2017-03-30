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

package misc

import (
	"strings"

	"github.com/miekg/dns"
)

const (
	TypeDYNC    = 0xFF10
	TypeDYNA    = 0xFF11
	TypeDYNAAAA = 0xFF12
	TypeDYNTXT  = 0xFF13
	TypeDYNMX   = 0xFF14
	TypeDYNPTR  = 0xFF15
	TypeDYNSRV  = 0xFF16
)

var DynamicStaticMap = map[uint16]uint16{
	TypeDYNC:    dns.TypeCNAME,
	TypeDYNA:    dns.TypeA,
	TypeDYNAAAA: dns.TypeAAAA,
	TypeDYNTXT:  dns.TypeTXT,
	TypeDYNMX:   dns.TypeMX,
	TypeDYNPTR:  dns.TypePTR,
	TypeDYNSRV:  dns.TypeSRV,
}

var StaticDynamicMap = reverseUint16toUint16(DynamicStaticMap)

type DYNRR struct {
	Resource string `dns:"octet"`
}

func init() {
	dns.PrivateHandle("DYNC", TypeDYNC, NewDYNARR)
	dns.PrivateHandle("DYNA", TypeDYNA, NewDYNARR)
	dns.PrivateHandle("DYNAAAA", TypeDYNAAAA, NewDYNARR)
	dns.PrivateHandle("DYNTXT", TypeDYNTXT, NewDYNARR)
	dns.PrivateHandle("DYNMX", TypeDYNMX, NewDYNARR)
	dns.PrivateHandle("DYNPTR", TypeDYNPTR, NewDYNARR)
	dns.PrivateHandle("DYNSRV", TypeDYNSRV, NewDYNARR)
}

func NewDYNARR() dns.PrivateRdata { return &DYNRR{} }

func (rd *DYNRR) Len() int       { return len([]byte(rd.Resource)) }
func (rd *DYNRR) String() string { return rd.Resource }
func (rd *DYNRR) Parse(txt []string) error {
	rd.Resource = strings.TrimSpace(strings.Join(txt, " "))
	return nil
}

func (rd *DYNRR) Pack(buf []byte) (int, error) {
	b := []byte(rd.Resource)
	n := copy(buf, b)
	if n != len(b) {
		return n, dns.ErrBuf
	}
	return n, nil
}

func (rd *DYNRR) Unpack(buf []byte) (int, error) {
	rd.Resource = string(buf)
	return len(buf), nil
}

func (rd *DYNRR) Copy(dest dns.PrivateRdata) error {
	d, ok := dest.(*DYNRR)
	if !ok {
		return dns.ErrRdata
	}
	d.Resource = rd.Resource
	return nil
}
