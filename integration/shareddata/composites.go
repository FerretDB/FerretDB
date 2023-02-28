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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Composites contain composite values for tests.
//
// This shared data set is not frozen yet, but please add to it only if it is really shared.
var Composites = &Values[string]{
	name:     "Composites",
	handlers: []string{"pg"},
	data: map[string]any{
		"document": bson.D{{"foo", int32(42)}},
		"document-composite": bson.D{
			{"foo", int32(42)},
			{"42", "foo"},
			{"array", bson.A{int32(42), "foo", nil}},
		},
		"document-composite-reverse": bson.D{
			{"array", bson.A{int32(42), "foo", nil}},
			{"42", "foo"},
			{"foo", int32(42)},
		},
		"document-null":  bson.D{{"foo", nil}},
		"document-empty": bson.D{},

		"array":               bson.A{int32(42)},
		"array-two":           bson.A{42.13, "foo"},
		"array-three":         bson.A{int32(42), "foo", nil},
		"array-three-reverse": bson.A{nil, "foo", int32(42)},
		"array-empty":         bson.A{},
		"array-null":          bson.A{nil},
		"array-numbers-asc":   bson.A{int32(42), int64(43), 45.5},
		"array-strings-desc":  bson.A{"c", "b", "a"},
		"array-documents":     bson.A{bson.D{{"field", int32(42)}}, bson.D{{"field", int32(44)}}},
		"array-composite": bson.A{
			42.13,
			"foo",
			primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}},
			primitive.ObjectID{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11},
			true,
			primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC)),
			nil,
			primitive.Regex{Pattern: "foo", Options: "i"},
			int32(42),
			primitive.Timestamp{T: 42, I: 13},
			int64(41),
		},
	},
}

// PostgresEdgeCases contains documents with keys and values that could be parsed in a wrong way
// on pg handler.
var PostgresEdgeCases = &Values[string]{
	name:     "PostgresEdgeCases",
	handlers: []string{"pg"},
	data: map[string]any{
		"document-notations": bson.D{
			{"foo[0]", int32(42)},
			{"*", int32(42)},
			{"foo[*]", int32(42)},
			{"@", int32(42)},
			{"f,oo", int32(42)},
		},
	},
}

// DocumentsDoubles contains documents with double values for tests.
var DocumentsDoubles = &Values[string]{
	name:     "DocumentsDoubles",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "object", "properties": {"v": {"type": "number"}}`),
		},
	},
	data: map[string]any{
		"document-double":          bson.D{{"v", 42.13}},
		"document-double-whole":    bson.D{{"v", 42.0}},
		"document-double-zero":     bson.D{{"v", 0.0}},
		"document-double-max":      bson.D{{"v", math.MaxFloat64}},
		"document-double-smallest": bson.D{{"v", math.SmallestNonzeroFloat64}},
		"document-double-big":      bson.D{{"v", doubleBig}},
		"document-double-empty":    bson.D{},
		"document-double-null":     nil,
	},
}

// DocumentsStrings contains documents with string values for tests.
var DocumentsStrings = &Values[string]{
	name:     "DocumentsStrings",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": tigrisSchema(`"type": "object", "properties": {"v": {"type": "string"}}`),
		},
	},
	data: map[string]any{
		"document-string":           bson.D{{"v", "foo"}},
		"document-string-double":    bson.D{{"v", "42.13"}},
		"document-string-whole":     bson.D{{"v", "42"}},
		"document-string-empty-str": bson.D{{"v", ""}},
		"document-string-empty":     bson.D{},
		"document-string-nil":       nil,
	},
}

// DocumentsDocuments contains documents with documents for tests.
var DocumentsDocuments = &Values[primitive.ObjectID]{
	name:     "DocumentsDocuments",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": `{
				"title": "%%collection%%",
				"primary_key": ["_id"],
				"properties": {
					"v": {
						"type": "object",
						"properties": {
							"foo": {"type": "integer", "format": "int32"},
							"bar": {"type": "object", "properties":{}}
						}
					},
					"_id": {"type": "string", "format": "byte"}
				}
			}`,
		},
	},
	data: map[primitive.ObjectID]any{
		{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}: bson.D{{"foo", int32(42)}},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}: bson.D{{"bar", bson.D{}}},
	},
}

// ArrayStrings contains an array with string values for tests.
// Tigris JSON schema validator contains extra properties to make it suitable for more tests.
var ArrayStrings = &Values[string]{
	name:     "ArrayStrings",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": `{
				"title": "%%collection%%",
				"primary_key": ["_id"],
				"properties": {
					"foo": {"type": "integer", "format": "int32"},
					"bar": {"type": "array", "items": {"type": "string"}},
					"v": {"type": "array", "items": {"type": "string"}},
					"_id": {"type": "string"}
				}
			}`,
		},
	},
	data: map[string]any{
		"array-string-desc":      bson.A{"c", "b", "a"},
		"array-string-duplicate": bson.A{nil, "foo", "b", "b", nil},
		"array-string-numbers":   bson.A{"42", "0", "42.13"},
		// "array-string-nil":    nil, TODO: https://github.com/FerretDB/FerretDB/issues/1836
		"array-string-empty":    bson.A{},
		"array-string-with-nil": bson.A{nil},
	},
}

// ArrayDoubles contains an array with float64 values for tests.
var ArrayDoubles = &Values[string]{
	name:     "ArrayDoubles",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": `{
				"title": "%%collection%%",
				"primary_key": ["_id"],
				"properties": {
					"v": {"type": "array", "items": {"type": "number"}},
					"_id": {"type": "string"}
				}
			}`,
		},
	},
	data: map[string]any{
		"array-double-desc":      bson.A{float64(40), float64(15), float64(10)},
		"array-double-duplicate": bson.A{float64(10), float64(10), float64(20)},
		// "array-double-nil":    nil,  TODO: https://github.com/FerretDB/FerretDB/issues/1836
		"array-double-empty": bson.A{},
	},
}

// ArrayInt32s contains an array with int32 values for tests.
var ArrayInt32s = &Values[string]{
	name:     "ArrayInt32s",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": `{
				"title": "%%collection%%",
				"primary_key": ["_id"],
				"properties": {
					"v": {"type": "array", "items": {"type": "integer", "format": "int32"}},
					"_id": {"type": "string"}
				}
			}`,
		},
	},
	data: map[string]any{
		"array-int32-one":   bson.A{int32(42)},
		"array-int32-two":   bson.A{int32(42), int32(42)},
		"array-int32-three": bson.A{int32(42), int32(43), int32(42)},
		// "array-int32-nil": nil,  TODO: https://github.com/FerretDB/FerretDB/issues/1836
		"array-int32-empty": bson.A{},
	},
}

// ArrayRegexes contains an array with regex values for tests.
var ArrayRegexes = &Values[string]{
	name:     "ArrayRegexes",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": `{
				"title": "%%collection%%",
				"primary_key": ["_id"],
				"properties": {
					"v": {"type": "array", "items":
						{
							"type": "object",
							"properties": {
								"$r": {"type": "string"},
								"o": {"type": "string"}
							}
						}
					},
					"_id": {"type": "string"}
				}
			}`,
		},
	},
	data: map[string]any{
		"array-regex": bson.A{primitive.Regex{Pattern: "foo", Options: "i"}, primitive.Regex{Pattern: "foo", Options: "i"}},
	},
}

// ArrayDocuments contains array with documents with arrays: {"v": [{"foo": [{"bar": "hello"}]}, ...]}.
// This data set is helpful for dot notation tests: v.0.foo.0.bar.
var ArrayDocuments = &Values[string]{
	name:     "ArrayDocuments",
	handlers: []string{"pg", "tigris"},
	validators: map[string]map[string]any{
		"tigris": {
			"$tigrisSchemaString": `{
				"title": "%%collection%%",
				"primary_key": ["_id"],
				"properties": {
					"v": {
						"type": "array", "items": {
							"type": "object",
							"properties": {
								"foo": {"type": "array", "items": {"type": "object", "properties": {"bar": {"type": "string"}}}}
							}
						}
					},
					"_id": {"type": "string"}
				}
			}`,
		},
	},
	data: map[string]any{
		"array-documents-nested": bson.A{
			bson.D{{
				"foo",
				bson.A{
					bson.D{{"bar", "hello"}},
					bson.D{{"bar", "world"}},
				},
			}},
		},
	},
}
