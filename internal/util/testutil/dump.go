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

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/types/fjson"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// Dump returns string representation for debugging.
func Dump[T types.Type](tb testtb.TB, o T) string {
	tb.Helper()

	// We might switch to go-spew or something else later.
	b, err := fjson.Marshal(o)
	require.NoError(tb, err)

	return string(IndentJSON(tb, b))
}

// DumpSlice returns string representation for debugging.
func DumpSlice[T types.Type](tb testtb.TB, s []T) string {
	tb.Helper()

	// We might switch to go-spew or something else later.

	res := []byte("[")

	for i, o := range s {
		b, err := fjson.Marshal(o)
		require.NoError(tb, err)

		res = append(res, b...)
		if i < len(s)-1 {
			res = append(res, ',')
		}
	}

	res = append(res, ']')

	return string(IndentJSON(tb, res))
}

// IndentJSON returns an indented form of the JSON input.
func IndentJSON(tb testtb.TB, b []byte) []byte {
	tb.Helper()

	dst := bytes.NewBuffer(make([]byte, 0, len(b)))
	err := json.Indent(dst, b, "", "  ")
	require.NoError(tb, err)
	return dst.Bytes()
}
