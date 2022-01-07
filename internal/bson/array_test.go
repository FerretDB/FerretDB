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

package bson

import (
	"testing"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func convertArray(a *types.Array) *Array {
	res := Array(*a)
	return &res
}

var arrayTestCases = []testCase{{
	name: "array_all",
	v: convertArray(types.MustNewArray(
		types.MustNewArray(),
		types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
		true,
		time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC).Local(),
		types.MustMakeDocument(),
		42.13,
		int32(42),
		int64(42),
		"foo",
		nil,
	)),
	b: testutil.MustParseDumpFile("testdata", "array_all.hex"),
}, {
	name: "EOF",
	b:    []byte{0x00},
	bErr: `unexpected EOF`,
}, {
	name: "array_fuzz1",
	b:    testutil.MustParseDumpFile("testdata", "array_fuzz1.hex"),
	bErr: `key 0 is "8"`,
}}

func TestArray(t *testing.T) {
	t.Parallel()
	testBinary(t, arrayTestCases, func() bsontype { return new(Array) })
}

func FuzzArray(f *testing.F) {
	fuzzBinary(f, arrayTestCases, func() bsontype { return new(Array) })
}

func BenchmarkArray(b *testing.B) {
	benchmark(b, arrayTestCases, func() bsontype { return new(Array) })
}
