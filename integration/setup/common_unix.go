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
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chars used for generating random string.
// MongoDB Unix socket path with upper case letter does not
// work hence using lower case alphanumeric.
const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

// getRandomString returns a random string in lower case alphanumeric of given length.
// It is intended for generating random for integration testing,
// but not recommended for reusing it for other purpose.
func getRandomString(tb testing.TB, length int) string {
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		require.NoError(tb, err)

		b[i] = chars[num.Int64()]
	}

	return string(b)
}

// listenUnix returns temporary Unix domain socket path for that test.
func listenUnix(tb testing.TB) string {
	// generate random string of length 20 for the directory name.
	dirName := getRandomString(tb, 20)

	// on mac, temp dir is length 49 and like /var/folders/9p/cc9b8krs2zd1x9qx89fs1sjw0000gn/t/.
	// on linux, temp dir is length 5 and like /tmp/.
	tmp := os.TempDir()
	basePath := filepath.Join(tmp, dirName)

	// The path must exist.
	err := os.MkdirAll(basePath, os.ModePerm)
	require.NoError(tb, err)

	socketPath := filepath.Join(basePath, "ferretdb.sock")

	if len(socketPath) >= 104 {
		// This is a way to fail fast before creating a client for this socket.
		// Unix socket path must be less than 104 chars for mac, 108 for linux.
		tb.Fatalf("listen Unix socket path too long len: %d, path: %s", len(socketPath), socketPath)
	}

	tb.Cleanup(func() {
		err := os.RemoveAll(basePath)
		assert.NoError(tb, err)
	})

	return socketPath
}
