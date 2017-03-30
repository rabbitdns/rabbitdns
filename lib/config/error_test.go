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

package config

import "testing"
import "fmt"

func TestSyntaxError(t *testing.T) {
	se := &SyntaxError{}

	if se.Return() != nil {
		t.Errorf("SyntaxError.Return must return nil when empty errors.")
	}

	se.Add(fmt.Errorf("hoge"))
	if se.Return() == nil {
		t.Errorf("SyntaxError.Return must not return nil when not empty errors.")
	}

	if se.Error() != "hoge\n" {
		t.Errorf("SyntaxError.Error() must return error string. %s", se.Error())
	}

	se.Add(fmt.Errorf("fuga"))
	if se.Error() != "hoge\nfuga\n" {
		t.Errorf("SyntaxError.Error() must return error string.")
	}
}
