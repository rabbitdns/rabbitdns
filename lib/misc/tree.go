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
	ErrChildExist = errors.New("child node is exists")
)

type Tree struct {
	Label      string
	Paramaters map[string]interface{}
	Parent     *Tree
	Children   map[string]*Tree
	Resources  map[uint16][]dns.RR
	Auth       bool
}

func NewTree() *Tree {
	return &Tree{Label: "",
		Paramaters: map[string]interface{}{},
		Children:   map[string]*Tree{},
		Resources:  map[uint16][]dns.RR{},
		Auth:       false,
	}
}

func (t *Tree) AddNode(labels []string) *Tree {
	var child *Tree
	if len(labels) == 0 {
		return t
	}

	last := labels[len(labels)-1]
	labels = labels[:len(labels)-1]
	if v, ok := t.Children[last]; ok {
		child = v
	} else {
		child = &Tree{
			Label:      last + "." + t.Label,
			Parent:     t,
			Paramaters: map[string]interface{}{},
			Children:   map[string]*Tree{},
			Resources:  map[uint16][]dns.RR{},
			Auth:       t.Auth,
		}
		t.Children[last] = child
	}
	return child.AddNode(labels)
}

func (t *Tree) SearchNode(labels []string, strict bool) *Tree {
	if len(labels) == 0 {
		return t
	}

	last := labels[len(labels)-1]
	labels = labels[:len(labels)-1]
	if v, ok := t.Children[last]; ok {
		return v.SearchNode(labels, strict)
	}
	if strict {
		return nil
	}
	return t
}

func (t *Tree) DeleteNode(labels []string, force bool) error {
	last := labels[len(labels)-1]
	labels = labels[:len(labels)-1]
	if v, ok := t.Children[last]; ok {
		if len(labels) == 0 {
			if force || len(t.Children) == 0 {
				delete(t.Children, labels[0])
			} else {
				return ErrChildExist
			}
		} else {
			return v.DeleteNode(labels, force)
		}
	}
	return nil
}

func (t *Tree) Set(name string, value interface{}) {
	t.Paramaters[name] = value
}
func (t *Tree) Get(name string) (interface{}, bool) {
	v, ok := t.Paramaters[name]
	return v, ok
}
func (t *Tree) Delete(name string) {
	delete(t.Paramaters, name)
}
func (t *Tree) DeleteAll() {
	for name, _ := range t.Paramaters {
		delete(t.Paramaters, name)
	}
}

// for RR
func (t *Tree) AddRR(rr dns.RR) *Tree {
	labels := Labels(rr.Header().Name)
	rrNode := t.AddNode(labels)
	rrNode.SetRR(rr)
	return rrNode
}

func (t *Tree) SetRR(rr dns.RR) {
	t.Resources[rr.Header().Rrtype] = append(t.Resources[rr.Header().Rrtype], rr)
}

func (t *Tree) GetRR(rrType uint16) ([]dns.RR, bool) {
	v, ok := t.Resources[rrType]
	return v, ok
}

func (t *Tree) DeleteRR(rrType uint16, rr dns.RR) {
	delete(t.Resources, rrType)
}
