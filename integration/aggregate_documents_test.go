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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestAggregateProjectErrors(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		expectedErr mongo.CommandError
		altMessage  string
		skip        string
		pipeline    bson.A
	}{
		"EmptyPipeline": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{}}},
			},
			expectedErr: mongo.CommandError{
				Code:    51272,
				Name:    "Location51272",
				Message: "Invalid $project :: caused by :: projection specification must have at least one field",
			},
		},
		"EmptyProjectionField": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v", bson.D{}}}}},
			},
			expectedErr: mongo.CommandError{
				Code:    51270,
				Name:    "Location51270",
				Message: "Invalid $project :: caused by :: An empty sub-projection is not a valid value. Found empty object at path",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2633",
		},
		"PositionalOperatorMultiple": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v.$.foo.$", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 31324,
				Name: "Location31324",
				Message: "Invalid $project :: caused by :: " +
					"Cannot use positional projection in aggregation projection",
			},
		},
		"PositionalOperatorMiddle": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v.$.foo", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 31394,
				Name: "Location31394",
				Message: "Invalid $project :: caused by :: " +
					"As of 4.4, it's illegal to specify positional operator " +
					"in the middle of a path.Positional projection may only be " +
					"used at the end, for example: a.b.$. If the query previously " +
					"used a form like a.b.$.d, remove the parts following the '$' and " +
					"the results will be equivalent.",
			},
			altMessage: "Invalid $project :: caused by :: " +
				"Positional projection may only be used at the end, " +
				"for example: a.b.$. If the query previously used a form " +
				"like a.b.$.d, remove the parts following the '$' and " +
				"the results will be equivalent.",
		},
		"PositionalOperatorWrongLocations": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"$.v.$.foo", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 31394,
				Name: "Location31394",
				Message: "Invalid $project :: caused by :: " +
					"As of 4.4, it's illegal to specify positional operator " +
					"in the middle of a path.Positional projection may only be " +
					"used at the end, for example: a.b.$. If the query previously " +
					"used a form like a.b.$.d, remove the parts following the '$' and " +
					"the results will be equivalent.",
			},
			altMessage: "Invalid $project :: caused by :: " +
				"Positional projection may only be used at the end, " +
				"for example: a.b.$. If the query previously used a form " +
				"like a.b.$.d, remove the parts following the '$' and " +
				"the results will be equivalent.",
		},
		"EmptyKey": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 40352,
				Name: "Location40352",
				Message: "Invalid $project :: caused by :: " +
					"FieldPath cannot be constructed with empty string",
			},
		},
		"PositionalOperatorEmptyPath": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v..$", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 31324,
				Name: "Location31324",
				Message: "Invalid $project :: caused by :: " +
					"Cannot use positional projection in aggregation projection",
			},
		},
		"PositionalOperatorDollarKey": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"$", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 16410,
				Name: "Location16410",
				Message: "Invalid $project :: caused by :: " +
					"FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"PositionalOperatorDollarInKey": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"$v", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 16410,
				Name: "Location16410",
				Message: "Invalid $project :: caused by :: " +
					"FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"PositionalOperatorDollarPrefix": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"$.foo", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 16410,
				Name: "Location16410",
				Message: "Invalid $project :: caused by :: " +
					"FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"PositionalOperatorDotDollarInKey": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v.$foo", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 16410,
				Name: "Location16410",
				Message: "Invalid $project :: caused by :: " +
					"FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"PositionalOperatorPrefixSuffix": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"$.v.$", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 31324,
				Name: "Location31324",
				Message: "Invalid $project :: caused by :: " +
					"Cannot use positional projection in aggregation projection",
			},
		},
		"PositionalOperatorExclusion": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v.$", false}}}},
			},
			expectedErr: mongo.CommandError{
				Code: 31324,
				Name: "Location31324",
				Message: "Invalid $project :: caused by :: " +
					"Cannot use positional projection in aggregation projection",
			},
		},
		"ProjectPositionalOperator": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v.$", true}}}},
			},
			expectedErr: mongo.CommandError{
				Code:    31324,
				Name:    "Location31324",
				Message: "Invalid $project :: caused by :: Cannot use positional projection in aggregation projection",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()
			ctx, collection := setup.Setup(t)

			_, err := collection.Aggregate(ctx, tc.pipeline)

			AssertEqualAltCommandError(t, tc.expectedErr, tc.altMessage, err)
		})
	}
}
