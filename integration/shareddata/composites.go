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

	"go.mongodb.org/mongo-driver/bson"
)

// Composites contain composite values for tests.
//
// This shared data set is not frozen yet, but please add to it only if it is really shared.
var Composites = &Values[string]{
	name:     "Composites",
	handlers: []string{"pg"},
	data: map[string]any{
		"document":                   bson.D{{"foo", int32(42)}},
		"document-composite":         bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}},
		"document-composite-reverse": bson.D{{"array", bson.A{int32(42), "foo", nil}}, {"42", "foo"}, {"foo", int32(42)}},
		"document-null":              bson.D{{"foo", nil}},
		"document-empty":             bson.D{},

		"array":                 bson.A{int32(42)},
		"array-two":             bson.A{42.13, math.NaN()},
		"array-three":           bson.A{int32(42), "foo", nil},
		"array-three-reverse":   bson.A{nil, "foo", int32(42)},
		"array-embedded":        bson.A{bson.A{"42", "foo"}},
		"array-first-embedded":  bson.A{bson.A{int32(42), "foo"}, nil},
		"array-middle-embedded": bson.A{nil, bson.A{int32(42), "foo"}, nil},
		"array-last-embedded":   bson.A{nil, bson.A{int32(42), "foo"}},
		"array-empty":           bson.A{},
		"array-empty-nested":    bson.A{bson.A{}},
		"array-null":            bson.A{nil},
	},
}

// DocumentsDoubles contains documents with double values for tests.
var DocumentsDoubles = &Values[string]{
	name:     "DocumentsDoubles",
	handlers: []string{"pg", "tigris"},
	data: map[string]any{
		"document-double":          bson.D{{"v", 42.13}},
		"document-double-whole":    bson.D{{"v", 42.0}},
		"document-double-zero":     bson.D{{"v", 0.0}},
		"document-double-max":      bson.D{{"v", math.MaxFloat64}},
		"document-double-smallest": bson.D{{"v", math.SmallestNonzeroFloat64}},
		"document-double-big":      bson.D{{"v", doubleBig}},
		// TODO Dealing with empty doc needs a schema to be defined https://github.com/FerretDB/FerretDB/issues/772
		// "document-empty":           bson.D{},
	},
}

// DocumentsStrings contains documents with string values for tests.
var DocumentsStrings = &Values[string]{
	name:     "DocumentsStrings",
	handlers: []string{"pg", "tigris"},
	data: map[string]any{
		"document-string":        bson.D{{"v", "foo"}},
		"document-string-double": bson.D{{"v", "42.13"}},
		"document-string-whole":  bson.D{{"v", "42"}},
		"document-string-empty":  bson.D{{"v", ""}},
		// TODO Dealing with empty doc needs a schema to be defined https://github.com/FerretDB/FerretDB/issues/772
		// "document-empty":         bson.D{},
	},
}
