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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestAggregateMatchExprErrors(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		pipeline bson.A // required, aggregation pipeline stages

		err        *mongo.CommandError // required
		altMessage string              // optional, alternative error message
		skip       string              // optional, skip test with a specified reason
	}{
		"TooManyFields": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$type", "v"}, {"$op", "v"}}}}}},
			},
			err: &mongo.CommandError{
				Code:    15983,
				Name:    "Location15983",
				Message: `An object representing an expression must have exactly one field: { $type: "v", $op: "v" }`,
			},
			altMessage: "An object representing an expression must have exactly one field",
		},
		"TypeWrongLen": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$type", bson.A{"foo", "bar"}}}}}}},
			},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Expression $type takes exactly 1 arguments. 2 were passed in.",
			},
		},
		"InvalidExpression": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$type", bson.D{{"$type", bson.A{"foo", "bar"}}}}}}}}},
			},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Expression $type takes exactly 1 arguments. 2 were passed in.",
			},
		},
		"InvalidNestedExpression": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$type", bson.D{{"$non-existent", "foo"}}}}}}}},
			},
			err: &mongo.CommandError{
				Code:    168,
				Name:    "InvalidPipelineOperator",
				Message: "Unrecognized expression '$non-existent'",
			},
		},
		"EmptyPath": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", "$"}}}},
			},
			err: &mongo.CommandError{
				Code:    16872,
				Name:    "Location16872",
				Message: "'$' by itself is not a valid FieldPath",
			},
		},
		"EmptyVariable": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", "$$"}}}},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "empty variable names are not allowed",
			},
		},
		"InvalidVariable$": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", "$$$"}}}},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "'$' starts with an invalid character for a user variable name",
			},
		},
		"InvalidVariable$s": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", "$$$s"}}}},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "'$s' starts with an invalid character for a user variable name",
			},
		},
		"Recursive": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$expr", int32(1)}}}}}},
			},
			err: &mongo.CommandError{
				Code:    168,
				Name:    "InvalidPipelineOperator",
				Message: "Unrecognized expression '$expr'",
			},
		},
		"GtNotArray": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$gt", 1}}}}}},
			},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Expression $gt takes exactly 2 arguments. 1 were passed in.",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1456",
		},
		"GtOneParameter": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$gt", bson.A{1}}}}}}},
			},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Expression $gt takes exactly 2 arguments. 1 were passed in.",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1456",
		},
		"GtThreeParameters": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$gt", bson.A{1, 2, 3}}}}}}},
			},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Expression $gt takes exactly 2 arguments. 3 were passed in.",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1456",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.pipeline, "pipeline must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			ctx, collection := setup.Setup(t)

			_, err := collection.Aggregate(ctx, tc.pipeline)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}
