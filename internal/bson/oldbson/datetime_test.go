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

package oldbson

import (
	"testing"
	"time"

	"github.com/AlekSi/pointer"
)

var dateTimeTestCases = []testCase{{
	name: "2021",
	v:    pointer.To(dateTimeType(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC).Local())),
	b:    []byte{0x4b, 0x20, 0x02, 0xdb, 0x7c, 0x01, 0x00, 0x00},
}, {
	name: "unix_zero",
	v:    pointer.To(dateTimeType(time.Unix(0, 0))),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
}, {
	name: "0",
	v:    pointer.To(dateTimeType(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Local())),
	b:    []byte{0x00, 0xa0, 0xfb, 0x90, 0x75, 0xc7, 0xff, 0xff},
}, {
	name: "9999",
	v:    pointer.To(dateTimeType(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC).Local())),
	b:    []byte{0xff, 0xdb, 0x1f, 0xd2, 0x77, 0xe6, 0x00, 0x00},
}, {
	name: "EOF",
	b:    []byte{0x00},
	bErr: `unexpected EOF`,
}}

func TestDateTime(t *testing.T) {
	t.Parallel()
	testBinary(t, dateTimeTestCases, func() bsontype { return new(dateTimeType) })
}

func FuzzDateTime(f *testing.F) {
	fuzzBinary(f, dateTimeTestCases, func() bsontype { return new(dateTimeType) })
}

func BenchmarkDateTime(b *testing.B) {
	benchmark(b, dateTimeTestCases, func() bsontype { return new(dateTimeType) })
}
