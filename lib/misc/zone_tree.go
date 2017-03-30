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
	"errors"

	"github.com/miekg/dns"
)

var (
	ErrVerifyZoneEmptyApex           = errors.New("Apex node is empty.")
	ErrVerifyZoneEmptySOA            = errors.New("SOA RR is empty.")
	ErrVerifyZoneDupulicateSOA       = errors.New("More than 1 SOA RR found.")
	ErrVerifyZoneEmptyApexNS         = errors.New("Apex NS not found.")
	ErrVerifyNodeDupulicateCNAME     = errors.New("More than 1 CNAME RR found in same name.")
	ErrVerifyNodeOtherRRInCNAMENode  = errors.New("Found other RR in CNAME node.")
	ErrVerifyNodeDupulicateDNAME     = errors.New("More than 1 DNAME RR found in same name.")
	ErrVerifyNodeFoundChildNodeDNAME = errors.New("Found child node in DNAME node.")
)

func (t *Tree) VerifyZone(origin_labels []string) error {
	// APEX CHECK
	apex := t.SearchNode(origin_labels, true)
	if apex == nil {
		return ErrVerifyZoneEmptyApex
	}
	// SOA CHECK
	soa, exist := apex.GetRR(dns.TypeSOA)
	if !exist {
		return ErrVerifyZoneEmptySOA
	}
	if len(soa) > 1 {
		return ErrVerifyZoneDupulicateSOA
	}
	// Zone APEX NS CHECK
	if _, exist := apex.GetRR(dns.TypeNS); !exist {
		return ErrVerifyZoneEmptyApexNS
	}

	return apex.VerifyNode()
}

func (t *Tree) VerifyNode() error {
	if len(t.Resources) > 0 {
		// CNAME CHECK
		if cname, exist := t.GetRR(dns.TypeCNAME); exist {
			if len(cname) > 1 {
				return ErrVerifyNodeDupulicateCNAME
			}
			if len(t.Resources) > 2 {
				return ErrVerifyNodeOtherRRInCNAMENode
			}
			for rrtype := range t.Resources {
				if rrtype != dns.TypeCNAME && rrtype != dns.TypeDNAME {
					return ErrVerifyNodeOtherRRInCNAMENode
				}
			}
		}
		// DNAME CHECK
		if dname, exist := t.GetRR(dns.TypeDNAME); exist {
			if len(dname) > 1 {
				return ErrVerifyNodeDupulicateDNAME
			}
			if len(t.Children) > 0 {
				return ErrVerifyNodeFoundChildNodeDNAME
			}
		}
	}
	for _, child := range t.Children {
		if err := child.VerifyNode(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tree) FindZoneCut() *Tree {
	if t.Auth {
		return nil
	}
	if t.Parent.Auth {
		return t
	}
	return t.Parent.FindZoneCut()
}
