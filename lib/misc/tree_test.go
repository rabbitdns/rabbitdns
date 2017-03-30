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
	"bytes"
	"testing"

	"github.com/miekg/dns"
)

func TestTreeAddNode(t *testing.T) {
	root := NewTree()
	if root.Label != "" {
		t.Errorf("root label need to be empty")
	}
	node := root.AddNode([]string{"www", "example", "jp"})
	node.Set("a", "b")
	if node.Label != "www.example.jp." {
		t.Errorf("label is need to be \"www.example.jp.\"")
	}

	node2 := root.SearchNode([]string{"example", "jp"}, false)
	if node2.Label != "example.jp." {
		t.Errorf("label is need to be \"example.jp.\"")
	}

	node3 := root.SearchNode([]string{"hogehoge", "example", "jp"}, false)
	if node3.Label != "example.jp." {
		t.Errorf("[longest match test] label is need to be \"example.jp.\"")
	}

	node4 := root.SearchNode([]string{"hogehoge", "example", "jp"}, true)
	if node4 != nil {
		t.Errorf("[strict match test] label is result to be nil")
	}

	node = root.AddNode([]string{"www", "example", "com"})
	node.Set("zonename", "hogehoge")
	if str, ok := node.Get("zonename"); ok {
		if str != "hogehoge" {
			t.Errorf("[set/get resource test] zonename is need to bo be hogehoge")
		}
	} else {
		t.Errorf("[set/get resource test] zonename is need to not empty")
	}
	node.Delete("zonename")
	if _, ok := node.Get("zonename"); ok {
		t.Errorf("[delete resource test] zonename is need to be empty")
	}

	node.Delete("hogehoge")
	if _, ok := node.Get("hoge2"); ok {
		t.Errorf("[get resource test] hoge2 is need to be empty")
	}

}

func TestingLoadZoneFile(t *testing.T) {
	var zoneData = `$ORIGIN www.example.com.
$TTL 300
@ IN SOA z.dns.jp. root.dns.jp. 1535461207 3600 900 1814400 900
@ IN NS ns.example.com.
@ IN NS ns.example.jp.
@ IN DYNA geoip!www
www IN DYNAC hoge
`
	root := NewTree()
	buffer := bytes.NewBufferString(zoneData)
	for x := range dns.ParseZone(buffer, "www.example.com.", "") {
		if x.Error != nil {
			t.Errorf(x.Error.Error())
		} else {
			root.AddRR(x.RR)
		}
	}
	node := root.SearchNode([]string{"www", "example", "jp"}, true)
	if node.Label != "www.example.jp." {
		t.Errorf("label need zone origin")
	}
	rrs, ok := node.GetRR(dns.TypeSOA)
	if ok == false {
		t.Errorf(" need SOA record")
	}
	rr := rrs[0]
	if v, ok := rr.(*dns.SOA); ok {
		if v.Ns != "z.dns.jp." {
			t.Errorf("error ns")
		}
	}
	rrs, ok = node.GetRR(dns.TypeNS)
	if len(rrs) != 2 {
		t.Errorf("need two ns.")
	}
	rrs, ok = node.GetRR(TypeDYNA)
	if ok == false {
		t.Errorf("need DYNA")
	}
	rr = rrs[0]
	if v, ok := rr.(*dns.PrivateRR); ok {
		if rdata, ok := v.Data.(*DYNRR); ok {
			if rdata.Resource != "geoip!www" {
				t.Errorf("error DYNA")
			}
		} else {
			t.Errorf("error DYNA")
		}
	} else {
		t.Errorf("error DYNA")
	}

}
