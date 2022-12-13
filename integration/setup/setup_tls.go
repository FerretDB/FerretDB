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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// GetTLSFilesPaths returns paths to TLS files.
func GetTLSFilesPaths(t testing.TB) (string, string) {
	certPath := filepath.Join("..", "build", "certs", "server-cert.pem")

	_, err := os.Stat(certPath)
	require.NoError(t, err)

	keyPath := filepath.Join("..", "build", "certs", "server-key.pem")

	_, err = os.Stat(keyPath)
	require.NoError(t, err)

	return certPath, keyPath
}
