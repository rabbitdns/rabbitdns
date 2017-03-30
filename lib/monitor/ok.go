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

type OK struct {
	path string
}

func NewOK(path string, v *viper.Viper) (Monitor, error) {
	return &OK{}, nil
}
func (o *OK) Path() string {
	return o.path
}

func (o *OK) CheckRegister(e *Entry) error {
	return nil
}
func (o *OK) Run(ctx context.Context, e *Entry) bool {
	return true
}

func init() {
	AddMonitorPlugin("OK", NewOK)
}
