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

func TestAggregateVariablesCompatRoot(t *testing.T) {
	t.Parallel()

	providers := shareddata.AllProviders().Remove(shareddata.Composites)

	testCases := map[string]aggregateStagesCompatTestCase{
		"AddFields": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"field", "$$ROOT"}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1413",
		},
		"GroupID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", "$$ROOT"}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1992",
		},
		"GroupIDTwice": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", "$$ROOT"}}}},
				bson.D{{"$group", bson.D{{"_id", "$$ROOT"}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1992",
		},
		"GroupIDFieldID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", "$$ROOT._id"}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1992",
		},
		"GroupIDFieldValue": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$$ROOT.v"},
					// set first _id of the documents as group's unique value
					{"unique", bson.D{{"$first", "$_id"}}},
				}}},
				bson.D{{"$sort", bson.D{{"unique", 1}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2185",
		},
		"GroupIDFieldDotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", "$$ROOT.v.foo"}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1992",
		},
		"GroupIDFieldArrayDotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", "$$ROOT.v.0"}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1992",
		},
		"GroupIDExpression": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"v", "$$ROOT"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1992",
		},
		"GroupSumAccumulator": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$_id"},
					{"sum", bson.D{{"$sum", "$$ROOT"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
		},
		"ProjectSumOperator": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$$ROOT"}}},
				}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1992",
		},
		"ProjectTypeOperator": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"type", bson.D{{"$type", "$$ROOT"}}},
				}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1992",
		},
		"Set": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"field", "$$ROOT"}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1413",
		},
		"Unwind": {
			pipeline: bson.A{
				bson.D{{"$unwind", "$$ROOT"}},
			},
			resultType: emptyResult,
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}
