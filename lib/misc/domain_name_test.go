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

import "testing"

func TestFQDN(t *testing.T) {
	if FQDN("") != "." {
		t.Fatalf(" empty string FQDN is \".\"")
	}
	if FQDN(".") != "." {
		t.Fatalf(" .  FQDN is \".\"")
	}
	if FQDN("example.jp") != "example.jp." {
		t.Fatalf("example.jp  FQDN is \"example.jp.\"")
	}
	if FQDN("example.jp.") != "example.jp." {
		t.Fatalf("example.jp.  FQDN is \"example.jp.\"")
	}
}
