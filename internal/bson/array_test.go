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

var arrayTestcases = []fuzzTestCase{{
	name: "all",
	v: &Array{
		types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
		true,
		time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC),
		42.13,
		int32(42),
		int64(42),
		"foo",
	},
	b: testutil.MustParseDump(`
		00000000  45 00 00 00 05 30 00 01  00 00 00 80 42 08 31 00  |E....0......B.1.|
		00000010  01 09 32 00 2b e6 51 e7  7a 01 00 00 01 33 00 71  |..2.+.Q.z....3.q|
		00000020  3d 0a d7 a3 10 45 40 10  34 00 2a 00 00 00 12 35  |=....E@.4.*....5|
		00000030  00 2a 00 00 00 00 00 00  00 02 36 00 04 00 00 00  |.*........6.....|
		00000040  66 6f 6f 00 00                                    |foo..|
	`),
	j: `[{"$b":"Qg==","s":128},true,{"$d":"1627378542123"},{"$f":"42.13"},42,{"$l":"42"},"foo"]`,
}}

func TestArray(t *testing.T) {
	t.Parallel()

	t.Run("Binary", func(t *testing.T) {
		t.Parallel()
		testBinary(t, arrayTestcases, func() bsontype { return new(Array) })
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		testJSON(t, arrayTestcases, func() bsontype { return new(Array) })
	})
}

func FuzzArrayBinary(f *testing.F) {
	fuzzBinary(f, arrayTestcases, func() bsontype { return new(Array) })
}

func FuzzArrayJSON(f *testing.F) {
	fuzzJSON(f, arrayTestcases, func() bsontype { return new(Array) })
}

func BenchmarkArray(b *testing.B) {
	benchmark(b, arrayTestcases, func() bsontype { return new(Array) })
}
