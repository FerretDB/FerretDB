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
	"os"
	"path/filepath"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/hex"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ParseDump parses string to bytes, in tests.
func ParseDump(tb TB, s string) []byte {
	tb.Helper()

	b, err := hex.ParseDump(s)
	require.NoError(tb, err)
	return b
}

// ParseDumpFile parses file input to bytes, in tests.
func ParseDumpFile(tb TB, path ...string) []byte {
	tb.Helper()

	b, err := os.ReadFile(filepath.Join(path...))
	require.NoError(tb, err)
	return ParseDump(tb, string(b))
}

// MustParseDumpFile panics if fails to parse file input to byte array.
func MustParseDumpFile(path ...string) []byte {
	b := must.NotFail(os.ReadFile(filepath.Join(path...)))
	return must.NotFail(hex.ParseDump(string(b)))
}
