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
	"errors"
	"strings"

	"github.com/miekg/dns"
	"github.com/rabbitdns/rabbitdns/lib/config"
	. "github.com/rabbitdns/rabbitdns/lib/misc"
	log "github.com/sirupsen/logrus"
)

var (
	ErrServ             = errors.New("failed to start dns server.")
	ErrZoneCutFind      = errors.New("failed to find zone cut.")
	ErrNotFoundZoneData = errors.New("faild to find zone data in zone node.")
)

const (
	Answer int = iota
	Authoritative
	Additional
)
const (
	NXDOMAIN int = iota
	FOUND
	NOTFOUND
)

type worker struct {
	listener       *dns.Server
	mux            *dns.ServeMux
	config         *config.Config
	zoneSet        *Tree
	serviceManager *serviceManager
}

func NewWorker(config *config.Config, zoneManager *zoneManager, serviceManager *serviceManager, addr string, proto string) *worker {
	var worker worker
	worker.config = config
	worker.zoneSet = zoneManager.zoneSet
	worker.serviceManager = serviceManager
	worker.mux = dns.NewServeMux()
	worker.mux.Handle(".", &worker)
	worker.listener = &dns.Server{Addr: addr,
		Net:           proto,
		Handler:       worker.mux,
		MaxTCPQueries: worker.config.MaxTCPQueries,
	}

	return &worker
}

func (s *worker) Run() {
	go func(l *dns.Server) {
		if err := l.ListenAndServe(); err != nil {
			log.WithFields(log.Fields{
				"Type":   "lib/server/Worker",
				"Func":   "Run",
				"Error":  err,
				"server": l,
			}).Fatal(ErrServ)
		}
	}(s.listener)
}

func (s *worker) serverDNSCAHOS(m *dns.Msg, req *dns.Msg) {
	qname := req.Question[0].Name
	m.MsgHdr.AuthenticatedData = false
	m.MsgHdr.CheckingDisabled = false
	m.MsgHdr.Authoritative = true

	if req.Question[0].Qtype != dns.TypeTXT {
		m.Rcode = dns.RcodeNXRrset
		return
	}
	switch qname {
	case "version.bind.", "version.server":
		hdr := dns.RR_Header{Name: qname, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
		m.Answer = []dns.RR{&dns.TXT{Hdr: hdr, Txt: []string{"1.0.0"}}}
		m.Rcode = dns.RcodeSuccess
	case "hostname.bind.", "id.server.":
		hdr := dns.RR_Header{Name: qname, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
		m.Answer = []dns.RR{&dns.TXT{Hdr: hdr, Txt: []string{"localhost."}}}
		m.Rcode = dns.RcodeSuccess
	default:
		m.Rcode = dns.RcodeNXRrset
	}
}
func (s *worker) SearchZone(labels []string) *Tree {
	tree := s.zoneSet.SearchNode(labels, false)
	for tree.Label != "" {
		if _, ok := tree.Get("provide"); ok == true {
			return tree
		}
		tree = tree.Parent
	}
	return nil
}

func (s *worker) serveDNSINET(w dns.ResponseWriter, m *dns.Msg, req *dns.Msg) error {
	qname := req.Question[0].Name
	labels := Labels(qname)
	zoneNode := s.SearchZone(labels)
	if zoneNode == nil {
		s.refused(m)
		return nil
	}
	if v, ok := zoneNode.Get("ZoneTree"); ok == true {
		m.Rcode = dns.RcodeNameError
		if zoneTree, ok := v.(*Tree); ok {
			err := s.servZoneResponse(w, m, req, qname, qname, req.Question[0].Qtype, zoneNode.Label, zoneTree, 16, false)
			if err != nil {
				return err
			}
			if len(m.Answer) > 0 {
				if s.config.MinimumResponse == false {
					if qname != zoneNode.Label || req.Question[0].Qtype != dns.TypeNS {
						s.addRR(m, zoneNode.Label, zoneTree, Authoritative, dns.TypeNS)
					}
				}
			}

			if len(m.Ns) > 0 {
				if s.config.MinimumResponse == false {
					for _, rr := range m.Ns {
						if ns, ok := rr.(*dns.NS); ok {
							s.addRR(m, ns.Ns, zoneTree, Additional, dns.TypeA)
							s.addRR(m, ns.Ns, zoneTree, Additional, dns.TypeAAAA)
						}
					}
				}
			}
			if len(m.Ns) == 0 && len(m.Answer) == 0 {
				s.addRR(m, zoneNode.Label, zoneTree, Authoritative, dns.TypeSOA)
			}
			return nil
		}
	}
	return ErrNotFoundZoneData
}

func (s *worker) addRR(m *dns.Msg, sname string, zoneTree *Tree, section int, rrType uint16) int {
	labels := Labels(sname)
	node := zoneTree.SearchNode(labels, true)
	if node == nil {
		return NXDOMAIN
	}
	if rrs, ok := node.GetRR(rrType); ok == true {
		switch section {
		case Answer:
			for _, rr := range rrs {
				m.Answer = append(m.Answer, rr)
			}
		case Authoritative:
			for _, rr := range rrs {
				m.Ns = append(m.Ns, rr)
			}
		case Additional:
			for _, rr := range rrs {
				m.Extra = append(m.Extra, rr)
			}
		}
		return FOUND
	}
	return NOTFOUND

}

func isAuth(t *Tree, zoneName string, rrtype uint16) bool {
	if rrtype == dns.TypeDS {
		return zoneName != t.Label
	}
	return t.Auth
}
func (s *worker) servZoneResponse(w dns.ResponseWriter, m *dns.Msg, req *dns.Msg, qname, sname string, stype uint16, zoneName string, zoneTree *Tree, count int, isWildcard bool) (err error) {
	labels := Labels(sname)
	if count <= 0 {
		return
	}

	node := zoneTree.SearchNode(labels, isWildcard)
	if node == nil {
		return
	}
	if isAuth(node, zoneName, stype) {
		if node.Label == sname {
			// found name
			m.Rcode = dns.RcodeSuccess
			if rrs, exist := node.GetRR(stype); exist {
				// found RR
				for _, rr := range rrs {
					if isWildcard {
						rr.Header().Name = qname
					}
					m.Answer = append(m.Answer, rr)
				}
			} else if rrs, exist := node.GetRR(dns.TypeCNAME); exist {
				// found CNAME
				m.Answer = append(m.Answer, rrs[0])
				if cname, ok := rrs[0].(*dns.CNAME); ok {
					err = s.servZoneResponse(w, m, req, cname.Target, cname.Target, stype, zoneName, zoneTree, count-1, false)
				}
			} else if dynamicRR, exist := StaticDynamicMap[stype]; exist {
				if rrs, ok := node.GetRR(dynamicRR); ok {
					if dyn, ok := rrs[0].(*dns.PrivateRR); ok {
						if rdata, ok := dyn.Data.(*DYNRR); ok {
							if resources, err := s.serviceManager.GetResources(w, req, stype, rdata.Resource); err == nil {
								for _, rr := range resources {
									rr.Header().Name = qname
									rr.Header().Rrtype = stype
									rr.Header().Class = dyn.Header().Class
									rr.Header().Ttl = dyn.Header().Ttl
									m.Answer = append(m.Answer, rr)
								}
							} else {
								return err
							}
						}
					}
				}
			}
		} else {
			if rrs, exist := node.GetRR(dns.TypeDNAME); exist {
				dname := rrs[0]
				dname.Header().Name = qname
				dname.Header().Ttl = 0
				m.Ns = append(m.Ns, dname)
			} else if !isWildcard {
				wildcard := FQDN("*." + strings.Join(labels[1:], "."))
				err = s.servZoneResponse(w, m, req, qname, wildcard, stype, zoneName, zoneTree, count-1, true)
			}
		}
		m.MsgHdr.Authoritative = true
	} else {
		// found Delegation
		zoneCut := node.FindZoneCut()
		if zoneCut == nil {
			return ErrZoneCutFind
		}
		if rrs, ok := zoneCut.GetRR(dns.TypeNS); ok == true {
			for _, rr := range rrs {
				m.Ns = append(m.Ns, rr)
			}
		}
		if rrs, ok := zoneCut.GetRR(dns.TypeDS); ok == true {
			for _, rr := range rrs {
				m.Ns = append(m.Ns, rr)
			}
		}
		m.MsgHdr.Authoritative = false
	}
	return
}

func (s *worker) servfail(m *dns.Msg) {
	m.Rcode = dns.RcodeServerFailure
}
func (s *worker) refused(m *dns.Msg) {
	m.Rcode = dns.RcodeRefused
}
func (s *worker) notImplemented(m *dns.Msg) {
	m.Rcode = dns.RcodeNotImplemented
}

func (s *worker) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)

	switch req.Question[0].Qclass {
	case dns.ClassCHAOS:
		s.serverDNSCAHOS(m, req)
	case dns.ClassINET:
		err := s.serveDNSINET(w, m, req)
		if err != nil {
			s.servfail(m)
		}
	default:
		s.notImplemented(m)
	}
	w.WriteMsg(m)
}
