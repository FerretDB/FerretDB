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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/bson2"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/hex"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// dump returns string representation for debugging.
func dump[T types.Type](tb testtb.TB, o T) string {
	tb.Helper()

	v, err := bson2.Convert(o)
	require.NoError(tb, err)

	return bson2.LogMessageBlock(v)
}

// dumpSlice returns string representation for debugging.
func dumpSlice[T types.Type](tb testtb.TB, s []T) string {
	tb.Helper()

	arr := bson2.MakeArray(len(s))

	for _, o := range s {
		v, err := bson2.Convert(o)
		require.NoError(tb, err)

		err = arr.Add(v)
		require.NoError(tb, err)
	}

	return bson2.LogMessageBlock(arr)
}

// MustParseDumpFile panics if fails to parse file input to byte array.
func MustParseDumpFile(path ...string) []byte {
	b := must.NotFail(os.ReadFile(filepath.Join(path...)))
	return must.NotFail(hex.ParseDump(string(b)))
}

// IndentJSON returns an indented form of the JSON input.
func IndentJSON(tb testtb.TB, b []byte) []byte {
	tb.Helper()

	dst := bytes.NewBuffer(make([]byte, 0, len(b)))
	err := json.Indent(dst, b, "", "  ")
	require.NoError(tb, err)
	return dst.Bytes()
}

// Unindent removes the common number of leading tabs from all lines in s.
func Unindent(tb testtb.TB, s string) string {
	tb.Helper()

	require.NotEmpty(tb, s)

	parts := strings.Split(s, "\n")
	require.Positive(tb, len(parts))

	if parts[0] == "" {
		parts = parts[1:]
	}

	indent := len(parts[0]) - len(strings.TrimLeft(parts[0], "\t"))
	require.GreaterOrEqual(tb, indent, 0)

	for i := range parts {
		require.Greater(tb, len(parts[i]), indent, "line: %q", parts[i])
		parts[i] = parts[i][indent:]
	}

	return strings.Join(parts, "\n")
}
