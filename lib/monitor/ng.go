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

	"github.com/spf13/viper"
)

type NG struct {
	path string
}

func NewNG(path string, v *viper.Viper) (Monitor, error) {
	return &NG{path: path}, nil
}
func (n *NG) Path() string {
	return n.path
}
func (n *NG) CheckRegister(e *Entry) error {
	return nil
}
func (n *NG) Run(ctx context.Context, e *Entry) bool {
	return false
}

func init() {
	AddMonitorPlugin("NG", NewNG)
}
