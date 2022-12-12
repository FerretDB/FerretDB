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

package setup

import (
	"os"
	"path"
)

// GetTLSFilesPaths returns paths to TLS files.
func GetTLSFilesPaths() (string, string) {
	certPath := path.Join("..", "build", "certs", "server-cert.pem")

	_, err := os.Stat(certPath)
	if err != nil {
		panic("server certificate not found")
	}

	keyPath := path.Join("..", "build", "certs", "server-key.pem")

	_, err = os.Stat(keyPath)
	if err != nil {
		panic("server key not found")
	}

	return certPath, keyPath
}
