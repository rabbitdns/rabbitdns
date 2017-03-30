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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	ErrEmptyUrl             = errors.New("Empty url")
	ErrInvilidURL           = errors.New("Invilid url")
	ErrHttpNotSupportRR     = errors.New("http only support A,AAAA and CNAME.")
	ErrHttpNotSupportMethod = errors.New("Not support this http method.")
)

type HTTP struct {
	path   string
	url    string
	method string
	check  string
	host   string

	body io.Reader
}

func NewHTTP(path string, v *viper.Viper) (Monitor, error) {
	url := v.GetString(path + ".url")
	if url == "" {
		return nil, ErrEmptyUrl
	}
	method := v.GetString(path + ".method")
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)
	switch method {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE":
	default:
		return nil, ErrHttpNotSupportMethod
	}
	check := v.GetString(path + ".check")
	http := &HTTP{path: path, url: url, method: method, check: check}
	if data := v.GetString(path + ".data"); data != "" {
		http.body = strings.NewReader(data)
	}
	return http, nil
}

func (h *HTTP) Path() string {
	return h.path
}

func (h *HTTP) CheckRegister(e *Entry) error {
	switch e.RRtype {
	case dns.TypeA, dns.TypeAAAA, dns.TypeCNAME:
	default:
		return ErrHttpNotSupportRR
	}
	return nil
}

func (h *HTTP) Run(ctx context.Context, e *Entry) bool {
	urlStr := h.url

	var connectTo string
	switch e.RRtype {
	case dns.TypeAAAA:
		connectTo = "[" + e.Value + "]"
	case dns.TypeA, dns.TypeCNAME:
		connectTo = e.Value
	}

	client := new(http.Client)
	url := strings.Replace(urlStr, "%%ITEM%%", connectTo, -1)
	req, err := http.NewRequest(h.method, url, h.body)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(byteArray))

	return true
}

func init() {
	AddMonitorPlugin("http", NewTcpcon)
}
