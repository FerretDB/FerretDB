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

package testutil

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// WriteSeedCorpusFile adds given data to the fuzzing seed corpus for given fuzz function.
//
// It can be an alternative to using f.Add.
func WriteSeedCorpusFile(tb testing.TB, funcName string, b []byte) {
	tb.Helper()

	var buf bytes.Buffer
	buf.WriteString("go test fuzz v1\n")
	_, err := fmt.Fprintf(&buf, "[]byte(%q)\n", b)
	require.NoError(tb, err)

	dir := filepath.Join("testdata", "fuzz", funcName)
	err = os.MkdirAll(dir, 0o777)
	require.NoError(tb, err)

	filename := filepath.Join(dir, fmt.Sprintf("test-%x", sha256.Sum256(b)))
	err = os.WriteFile(filename, buf.Bytes(), 0o666)
	require.NoError(tb, err)
}
