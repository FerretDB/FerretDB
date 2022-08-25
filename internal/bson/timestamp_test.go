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

	"github.com/AlekSi/pointer"
)

var timestampTestCases = []testCase{{
	name: "one",
	v:    pointer.To(timestampType(1)),
	b:    []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
}, {
	name: "zero",
	v:    pointer.To(timestampType(0)),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
}, {
	name: "EOF",
	b:    []byte{0x00},
	bErr: `unexpected EOF`,
}}

func TestTimestamp(t *testing.T) {
	t.Parallel()
	testBinary(t, timestampTestCases, func() bsontype { return new(timestampType) })
}

func FuzzTimestamp(f *testing.F) {
	fuzzBinary(f, timestampTestCases, func() bsontype { return new(timestampType) })
}

func BenchmarkTimestamp(b *testing.B) {
	benchmark(b, timestampTestCases, func() bsontype { return new(timestampType) })
}
