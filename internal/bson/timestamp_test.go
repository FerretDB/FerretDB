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
	v:    pointer.To(Timestamp(1)),
	b:    []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	j:    `{"$t":"1"}`,
}, {
	name: "zero",
	v:    pointer.To(Timestamp(0)),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	j:    `{"$t":"0"}`,
}}

func TestTimestamp(t *testing.T) {
	t.Parallel()

	t.Run("Binary", func(t *testing.T) {
		t.Parallel()
		testBinary(t, timestampTestCases, func() bsontype { return new(Timestamp) })
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		testJSON(t, timestampTestCases, func() bsontype { return new(Timestamp) })
	})
}

func FuzzTimestampBinary(f *testing.F) {
	fuzzBinary(f, timestampTestCases, func() bsontype { return new(Timestamp) })
}

func FuzzTimestampJSON(f *testing.F) {
	fuzzJSON(f, timestampTestCases, func() bsontype { return new(Timestamp) })
}

func BenchmarkTimestamp(b *testing.B) {
	benchmark(b, timestampTestCases, func() bsontype { return new(Timestamp) })
}
