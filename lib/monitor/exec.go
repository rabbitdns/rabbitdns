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
	"errors"
	"os/exec"

	"github.com/mattn/go-shellwords"
	"github.com/spf13/viper"
)

var (
	ErrEmptyCommand = errors.New("empty command string or array.")
)

type ExecMonitor struct {
	command string
	path    string
	argv    []string
}

func NewExecMonitor(path string, v *viper.Viper) (Monitor, error) {
	execMonitor := &ExecMonitor{path: path}
	c := v.Get(path + ".command")
	if c == nil {
		return nil, ErrEmptyCommand
	}
	switch command := c.(type) {
	case string:
		args, err := shellwords.Parse(command)
		if err != nil || len(args) == 0 {
			return nil, ErrEmptyCommand
		}
		execMonitor.command = args[0]
		execMonitor.argv = args[1:]
	case []string:
		execMonitor.command = command[0]
		execMonitor.argv = command[1:]
	}

	return execMonitor, nil
}
func (em *ExecMonitor) Path() string {
	return em.path
}

func (em *ExecMonitor) CheckRegister(e *Entry) error {
	return nil
}
func (em *ExecMonitor) Run(ctx context.Context, e *Entry) bool {
	argv := make([]string, len(em.argv))
	copy(argv, em.argv)

	for i, v := range argv {
		if v == "%%ITEM%%" {
			argv[i] = e.Value
		}
	}
	errCh := make(chan error, 1)
	go func() {
		if len(argv) == 0 {
			errCh <- exec.Command(em.command).Run()
		} else {
			errCh <- exec.Command(em.command, argv...).Run()
		}
	}()

	select {
	case <-ctx.Done():
	case result := <-errCh:
		if result != nil {
			return true
		}
	}
	return false
}

func init() {
	AddMonitorPlugin("EXEC", NewExecMonitor)
}
