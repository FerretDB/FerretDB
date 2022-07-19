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

package shareddata

import (
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	doubleBig = float64(2 << 60)
	int64Big  = int64(2 << 61)
)

// Scalars contain scalar values for tests.
//
// This shared data set is frozen. If you need more values, add them in the test itself.
var Scalars = &Values[string]{
	data: map[string]any{
		"double":                   42.13,
		"double-whole":             42.0,
		"double-zero":              math.Copysign(0, +1), // the same as just 0.0 in Go
		"double-negative-zero":     math.Copysign(0, -1),
		"double-max":               math.MaxFloat64,
		"double-smallest":          math.SmallestNonzeroFloat64,
		"double-positive-infinity": math.Inf(+1),
		"double-negative-infinity": math.Inf(-1),
		"double-nan":               math.NaN(),
		"double-big":               doubleBig,

		"string":        "foo",
		"string-double": "42.13",
		"string-whole":  "42",
		"string-empty":  "",

		"binary":       primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}},
		"binary-empty": primitive.Binary{Data: []byte{}},

		"objectid":       primitive.ObjectID{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11},
		"objectid-empty": primitive.NilObjectID,

		// no Undefined

		"bool-false": false,
		"bool-true":  true,

		"datetime":          primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC)),
		"datetime-epoch":    primitive.NewDateTimeFromTime(time.Unix(0, 0)),
		"datetime-year-min": primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)),
		"datetime-year-max": primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC)),

		"null": nil,

		"regex":       primitive.Regex{Pattern: "foo", Options: "i"},
		"regex-empty": primitive.Regex{},

		// no DBPointer
		// no JavaScript code
		// no Symbol
		// no JavaScript code w/ scope

		"int32":      int32(42),
		"int32-zero": int32(0),
		"int32-max":  int32(math.MaxInt32),
		"int32-min":  int32(math.MinInt32),

		"timestamp":   primitive.Timestamp{T: 42, I: 13},
		"timestamp-i": primitive.Timestamp{I: 1},

		"int64":      int64(42),
		"int64-zero": int64(0),
		"int64-max":  int64(math.MaxInt64),
		"int64-min":  int64(math.MinInt64),
		"int64-big":  int64Big,

		// no 128-bit decimal floating point (yet)

		// no Min key
		// no Max key

		// TODO "unset": unset, https://github.com/FerretDB/FerretDB/issues/914
	},
}

// FixedScalars is an experiment and will be changed in the future.
//
// TODO https://github.com/FerretDB/FerretDB/issues/786
var FixedScalars = &Maps[string]{
	data: map[string]map[string]any{
		"fixed_double":          {"double_value": 42.13},
		"fixed_double-whole":    {"double_value": 42.0},
		"fixed_double-zero":     {"double_value": 0.0},
		"fixed_double-max":      {"double_value": math.MaxFloat64},
		"fixed_double-smallest": {"double_value": math.SmallestNonzeroFloat64},
		"fixed_double-big":      {"double_value": doubleBig},
	},
}
