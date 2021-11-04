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

	"github.com/AlekSi/pointer"
)

var dateTimeTestCases = []testCase{{
	name: "2021",
	v:    pointer.To(DateTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))),
	b:    []byte{0x4b, 0x20, 0x02, 0xdb, 0x7c, 0x01, 0x00, 0x00},
	j:    `{"$d":"1635761922123"}`,
}, {
	name: "zero",
	v:    pointer.To(DateTime(time.Unix(0, 0).UTC())),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	j:    `{"$d":"0"}`,
}, {
	name: "1000",
	v:    pointer.To(DateTime(time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC))),
	b:    []byte{0x00, 0xd4, 0x78, 0x00, 0x29, 0xe4, 0xff, 0xff},
	j:    `{"$d":"-30610224000000"}`,
}, {
	name: "3000",
	v:    pointer.To(DateTime(time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC))),
	b:    []byte{0x00, 0xe0, 0x4c, 0xda, 0x8f, 0x1d, 0x00, 0x00},
	j:    `{"$d":"32503680000000"}`,
}}

func TestDateTime(t *testing.T) {
	t.Parallel()

	t.Run("Binary", func(t *testing.T) {
		t.Parallel()
		testBinary(t, dateTimeTestCases, func() bsontype { return new(DateTime) })
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		testJSON(t, dateTimeTestCases, func() bsontype { return new(DateTime) })
	})
}

func FuzzDateTimeBinary(f *testing.F) {
	fuzzBinary(f, dateTimeTestCases, func() bsontype { return new(DateTime) })
}

func FuzzDateTimeJSON(f *testing.F) {
	fuzzJSON(f, dateTimeTestCases, func() bsontype { return new(DateTime) })
}

func BenchmarkDateTime(b *testing.B) {
	benchmark(b, dateTimeTestCases, func() bsontype { return new(DateTime) })
}
