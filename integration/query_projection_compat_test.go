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
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryProjectionCompat(t *testing.T) {
	t.Parallel()

	// topLevelFieldsIntegers contains documents with several top level fields with integer values.
	topLevelFieldsIntegers := shareddata.NewTopLevelFieldsProvider[string](
		"TopLevelFieldsIntegers",
		[]string{"ferretdb-pg", "ferretdb-tigris", "mongodb"},
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
				{"foo", int32(1)},
				{"bar", int32(2)},
			},
		},
	)

	providers := append(shareddata.AllProviders(), topLevelFieldsIntegers)

	testCases := map[string]queryCompatTestCase{
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
		"Exclude2Fields": {
			filter:     bson.D{},
			projection: bson.D{{"foo", int32(0)}, {"bar", false}},
		},
		"Include1FieldExclude1Field": {
			filter:     bson.D{},
			projection: bson.D{{"foo", int32(1)}, {"bar", true}},
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
		"DotNotationInclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", true}},
			skip:       "https://github.com/FerretDB/FerretDB/issues/2430",
		},
		"DotNotationIncludeTwo": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", true}, {"v.array", true}},
			skip:       "https://github.com/FerretDB/FerretDB/issues/2430",
		},
		"DotNotationExclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", false}},
			skip:       "https://github.com/FerretDB/FerretDB/issues/2430",
		},
		"DotNotationExcludeTwo": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", false}, {"v.array", false}},
			skip:       "https://github.com/FerretDB/FerretDB/issues/2430",
		},
		"DotNotationExcludeSecondLevel": {
			filter:     bson.D{},
			projection: bson.D{{"v.array.42", false}},
			skip:       "https://github.com/FerretDB/FerretDB/issues/2430",
		},
		"DotNotationIncludeExclude": {
			filter:     bson.D{},
			projection: bson.D{{"v.foo", true}, {"v.array", false}},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/2430",
		},
	}

	testQueryCompatWithProviders(t, providers, testCases)
}
