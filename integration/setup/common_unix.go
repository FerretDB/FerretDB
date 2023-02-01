// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build unix

package setup

import (
	"os"
)

// unixSocketPath returns temporary Unix domain socket path for that test.
func unixSocketPath() (string, error) {
	f, err := os.CreateTemp("", "ferretdb-*.sock")
	if err != nil {
		return "", err
	}

	// remove file so listener could create it (and remove it itself on stop)
	if err = f.Close(); err != nil {
		return "", err
	}
	if err = os.Remove(f.Name()); err != nil {
		return "", err
	}

	return f.Name(), nil
}
