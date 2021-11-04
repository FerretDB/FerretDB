// Copyright 2021 Baltoro OÜ.
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

	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/testutil"
)

var arrayTestCases = []testCase{{
	name: "array_all",
	v: &Array{
		types.Array{},
		types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
		true,
		time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC),
		types.MakeDocument(),
		42.13,
		int32(42),
		int64(42),
		"foo",
		nil,
	},
	b: testutil.MustParseDumpFile("testdata", "array_all.hex"),
	j: `[[],{"$b":"Qg==","s":128},true,{"$d":"1627378542123"},{"$k":[]},{"$f":"42.13"},42,{"$l":"42"},"foo",null]`,
}}

func TestArray(t *testing.T) {
	t.Parallel()

	t.Run("Binary", func(t *testing.T) {
		t.Parallel()
		testBinary(t, arrayTestCases, func() bsontype { return new(Array) })
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		testJSON(t, arrayTestCases, func() bsontype { return new(Array) })
	})
}

func FuzzArrayBinary(f *testing.F) {
	fuzzBinary(f, arrayTestCases, func() bsontype { return new(Array) })
}

func FuzzArrayJSON(f *testing.F) {
	fuzzJSON(f, arrayTestCases, func() bsontype { return new(Array) })
}

func BenchmarkArray(b *testing.B) {
	benchmark(b, arrayTestCases, func() bsontype { return new(Array) })
}
