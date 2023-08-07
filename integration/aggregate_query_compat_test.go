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
)

func TestAggregateCompatMatchExpr(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"Expression": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", "$v"}}}},
			},
		},
		"ExpressionDotNotation": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", "$v.foo"}}}},
			},
		},
		"Document": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"v", "$v"}}}}}},
			},
		},
		"DocumentInvalid": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"v", "$"}}}}}},
			},
			resultType: emptyResult,
		},
		"Array": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.A{"$v"}}}}},
			},
		},
		"ArrayInvalid": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.A{"$"}}}}},
			},
			resultType: emptyResult,
		},
		"String": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", "v"}}}},
			},
		},
		"Int": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", int32(1)}}}},
			},
		},
		"NonExistent": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", "$non-existent"}}}},
			},
			resultType: emptyResult,
		},
		"Type": {
			pipeline: bson.A{bson.D{{"$match", bson.D{
				{"$expr", bson.D{{"$type", "$v"}}},
			}}}},
		},
		"Sum": {
			pipeline: bson.A{bson.D{{"$match", bson.D{
				{"$expr", bson.D{{"$sum", "$v"}}},
			}}}},
		},
		"SumType": {
			pipeline: bson.A{bson.D{{"$match", bson.D{
				{"$expr", bson.D{{"$type", bson.D{{"$sum", "$v"}}}}},
			}}}},
		},
		"Gt": {
			pipeline: bson.A{bson.D{{"$match", bson.D{
				{"$expr", bson.D{{"$gt", bson.A{"$v", 2}}}},
			}}}},
			skip: "https://github.com/FerretDB/FerretDB/issues/1456",
		},
	}

	testAggregateStagesCompat(t, testCases)
}
