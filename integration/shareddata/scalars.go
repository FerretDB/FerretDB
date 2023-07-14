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

	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	doubleMaxPrec float64 = float64(1<<53 - 1) // 9007199254740991.0:    largest double values that could be represented as integer exactly
	doubleBig     float64 = float64(1 << 61)   // 2305843009213693952.0: some number larger than safe integer (doubleBig+1 == doubleBig)
	longBig       int64   = int64(1 << 61)     // 2305843009213693952:   same as doubleBig but integer

	doubleMaxOverflow float64 = float64(math.MaxInt64) + 1100.0 // 9223372036854777856.0: double value that is larger than MaxInt64
	doubleMinOverflow float64 = float64(math.MinInt64) - 1100.0 // -9223372036854777856.0: double value that is smaller than MinInt64
)

var doubleMaxVerge float64 = math.Nextafter(math.MaxFloat64, 0) // 1.7976931348623155e+308: double value that is close to the MaxFloat64 (1.7976931348623157e+308)

// Scalars contain scalar values for tests.
//
// This shared data set is frozen. If you need more values, add them in the test itself.
var Scalars = &Values[string]{
	name: "Scalars",
	data: map[string]any{
		"double":              42.13,
		"double-whole":        42.0,
		"double-zero":         0.0,
		"double-max":          math.MaxFloat64,
		"double-smallest":     math.SmallestNonzeroFloat64,
		"double-big":          doubleBig,
		"double-1":            float64(math.MinInt64 - 1),
		"double-2":            float64(math.MinInt64),
		"double-3":            float64(-123456789), // random
		"double-4":            float64(123456789),  // random
		"double-5":            float64(math.MaxInt64),
		"double-6":            float64(math.MaxInt64 + 1),
		"double-7":            doubleMaxVerge,
		"double-max-overflow": doubleMaxOverflow,
		"double-min-overflow": doubleMinOverflow,

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
		"int32-1":    int32(4080),
		"int32-2":    int32(1048560),   // random
		"int32-3":    int32(268435440), // random

		"timestamp":   primitive.Timestamp{T: 42, I: 13},
		"timestamp-i": primitive.Timestamp{I: 1},

		"int64":            int64(42),
		"int64-zero":       int64(0),
		"int64-max":        int64(math.MaxInt64),
		"int64-min":        int64(math.MinInt64),
		"int64-big":        longBig,
		"int64-double-big": doubleBig,
		"int64-1":          int64(1099511628000),     // random,
		"int64-2":          int64(281474976700000),   // random
		"int64-3":          int64(72057594040000000), // random

		// no 128-bit decimal floating point (yet)

		// no Min key
		// no Max key

		"unset": unset,
	},
}

// Doubles contains double values for tests.
var Doubles = &Values[string]{
	name: "Doubles",
	data: map[string]any{
		"double":          42.13,
		"double-whole":    42.0,
		"double-zero":     0.0,
		"double-smallest": math.SmallestNonzeroFloat64,

		// double big values ~1<<61
		"double-big":       doubleBig,
		"double-big-plus":  doubleBig + 1,
		"double-big-minus": doubleBig - 1,

		// double max precision ~1<<53 - 1
		"double-prec-max":          doubleMaxPrec,
		"double-prec-max-plus":     doubleMaxPrec + 1,
		"double-prec-max-plus-two": doubleMaxPrec + 2,
		"double-prec-max-minus":    doubleMaxPrec - 1,

		// negative double big values ~ -(1<<61)
		"double-neg-big":       -doubleBig,
		"double-neg-big-plus":  -doubleBig + 1,
		"double-neg-big-minus": -doubleBig - 1,

		// double min precision ~ -(1<<53 - 1)
		"double-prec-min":           -doubleMaxPrec,
		"double-prec-min-plus":      -doubleMaxPrec + 1,
		"double-prec-min-minus":     -doubleMaxPrec - 1,
		"double-prec-min-minus-two": -doubleMaxPrec - 2,

		"double-null":         nil,
		"double-1":            float64(math.MinInt64 - 1),
		"double-2":            float64(math.MinInt64),
		"double-3":            float64(-123456789), // random
		"double-4":            float64(123456789),  // random
		"double-5":            float64(math.MaxInt64),
		"double-6":            float64(math.MaxInt64 + 1),
		"double-max-overflow": doubleMaxOverflow,
		"double-min-overflow": doubleMinOverflow,
	},
}

// OverflowVergeDoubles contains double values which would overflow on
// numeric update operation such as $mul. Upon such,
// target returns error and compat returns +INF or -INF.
// OverflowVergeDoubles may be excluded on such update tests and tested
// in diff tests https://github.com/FerretDB/dance.
var OverflowVergeDoubles = &Values[string]{
	name: "OverflowVergeDoubles",
	data: map[string]any{
		"double-max": math.MaxFloat64,
		"double-7":   doubleMaxVerge,
	},
}

// SmallDoubles contains double values that does not go close to
// the maximum safe precision for tests.
var SmallDoubles = &Values[string]{
	name: "SmallDoubles",
	data: map[string]any{
		"double":       42.13,
		"double-whole": 42.0,
		"double-1":     4080.1234,    // random
		"double-2":     1048560.0099, // random
		"double-3":     268435440.2,  // random
	},
}

// Strings contains string values for tests.
var Strings = &Values[string]{
	name: "Strings",
	data: map[string]any{
		"string":           "foo",
		"string-double":    "42.13",
		"string-whole":     "42",
		"string-empty":     "",
		"string-duplicate": "foo",
		"string-null":      nil,
	},
}

// Binaries contains binary values for tests.
var Binaries = &Values[string]{
	name: "Binaries",
	data: map[string]any{
		"binary":       primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}},
		"binary-empty": primitive.Binary{Data: []byte{}},
		"binary-null":  nil,
	},
}

// ObjectIDs contains ObjectID values for tests.
var ObjectIDs = &Values[string]{
	name: "ObjectIDs",
	data: map[string]any{
		"objectid":       primitive.ObjectID{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11},
		"objectid-empty": primitive.NilObjectID,
		"objectid-null":  nil,
	},
}

// Bools contains bool values for tests.
var Bools = &Values[string]{
	name: "Bools",
	data: map[string]any{
		"bool-false": false,
		"bool-true":  true,
		"bool-null":  nil,
	},
}

// DateTimes contains datetime values for tests.
var DateTimes = &Values[string]{
	name: "DateTimes",
	data: map[string]any{
		"datetime":          primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC)),
		"datetime-epoch":    primitive.NewDateTimeFromTime(time.Unix(0, 0)),
		"datetime-year-min": primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)),
		"datetime-year-max": primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC)),
		"datetime-null":     nil,
	},
}

// Nulls contains null value for tests.
var Nulls = &Values[string]{
	name: "Nulls",
	data: map[string]any{
		"null": nil,
	},
}

// Regexes contains regex values for tests.
var Regexes = &Values[string]{
	name: "Regexes",
	data: map[string]any{
		"regex":       primitive.Regex{Pattern: "foo", Options: "i"},
		"regex-empty": primitive.Regex{},
		"regex-null":  nil,
	},
}

// Int32s contains int32 values for tests.
var Int32s = &Values[string]{
	name: "Int32s",
	data: map[string]any{
		"int32":      int32(42),
		"int32-zero": int32(0),
		"int32-max":  int32(math.MaxInt32),
		"int32-min":  int32(math.MinInt32),
		// "int32-null": nil, TODO: https://github.com/FerretDB/FerretDB/issues/1821
		"int32-1": int32(4080),
		"int32-2": int32(1048560),   // random
		"int32-3": int32(268435440), // random
	},
}

// Timestamps contains timestamp values for tests.
var Timestamps = &Values[string]{
	name: "Timestamps",
	data: map[string]any{
		"timestamp":      primitive.Timestamp{T: 42, I: 13},
		"timestamp-i":    primitive.Timestamp{I: 1},
		"timestamp-null": nil,
	},
}

// Int64s contains int64 values for tests.
var Int64s = &Values[string]{
	name: "Int64s",
	data: map[string]any{
		"int64":      int64(42),
		"int64-zero": int64(0),
		"int64-max":  int64(math.MaxInt64),
		"int64-min":  int64(math.MinInt64),
		// "int64-null": nil, TODO: https://github.com/FerretDB/FerretDB/issues/1821
		"int64-1": int64(1099511628000),     // random
		"int64-2": int64(281474976700000),   // random
		"int64-3": int64(72057594040000000), // random

		// long big values ~1<<61
		"int64-big":       longBig,
		"int64-big-plus":  longBig + 1,
		"int64-big-minus": longBig - 1,

		// long representation of double max precision ~1<<53 - 1
		"int64-prec-max":          int64(doubleMaxPrec),
		"int64-prec-max-plus":     int64(doubleMaxPrec) + 1,
		"int64-prec-max-plus-two": int64(doubleMaxPrec) + 2,
		"int64-prec-max-minus":    int64(doubleMaxPrec) - 1,

		// negative long big values ~ -(1<<61)
		"int64-neg-big":       -longBig,
		"int64-neg-big-plus":  -longBig + 1,
		"int64-neg-big-minus": -longBig - 1,

		// long representation of double min precision ~ -(1<<53 - 1)
		"int64-prec-min":           -int64(doubleMaxPrec),
		"int64-prec-min-plus":      -int64(doubleMaxPrec) + 1,
		"int64-prec-min-minus":     -int64(doubleMaxPrec) - 1,
		"int64-prec-min-minus-two": -int64(doubleMaxPrec) - 2,
	},
}

// Unsets contains unset value for tests.
var Unsets = &Values[string]{
	name: "Unsets",
	data: map[string]any{
		"unset": unset,
	},
}

// ObjectIDKeys contains documents with ObjectID keys for tests.
var ObjectIDKeys = &Values[primitive.ObjectID]{
	name: "ObjectIDKeys",
	data: map[primitive.ObjectID]any{
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11}: "objectid",
		primitive.NilObjectID: "objectid-empty",
	},
}

func init() {
	// If any of those assumptions fails, it means that
	// there's an issue related to precision double conversion.

	must.BeTrue(float64(int64(doubleBig)) == doubleBig)
	must.BeTrue(float64(int64(doubleBig)+1) == doubleBig)

	must.BeTrue(float64(longBig) == doubleBig)

	must.BeTrue(doubleMaxPrec != doubleMaxPrec+1)
	must.BeTrue(doubleMaxPrec+1 == doubleMaxPrec+2)

	must.BeTrue(-doubleMaxPrec != -doubleMaxPrec-1)
	must.BeTrue(-doubleMaxPrec-1 == -doubleMaxPrec-2)

	must.BeTrue(doubleMaxOverflow > float64(math.MaxInt64))
	must.BeTrue(doubleMinOverflow <= float64(math.MinInt64))

	must.BeTrue(doubleMaxVerge < math.MaxFloat64)
}
