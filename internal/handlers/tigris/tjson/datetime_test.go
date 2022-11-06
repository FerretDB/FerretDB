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

package tjson

import (
	"testing"
	"time"

	"github.com/AlekSi/pointer"
)

var dateTimeTestCases = []testCase{{
	name:   "2021",
	v:      pointer.To(dateTimeType(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))),
	schema: dateTimeSchema,
	j:      `"2021-11-01T10:18:42.123Z"`,
}, {
	name:   "unix_zero",
	v:      pointer.To(dateTimeType(time.Unix(0, 0).UTC())),
	schema: dateTimeSchema,
	j:      `"1970-01-01T00:00:00Z"`,
}, {
	name:   "0",
	v:      pointer.To(dateTimeType(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))),
	schema: dateTimeSchema,
	j:      `"0000-01-01T00:00:00Z"`,
}, {
	name:   "9999",
	v:      pointer.To(dateTimeType(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC))),
	schema: dateTimeSchema,
	j:      `"9999-12-31T23:59:59.999Z"`,
}, {
	name:   "EOF",
	schema: stringSchema,
	j:      `{`,
	jErr:   `unexpected EOF`,
}, {
	name:   "schema mismatch",
	schema: boolSchema,
	v:      pointer.To(dateTimeType(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))),
	j:      `"2021-11-01T10:18:42.123Z"`,
	sErr:   "json: cannot unmarshal string into Go value of type bool",
}, {
	name:   "invalid schema",
	schema: &Schema{Type: "invalid"},
	v:      pointer.To(dateTimeType(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))),
	j:      `"2021-11-01T10:18:42.123Z"`,
	sErr:   `tjson.Unmarshal: unhandled type "invalid"`,
}}

func TestDateTime(t *testing.T) {
	t.Parallel()
	testJSON(t, dateTimeTestCases, func() tjsontype { return new(dateTimeType) })
}

func FuzzDateTime(f *testing.F) {
	fuzzJSON(f, dateTimeTestCases)
}

func BenchmarkDateTime(b *testing.B) {
	benchmark(b, dateTimeTestCases)
}
