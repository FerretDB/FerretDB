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
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// listenUnix returns temporary Unix domain socket path for that test.
func listenUnix(tb testing.TB) string {
	// The MongoDB unix socket path with upper case letter does not work.
	path := strings.ToLower(tb.TempDir())

	// The path must exist.
	err := os.MkdirAll(path, os.ModePerm)
	require.NoError(tb, err)

	socketPath := filepath.Join(path, "ferretdb.sock")

	// The unix socket path must be less than 108 chars.
	// https://man7.org/linux/man-pages/man7/unix.7.html
	if len(socketPath) >= 108 {
		tb.Fatalf("listen unix socket path of length %d is too long: %d %s", len(socketPath), socketPath)
	}

	return socketPath
}
