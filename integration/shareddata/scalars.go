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
	"strings"
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
	name:     "Scalars",
	handlers: []string{"pg"},
	data: map[string]any{
		"double":               42.13,
		"double-whole":         42.0,
		"double-zero":          math.Copysign(0, +1), // the same as just 0.0 in Go
		"double-negative-zero": math.Copysign(0, -1),
		"double-max":           math.MaxFloat64,
		"double-smallest":      math.SmallestNonzeroFloat64,
		"double-nan":           math.NaN(),
		"double-big":           doubleBig,

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

		// TODO "unset": unset, https://github.com/FerretDB/FerretDB/issues/1023
	},
}

// Doubles contains double values for tests.
var Doubles = &Values[string]{
	name:     "Doubles",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "number"`),
		},
	},
	data: map[string]any{
		"double":          42.13,
		"double-whole":    42.0,
		"double-zero":     0.0,
		"double-max":      math.MaxFloat64,
		"double-smallest": math.SmallestNonzeroFloat64,
		"double-big":      doubleBig,
		"double-null":     nil,
	},
}

// Strings contains string values for tests.
var Strings = &Values[string]{
	name:     "Strings",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "string"`),
		},
	},
	data: map[string]any{
		"string":        "foo",
		"string-double": "42.13",
		"string-whole":  "42",
		"string-empty":  "",
		"string-null":   nil,
	},
}

// Binaries contains binary values for tests.
var Binaries = &Values[string]{
	name:     "Binaries",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "object", "properties": {"$b": {"type": "string", "format": "byte"}, "s": {"type": "integer", "format": "int32"}}`),
		},
	},
	data: map[string]any{
		"binary":       primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}},
		"binary-empty": primitive.Binary{Data: []byte{}},
		"binary-null":  nil,
	},
}

// ObjectIDs contains ObjectID values for tests.
var ObjectIDs = &Values[string]{
	name:     "ObjectIDs",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "string", "format": "byte"`),
		},
	},
	data: map[string]any{
		"objectid":       primitive.ObjectID{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11},
		"objectid-empty": primitive.NilObjectID,
		"objectid-null":  nil,
	},
}

// Bools contains bool values for tests.
var Bools = &Values[string]{
	name:     "Bools",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "boolean"`),
		},
	},
	data: map[string]any{
		"bool-false": false,
		"bool-true":  true,
		"bool-null":  nil,
	},
}

// DateTimes contains datetime values for tests.
var DateTimes = &Values[string]{
	name:     "DateTimes",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "string", "format": "date-time"`),
		},
	},
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
	name:     "Nulls",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": `{
				"title": "%%collection%%",
				"primary_key": ["_id"],
				"properties": {
					"_id": {"type": "string"}
				}
			}`,
		},
	},
	data: map[string]any{
		"null": nil,
	},
}

// Regexes contains regex values for tests.
var Regexes = &Values[string]{
	name:     "Regexes",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "object", "properties": {"$r": {"type": "string"}, "o": {"type": "string"}}`),
		},
	},
	data: map[string]any{
		"regex":       primitive.Regex{Pattern: "foo", Options: "i"},
		"regex-empty": primitive.Regex{},
		"regex-null":  nil,
	},
}

// Int32s contains int32 values for tests.
var Int32s = &Values[string]{
	name:     "Int32s",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "integer", "format": "int32"`),
		},
	},
	data: map[string]any{
		"int32":      int32(42),
		"int32-zero": int32(0),
		"int32-max":  int32(math.MaxInt32),
		"int32-min":  int32(math.MinInt32),
		"int32-null": nil,
	},
}

// Timestamps contains timestamp values for tests.
var Timestamps = &Values[string]{
	name:     "Timestamps",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "object", "properties": {"$t": {"type": "string"}}`),
		},
	},
	data: map[string]any{
		"timestamp":      primitive.Timestamp{T: 42, I: 13},
		"timestamp-i":    primitive.Timestamp{I: 1},
		"timestamp-null": nil,
	},
}

// Int64s contains int64 values for tests.
var Int64s = &Values[string]{
	name:     "Int64s",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "integer", "format": "int64"`),
		},
	},
	data: map[string]any{
		"int64":      int64(42),
		"int64-zero": int64(0),
		"int64-max":  int64(math.MaxInt64),
		"int64-min":  int64(math.MinInt64),
		"int64-big":  int64Big,
		"int64-null": nil,
	},
}

// Unsets contains unset value for tests.
var Unsets = &Values[string]{
	name:     "Unsets",
	handlers: []string{"pg", "tigris"},
	data: map[string]any{
		"unset": unset,
	},
}

// ObjectIDKeys contains documents with ObjectID keys for tests.
var ObjectIDKeys = &Values[primitive.ObjectID]{
	name:     "ObjectIDKeys",
	handlers: []string{"pg", "tigris"},
	data: map[primitive.ObjectID]any{
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11}: "objectid",
		primitive.NilObjectID: "objectid-empty",
	},
}

func tigrisSchema(typeString string) string {
	common := `{
				"title": "%%collection%%",
				"primary_key": ["_id"],
				"properties": {
					"v": {%%type%%},
					"_id": {"type": "string"}
				}
			}`
	return strings.ReplaceAll(common, "%%type%%", typeString)
}
