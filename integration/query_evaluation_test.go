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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryEvaluationRegex(t *testing.T) {
	// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/1576

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "multiline-string"}, {"v", "bar\nfoo"}},
		bson.D{
			{"_id", "document-nested-strings"},
			{"v", bson.D{{"foo", bson.D{{"bar", "quz"}}}}},
		},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      any   // required
		expectedIDs []any // optional
	}{
		"Regex": {
			filter:      bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "foo"}}}}},
			expectedIDs: []any{"multiline-string", "string"},
		},
		"RegexNested": {
			filter:      bson.D{{"v.foo.bar", bson.D{{"$regex", primitive.Regex{Pattern: "quz"}}}}},
			expectedIDs: []any{"document-nested-strings"},
		},
		"RegexWithOption": {
			filter:      bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "42", Options: "i"}}}}},
			expectedIDs: []any{"string-double", "string-whole"},
		},
		"RegexStringOptionMatchCaseInsensitive": {
			filter:      bson.D{{"v", bson.D{{"$regex", "foo"}, {"$options", "i"}}}},
			expectedIDs: []any{"multiline-string", "regex", "string"},
		},
		"RegexStringOptionMatchLineEnd": {
			filter:      bson.D{{"v", bson.D{{"$regex", "b.*foo"}, {"$options", "s"}}}},
			expectedIDs: []any{"multiline-string"},
		},
		"RegexStringOptionMatchMultiline": {
			filter:      bson.D{{"v", bson.D{{"$regex", "^foo"}, {"$options", "m"}}}},
			expectedIDs: []any{"multiline-string", "string"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, tc.filter, "filter must not be nil")

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

func TestQueryEvaluationExprErrors(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Composites)

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		filter bson.D // required, aggregation pipeline stages

		err        *mongo.CommandError // required
		altMessage string              // optional, alternative error message
		skip       string              // optional, skip test with a specified reason
	}{
		"TooManyFields": {
			filter: bson.D{{"$expr", bson.D{{"$type", "v"}, {"$op", "v"}}}},
			err: &mongo.CommandError{
				Code:    15983,
				Name:    "Location15983",
				Message: `An object representing an expression must have exactly one field: { $type: "v", $op: "v" }`,
			},
			altMessage: "An object representing an expression must have exactly one field",
		},
		"TypeWrongLen": {
			filter: bson.D{{"$expr", bson.D{{"$type", bson.A{"foo", "bar"}}}}},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Expression $type takes exactly 1 arguments. 2 were passed in.",
			},
		},
		"InvalidExpression": {
			filter: bson.D{{"$expr", bson.D{{"$type", bson.D{{"$type", bson.A{"foo", "bar"}}}}}}},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Expression $type takes exactly 1 arguments. 2 were passed in.",
			},
		},
		"InvalidNestedExpression": {
			filter: bson.D{{"$expr", bson.D{{"$type", bson.D{{"$non-existent", "foo"}}}}}},
			err: &mongo.CommandError{
				Code:    168,
				Name:    "InvalidPipelineOperator",
				Message: "Unrecognized expression '$non-existent'",
			},
		},
		"EmptyPath": {
			filter: bson.D{{"$expr", "$"}},
			err: &mongo.CommandError{
				Code:    16872,
				Name:    "Location16872",
				Message: "'$' by itself is not a valid FieldPath",
			},
		},
		"EmptyVariable": {
			filter: bson.D{{"$expr", "$$"}},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "empty variable names are not allowed",
			},
		},
		"InvalidVariable$": {
			filter: bson.D{{"$expr", "$$$"}},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "'$' starts with an invalid character for a user variable name",
			},
		},
		"InvalidVariable$s": {
			filter: bson.D{{"$expr", "$$$s"}},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "'$s' starts with an invalid character for a user variable name",
			},
		},
		"Recursive": {
			filter: bson.D{{"$expr", bson.D{{"$expr", int32(1)}}}},
			err: &mongo.CommandError{
				Code:    168,
				Name:    "InvalidPipelineOperator",
				Message: "Unrecognized expression '$expr'",
			},
		},
		"GtNotArray": {
			filter: bson.D{{"$expr", bson.D{{"$gt", 1}}}},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Expression $gt takes exactly 2 arguments. 1 were passed in.",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1456",
		},
		"GtOneParameter": {
			filter: bson.D{{"$expr", bson.D{{"$gt", bson.A{1}}}}},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Expression $gt takes exactly 2 arguments. 1 were passed in.",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1456",
		},
		"GtThreeParameters": {
			filter: bson.D{{"$expr", bson.D{{"$gt", bson.A{1, 2, 3}}}}},
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

			require.NotNil(t, tc.filter, "filter must not be nil")

			_, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}
