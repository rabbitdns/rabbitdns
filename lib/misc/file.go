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
	"io"
	"os"
	"strconv"
)

func FileExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		return false
	}
	return true
}

func SaveToFile(path string, reader io.Reader) error {
	tmpfile := path + "." + strconv.Itoa(os.Getpid())
	f, err := os.Open(tmpfile)
	defer f.Close()

	if err != nil {
		return err
	}
	_, err = io.Copy(f, reader)
	if err != nil {
		return err
	}
	f.Close()

	err = os.Rename(tmpfile, path)

	return err
}
