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

package integration

import (
	"math"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryProjectionCompat(t *testing.T) {
	t.Parallel()

	// topLevelFieldsIntegers contains documents with several top level fields with integer values.
	topLevelFieldsIntegers := shareddata.NewTopLevelFieldsProvider(
		"TopLevelFieldsIntegers",
		nil,
		map[string]map[string]any{
			"ferretdb-tigris": {
				"$tigrisSchemaString": `{
				"title": "%%collection%%",
				"primary_key": ["_id"],
				"properties": {
					"foo": {"type": "integer", "format": "int32"},
					"bar": {"type": "integer", "format": "int32"},
					"_id": {"type": "string"}
				}
			}`,
			},
		},
		map[string]shareddata.Fields{
			"int32-two": {
				{Key: "foo", Value: int32(1)},
				{Key: "bar", Value: int32(2)},
			},
		},
	)

	providers := append(shareddata.AllProviders(), topLevelFieldsIntegers)

	testCases := map[string]queryCompatTestCase{
		"EmptyProjection": {
			filter:     bson.D{},
			projection: bson.D{},
		},
		"NilProjection": {
			filter:     bson.D{},
			projection: nil,
		},
		"Include1Field": {
			filter:     bson.D{},
			projection: bson.D{{"v", int32(1)}},
		},
		"Exclude1Field": {
			filter:     bson.D{},
			projection: bson.D{{"v", int32(0)}},
		},
		"Include2Fields": {
			filter:     bson.D{},
			projection: bson.D{{"foo", 1.24}, {"bar", true}},
		},
		"Include2FieldsReverse": {
			filter:     bson.D{},
			projection: bson.D{{"bar", true}, {"foo", 1.24}},
		},
		"Exclude2Fields": {
			filter:     bson.D{},
			projection: bson.D{{"foo", int32(0)}, {"bar", false}},
		},
		"Include1FieldExclude1Field": {
			filter:     bson.D{},
			projection: bson.D{{"foo", int32(0)}, {"bar", true}},
			resultType: emptyResult,
		},
		"Exclude1FieldInclude1Field": {
			filter:     bson.D{},
			projection: bson.D{{"foo", int32(1)}, {"bar", false}},
			resultType: emptyResult,
		},
		"IncludeID": {
			filter:     bson.D{},
			projection: bson.D{{"_id", int64(-1)}},
		},
		"ExcludeID": {
			filter:      bson.D{},
			projection:  bson.D{{"_id", false}},
			skipIDCheck: true,
		},
		"IncludeFieldExcludeID": {
			filter:      bson.D{},
			projection:  bson.D{{"_id", false}, {"v", true}},
			skipIDCheck: true,
		},
		"ExcludeFieldIncludeID": {
			filter:     bson.D{},
			projection: bson.D{{"_id", true}, {"v", false}},
		},
		"ExcludeFieldExcludeID": {
			filter:      bson.D{},
			projection:  bson.D{{"_id", false}, {"v", false}},
			skipIDCheck: true,
		},
		"IncludeFieldIncludeID": {
			filter:     bson.D{},
			projection: bson.D{{"_id", true}, {"v", true}},
		},
		"Assign1Field": {
			filter:     bson.D{},
			projection: bson.D{{"foo", primitive.NewObjectID()}},
		},
		"AssignID": {
			filter:      bson.D{},
			projection:  bson.D{{"_id", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}},
			skipIDCheck: true,
		},
		"Assign1FieldIncludeID": {
			filter:     bson.D{},
			projection: bson.D{{"_id", true}, {"foo", primitive.NewDateTimeFromTime(time.Unix(0, 0))}},
		},
		"Assign2FieldsIncludeID": {
			filter:     bson.D{},
			projection: bson.D{{"_id", true}, {"foo", nil}, {"bar", "qux"}},
		},
		"Assign1FieldExcludeID": {
			filter:      bson.D{},
			projection:  bson.D{{"_id", false}, {"foo", primitive.Regex{Pattern: "^fo"}}},
			skipIDCheck: true,
		},
		"DotNotationInclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", true}},
		},
		"DotNotationIncludeTwo": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", true}, {"v.array", true}},
		},
		"DotNotationIncludeTwoReverse": {
			filter:     bson.D{},
			projection: bson.D{{"v.array", true}, {"v.foo", true}},
		},
		"DotNotationIncludeTwoArray": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", true}, {"v.bar", true}},
		},
		"DotNotationExclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", false}},
		},
		"DotNotationExcludeTwo": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", false}, {"v.array", false}},
		},
		"DotNotationExcludeSecondLevel": {
			filter:     bson.D{},
			projection: bson.D{{"v.array.42", false}},
		},
		"DotNotationIncludeSecondLevel": {
			filter:     bson.D{},
			projection: bson.D{{"v.array.42", true}},
		},
		"DotNotationIncludeExclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", true}, {"v.array", false}},
			resultType: emptyResult,
		},
		"DotNotation5LevelInclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.a.b.c.d", true}},
		},
		"DotNotation5LevelExclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.a.b.c.d", false}},
		},
		"DotNotation4LevelInclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.a.b.c", true}},
		},
		"DotNotation4LevelExclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.a.b.c", false}},
		},
		"DotNotationArrayInclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.array.0", true}},
		},
		"DotNotationArrayExclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.array.0", false}},
		},
		"DotNotationArrayPathInclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.0.foo", true}},
		},
		"DotNotationArrayPathExclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.0.foo", false}},
		},
		"DotNotationManyInclude": {
			filter: bson.D{},
			projection: bson.D{
				{"v.42", true},
				{"v.non-existent", true},
				{"v.foo", true},
				{"v.array", true},
			},
		},
		"DotNotationManyExclude": {
			filter: bson.D{},
			projection: bson.D{
				{"v.42", false},
				{"v.non-existent", false},
				{"v.foo", false},
				{"v.array", false},
			},
		},
	}

	testQueryCompatWithProviders(t, providers, testCases)
}

func TestQueryProjectionPositionalOperatorCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"IDFilter": {
			// it returns error only if collection contains a doc that matches the filter
			// and the filter does not contain positional operator path,
			// e.g. missing {v: <val>} in the filter.
			filter:         bson.D{{"_id", "array"}},
			projection:     bson.D{{"v.$", true}},
			resultPushdown: true,
		},
		"Implicit": {
			filter:         bson.D{{"v", float64(42)}},
			projection:     bson.D{{"v.$", true}},
			resultPushdown: true,
		},
		"ImplicitNoMatch": {
			filter:         bson.D{{"v", "non-existent"}},
			projection:     bson.D{{"v.$", true}},
			resultPushdown: true,
			resultType:     emptyResult,
		},
		"Eq": {
			filter:         bson.D{{"v", bson.D{{"$eq", 45.5}}}},
			projection:     bson.D{{"v.$", true}},
			resultPushdown: true,
		},
		"Gt": {
			filter:     bson.D{{"v", bson.D{{"$gt", 42}}}},
			projection: bson.D{{"v.$", true}},
		},
		"GtNoMatch": {
			filter:     bson.D{{"v", bson.D{{"$gt", math.MaxFloat64}}}},
			projection: bson.D{{"v.$", true}},
			resultType: emptyResult,
		},
		"DollarEndingKey": {
			filter:     bson.D{{"v", bson.D{{"$gt", 42}}}},
			projection: bson.D{{"v$", true}},
		},
		"DollarPartOfKey": {
			filter:     bson.D{{"v", bson.D{{"$gt", 42}}}},
			projection: bson.D{{"v$v", true}},
		},
		"ImplicitDotNotation": {
			filter:         bson.D{{"v", float64(42)}},
			projection:     bson.D{{"v.foo.$", true}},
			resultPushdown: true,
		},
		"ImplicitDotNoMatch": {
			filter:         bson.D{{"v", "non-existent"}},
			projection:     bson.D{{"v.foo.$", true}},
			resultPushdown: true,
			resultType:     emptyResult,
		},
		"GtDotNotation": {
			filter:     bson.D{{"v", bson.D{{"$gt", 42}}}},
			projection: bson.D{{"v.foo.$", true}},
		},
		"GtDotNoMatch": {
			filter:     bson.D{{"v", bson.D{{"$gt", 42}}}},
			projection: bson.D{{"v.foo.$", true}},
		},
		"DotNotationDollarEndingKey": {
			filter:     bson.D{{"v", bson.D{{"$gt", 42}}}},
			projection: bson.D{{"v.foo$", true}},
		},
		"IDValueFilters": {
			filter: bson.D{
				{"_id", "array"},
				{"v", bson.D{{"$gt", 41}}},
			},
			projection:     bson.D{{"v.$", true}},
			resultPushdown: true,
		},
		"TwoFilter": {
			filter: bson.D{
				{"v", bson.D{{"$lt", 43}}},
				{"v", bson.D{{"$gt", 41}}},
			},
			projection: bson.D{{"v.$", true}},
		},
		"TwoConflictingLtGt": {
			filter: bson.D{
				{"v", bson.D{{"$lt", 42}}},
				{"v", bson.D{{"$gt", 42}}},
			},
			projection: bson.D{{"v.$", true}},
			skip:       "https://github.com/FerretDB/FerretDB/issues/2522",
		},
		"TwoConflictingGtLt": {
			filter: bson.D{
				{"v", bson.D{{"$gt", 42}}},
				{"v", bson.D{{"$lt", 42}}},
			},
			projection: bson.D{{"v.$", true}},
			skip:       "https://github.com/FerretDB/FerretDB/issues/2522",
		},
		"PartialProjection": {
			filter: bson.D{
				{"v.foo", bson.D{{"$gt", 42}}},
			},
			projection: bson.D{{"v.$", true}},
			resultType: emptyResult,
		},
		"PartialFilter": {
			filter: bson.D{
				{"v", bson.D{{"$gt", 42}}},
			},
			projection: bson.D{{"v.foo.$", true}},
		},
		"TypeOperator": {
			filter:     bson.D{},
			projection: bson.D{{"type", bson.D{{"$type", "$v"}}}},
			skip:       "https://github.com/FerretDB/FerretDB/issues/2679",
		},
	}

	testQueryCompat(t, testCases)
}
